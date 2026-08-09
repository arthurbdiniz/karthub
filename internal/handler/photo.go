package handler

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/karthub/karthub/internal/middleware"
	"github.com/karthub/karthub/internal/models"
	"github.com/karthub/karthub/internal/repository"
)

type Photo struct {
	photos    *repository.PhotoRepository
	drivers   *repository.DriverRepository
	uploadDir string
}

func NewPhoto(photos *repository.PhotoRepository, drivers *repository.DriverRepository, uploadDir string) *Photo {
	if err := os.MkdirAll(uploadDir, 0o750); err != nil {
		slog.Error("creating upload directory", "error", err)
	}
	return &Photo{photos: photos, drivers: drivers, uploadDir: uploadDir}
}

func (h *Photo) Upload(w http.ResponseWriter, r *http.Request) {
	eventID, _ := strconv.ParseInt(chi.URLParam(r, "eventID"), 10, 64)
	user := middleware.UserFromContext(r.Context())

	driver, _ := h.drivers.GetByUserID(r.Context(), user.ID)
	if driver == nil {
		http.Error(w, "No driver profile", http.StatusForbidden)
		return
	}

	participated, _ := h.photos.DriverParticipated(r.Context(), eventID, driver.ID)
	if !participated {
		http.Error(w, "Only participants can upload photos", http.StatusForbidden)
		return
	}

	if err := r.ParseMultipartForm(200 << 20); err != nil {
		http.Error(w, "Upload too large", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["photos"]
	if len(files) == 0 {
		http.Error(w, "No files uploaded", http.StatusBadRequest)
		return
	}

	for _, header := range files {
		file, err := header.Open()
		if err != nil {
			continue
		}

		ext := filepath.Ext(header.Filename)
		name, _ := randomFilename()
		filename := name + ext

		dst, err := os.Create(filepath.Join(h.uploadDir, filename))
		if err != nil {
			_ = file.Close()
			continue
		}

		if _, err := io.Copy(dst, file); err != nil {
			_ = dst.Close()
			_ = file.Close()
			continue
		}
		_ = dst.Close()
		_ = file.Close()

		photo := &models.EventPhoto{
			EventID:      eventID,
			DriverID:     driver.ID,
			Filename:     filename,
			OriginalName: header.Filename,
		}
		if err := h.photos.Create(r.Context(), photo); err != nil {
			slog.Error("saving photo record", "error", err)
		}
	}

	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (h *Photo) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	user := middleware.UserFromContext(r.Context())
	_ = user

	filename, err := h.photos.Delete(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to delete", http.StatusInternalServerError)
		return
	}

	_ = os.Remove(filepath.Join(h.uploadDir, filename))

	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (h *Photo) ServeFile(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "filename")
	http.ServeFile(w, r, filepath.Join(h.uploadDir, filename))
}

func randomFilename() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
