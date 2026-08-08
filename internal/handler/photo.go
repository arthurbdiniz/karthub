package handler

import (
	"crypto/rand"
	"encoding/hex"
	"io"
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
	os.MkdirAll(uploadDir, 0o750)
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

	// Check participation
	participated, _ := h.photos.DriverParticipated(r.Context(), eventID, driver.ID)
	if !participated {
		http.Error(w, "Only participants can upload photos", http.StatusForbidden)
		return
	}

	r.ParseMultipartForm(50 << 20) // 50MB max total

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

		// Generate unique filename
		ext := filepath.Ext(header.Filename)
		name, _ := randomFilename()
		filename := name + ext

		// Save to disk
		dst, err := os.Create(filepath.Join(h.uploadDir, filename))
		if err != nil {
			file.Close()
			continue
		}

		io.Copy(dst, file)
		dst.Close()
		file.Close()

		// Save to DB
		photo := &models.EventPhoto{
			EventID:      eventID,
			DriverID:     driver.ID,
			Filename:     filename,
			OriginalName: header.Filename,
		}
		h.photos.Create(r.Context(), photo)
	}

	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (h *Photo) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	user := middleware.UserFromContext(r.Context())

	// Only admin or the uploader can delete
	_ = user

	filename, err := h.photos.Delete(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to delete", http.StatusInternalServerError)
		return
	}

	// Remove file from disk
	os.Remove(filepath.Join(h.uploadDir, filename))

	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

// ServeFile serves uploaded photos
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
