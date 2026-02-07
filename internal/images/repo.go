package images

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var ErrNoImages = errors.New("no images found")

type Repository interface {
	GetAll(ctx context.Context) ([]string, error)
	GetRandom(ctx context.Context) (string, error)
}

type FSRepository struct {
	root     string
	ttl      time.Duration
	mu       sync.RWMutex
	files    []string // relative to root
	lastScan time.Time
}

func NewFSRepository(root string, ttl time.Duration) *FSRepository {
	return &FSRepository{root: root, ttl: ttl}
}

func (r *FSRepository) shouldRescan() bool {
	return time.Since(r.lastScan) > r.ttl || r.lastScan.IsZero()
}

func (r *FSRepository) scanLocked() error {
	var images []string
	err := filepath.WalkDir(r.root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type().IsRegular() {
			ext := strings.ToLower(filepath.Ext(path))
			if IsSupported(ext) {
				rel, _ := filepath.Rel(r.root, path)
				images = append(images, rel)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	r.files = images
	r.lastScan = time.Now()
	return nil
}

func (r *FSRepository) ensureScannedLocked() error {
	if r.shouldRescan() {
		return r.scanLocked()
	}
	return nil
}

func (r *FSRepository) GetAll(ctx context.Context) ([]string, error) {
	r.mu.RLock()
	if !r.shouldRescan() {
		defer r.mu.RUnlock()
		out := make([]string, len(r.files))
		copy(out, r.files)
		return out, nil
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.ensureScannedLocked(); err != nil {
		return nil, err
	}
	out := make([]string, len(r.files))
	copy(out, r.files)
	return out, nil
}

func (r *FSRepository) GetRandom(ctx context.Context) (string, error) {
	r.mu.RLock()
	if !r.shouldRescan() && len(r.files) > 0 {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(r.files))))
		if err != nil {
			r.mu.RUnlock()
			return "", err
		}
		result := r.files[int(n.Int64())]
		r.mu.RUnlock()
		return result, nil
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.ensureScannedLocked(); err != nil {
		return "", err
	}
	if len(r.files) == 0 {
		return "", ErrNoImages
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(r.files))))
	if err != nil {
		return "", err
	}
	return r.files[int(n.Int64())], nil
}
