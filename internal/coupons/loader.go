package coupons

import (
	"bufio"
	"compress/gzip"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

var promoCodeRe = regexp.MustCompile(`[A-Za-z0-9]{8,10}`)

// LoadValidCoupons downloads and parses each gzip URL, then returns all promo codes
// that can be found in at least two files.
//
// Discount percent is handled elsewhere (service layer); this loader only validates codes.
func LoadValidCoupons(ctx context.Context, urls []string) (*InMemoryStore, error) {
	return LoadValidCouponsWithCache(ctx, urls, "")
}

// LoadValidCouponsWithCache does the same work as LoadValidCoupons, but optionally
// persists the resulting hashed code set to disk for faster subsequent boots.
func LoadValidCouponsWithCache(ctx context.Context, urls []string, cachePath string) (*InMemoryStore, error) {
	if len(urls) == 0 {
		return nil, fmt.Errorf("no coupon base URLs provided")
	}

	if cachePath != "" {
		if store, ok, err := loadStoreFromCache(cachePath); err != nil {
			return nil, fmt.Errorf("load coupon cache: %w", err)
		} else if ok {
			log.Printf("coupons: loaded %d valid promo hashes from cache: %s", len(store.valid), cachePath)
			return store, nil
		}
	}

	startAll := time.Now()
	log.Printf("coupons: building promo cache from %d source files", len(urls))

	// code hash -> number of distinct files where it appears
	foundInFiles := make(map[uint64]int)

	for i, url := range urls {
		fileStart := time.Now()
		log.Printf("coupons: [%d/%d] downloading+parsing %s", i+1, len(urls), url)
		localCodes, err := loadUniqueCodesFromGzipURL(ctx, http.DefaultClient, url)
		if err != nil {
			return nil, fmt.Errorf("coupon base %d: %s: %w", i+1, url, err)
		}
		log.Printf("coupons: [%d/%d] parsed %d unique candidates in %s", i+1, len(urls), len(localCodes), time.Since(fileStart).Round(time.Millisecond))
		for h := range localCodes {
			foundInFiles[h]++
		}
	}

	valid := make(map[uint64]struct{})
	for h, n := range foundInFiles {
		if n >= 2 {
			valid[h] = struct{}{}
		}
	}

	store := NewInMemoryStore(valid)
	if cachePath != "" {
		if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
			return nil, fmt.Errorf("create cache dir: %w", err)
		}
		if err := saveStoreToCache(cachePath, store); err != nil {
			return nil, fmt.Errorf("save coupon cache: %w", err)
		}
		log.Printf("coupons: saved cache with %d valid promo hashes to %s", len(store.valid), cachePath)
	}

	log.Printf("coupons: build complete in %s (valid promo hashes: %d)", time.Since(startAll).Round(time.Millisecond), len(store.valid))
	return store, nil
}

func loadUniqueCodesFromGzipURL(ctx context.Context, client *http.Client, url string) (map[uint64]struct{}, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	pr := newProgressReader(resp.Body, url, resp.ContentLength)
	defer pr.LogFinal()

	return loadUniqueCodesFromGzipStream(pr)
}

