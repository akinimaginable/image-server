package server

import (
	"encoding/json"
	"errors"
	"image-server/internal/images"
	"net/http"
	"os"
	"path/filepath"
)

type Server struct {
	repo     images.Repository
	imageDir string
	mux      *http.ServeMux
}

func New(repo images.Repository, imageDir string) *Server {
	s := &Server{
		repo:     repo,
		imageDir: imageDir,
		mux:      http.NewServeMux(),
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
	s.mux.HandleFunc("/", s.randomImageHandler)
	s.mux.HandleFunc("/list", s.listImagesHandler)
	s.mux.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir(s.imageDir))))
}

func (s *Server) Handler() http.Handler { return s.mux }

func (s *Server) randomImageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	img, err := s.repo.GetRandom(ctx)
	if err != nil {
		code := http.StatusInternalServerError
		if errors.Is(err, images.ErrNoImages) {
			code = http.StatusNotFound
		}
		http.Error(w, err.Error(), code)
		return
	}
	full := filepath.Join(s.imageDir, img)
	if _, err := os.Stat(full); os.IsNotExist(err) {
		http.Error(w, "Image not found, try again", http.StatusNotFound)
		return
	}
	// Disable caching for random image
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	http.ServeFile(w, r, full)
}

func (s *Server) listImagesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	files, err := s.repo.GetAll(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp := map[string]any{
		"total":  len(files),
		"images": files,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