func loadUniqueCodesFromGzipStream(gzipStream io.Reader) (map[uint64]struct{}, error) {
	zr, err := gzip.NewReader(gzipStream)
	if err != nil {
		return nil, fmt.Errorf("gzip reader: %w", err)
	}
	defer func() { _ = zr.Close() }()

	// Local set for “appears in this file”.
	local := make(map[uint64]struct{})

	// We stream-read and apply a regex over a sliding tail, so codes can be detected even when
	// they span arbitrary chunk boundaries.
	const tailKeep = 32
	const bufSize = 64 * 1024

	r := bufio.NewReader(zr)
	buf := make([]byte, bufSize)
	tail := ""

	for {
		n, err := r.Read(buf)
		if n > 0 {
			chunk := tail + string(buf[:n])
			matches := promoCodeRe.FindAllString(chunk, -1)
			for _, m := range matches {
				if h, ok := hashPromoCodeUpper(m); ok {
					local[h] = struct{}{}
				}
			}

			if len(chunk) > tailKeep {
				tail = chunk[len(chunk)-tailKeep:]
			} else {
				tail = chunk
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read stream: %w", err)
		}
	}

	return local, nil
}

func loadStoreFromCache(cachePath string) (*InMemoryStore, bool, error) {
	f, err := os.Open(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	defer func() { _ = f.Close() }()

	// Cache format: magic(8 bytes) + version(u32) + count(u64) + count*uint64
	var magic [8]byte
	if _, err := io.ReadFull(f, magic[:]); err != nil {
		return nil, false, err
	}
	if string(magic[:]) != "CPR1SET\x00"[:8] {
		return nil, false, fmt.Errorf("unknown cache magic")
	}

	var version uint32
	if err := binary.Read(f, binary.LittleEndian, &version); err != nil {
		return nil, false, err
	}
	if version != 1 {
		return nil, false, fmt.Errorf("unsupported cache version %d", version)
	}

	var count uint64
	if err := binary.Read(f, binary.LittleEndian, &count); err != nil {
		return nil, false, err
	}

	valid := make(map[uint64]struct{}, count)
	for i := uint64(0); i < count; i++ {
		var h uint64
		if err := binary.Read(f, binary.LittleEndian, &h); err != nil {
			return nil, false, err
		}
		valid[h] = struct{}{}
	}

	return NewInMemoryStore(valid), true, nil
}

func saveStoreToCache(cachePath string, store *InMemoryStore) error {
	tmpPath := cachePath + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	// Ensure cleanup on early return.
	defer func() {
		_ = f.Close()
		_ = os.Remove(tmpPath)
	}()

	magic := [8]byte{'C', 'P', 'R', '1', 'S', 'E', 'T', 0}
	if _, err := f.Write(magic[:]); err != nil {
		return err
	}

	var version uint32 = 1
	if err := binary.Write(f, binary.LittleEndian, &version); err != nil {
		return err
	}

	count := uint64(len(store.valid))
	if err := binary.Write(f, binary.LittleEndian, &count); err != nil {
		return err
	}

	for h := range store.valid {
		if err := binary.Write(f, binary.LittleEndian, &h); err != nil {
			return err
		}
	}

	// Flush to disk before rename.
	if err := f.Sync(); err != nil {
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, cachePath)
}

type progressReader struct {
	r            io.Reader
	label        string
	total        int64
	read         int64
	nextLogBytes int64
}

func newProgressReader(r io.Reader, label string, total int64) *progressReader {
	const logEvery = 25 * 1024 * 1024 // 25MB
	return &progressReader{
		r:            r,
		label:        label,
		total:        total,
		nextLogBytes: logEvery,
	}
}

func (p *progressReader) Read(buf []byte) (int, error) {
	n, err := p.r.Read(buf)
	if n > 0 {
		p.read += int64(n)
		for p.read >= p.nextLogBytes {
			p.logProgress()
			p.nextLogBytes += 25 * 1024 * 1024
		}
	}
	return n, err
}

func (p *progressReader) logProgress() {
	if p.total > 0 {
		percent := (float64(p.read) / float64(p.total)) * 100
		log.Printf("coupons: downloading %s %.1f%% (%d/%d MB)", p.label, percent, p.read/(1024*1024), p.total/(1024*1024))
		return
	}
	log.Printf("coupons: downloading %s (%d MB read)", p.label, p.read/(1024*1024))
}

func (p *progressReader) LogFinal() {
	if p.total > 0 {
		log.Printf("coupons: finished download %s (%d/%d MB)", p.label, p.read/(1024*1024), p.total/(1024*1024))
		return
	}
	log.Printf("coupons: finished download %s (%d MB read)", p.label, p.read/(1024*1024))
}

