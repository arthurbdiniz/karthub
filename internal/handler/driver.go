package handler

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/karthub/karthub/internal/middleware"
	"github.com/karthub/karthub/internal/models"
	"github.com/karthub/karthub/internal/repository"
	"github.com/karthub/karthub/internal/templates"
	xdraw "golang.org/x/image/draw"
)

type Driver struct {
	drivers *repository.DriverRepository
	results *repository.ResultRepository
	tmpl    *templates.Templates
}

func NewDriver(drivers *repository.DriverRepository, results *repository.ResultRepository, tmpl *templates.Templates) *Driver {
	return &Driver{drivers: drivers, results: results, tmpl: tmpl}
}

func (h *Driver) List(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	drivers, _ := h.drivers.List(r.Context())
	myDriver, _ := h.drivers.GetByUserID(r.Context(), user.ID)
	if err := h.tmpl.Render(w, "drivers", map[string]any{
		"User":     user,
		"Drivers":  drivers,
		"MyDriver": myDriver,
	}); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

func (h *Driver) Show(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	user := middleware.UserFromContext(r.Context())
	driver, err := h.drivers.GetByID(r.Context(), id)
	if err != nil || driver == nil {
		http.NotFound(w, r)
		return
	}

	isMe := driver.UserID != nil && *driver.UserID == user.ID
	history, _ := h.results.HistoryByDriver(r.Context(), id)
	stats, _ := h.results.StatsByDriver(r.Context(), id)

	if err := h.tmpl.Render(w, "driver-detail", map[string]any{
		"User":    user,
		"Driver":  driver,
		"IsMe":    isMe,
		"History": history,
		"Stats":   stats,
	}); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

// SetupForm shows the initial profile creation after first login.
func (h *Driver) SetupForm(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	existing, _ := h.drivers.GetByUserID(r.Context(), user.ID)
	if existing != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	if err := h.tmpl.Render(w, "setup", map[string]any{"User": user}); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

func (h *Driver) Setup(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	existing, _ := h.drivers.GetByUserID(r.Context(), user.ID)
	if existing != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	nickname := r.FormValue("nickname")
	if nickname == "" {
		if err := h.tmpl.Render(w, "setup", map[string]any{"User": user, "Error": "Nickname is required"}); err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
		}
		return
	}

	// Check nickname uniqueness
	taken, err := h.drivers.NicknameExists(r.Context(), nickname, 0)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	if taken {
		if err := h.tmpl.Render(w, "setup", map[string]any{"User": user, "Error": "That nickname is already taken. Please choose another."}); err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
		}
		return
	}

	d := &models.Driver{
		UserID:   &user.ID,
		Name:     r.FormValue("name"),
		Nickname: &nickname,
		JoinedAt: time.Now(),
	}

	if d.Name == "" {
		d.Name = nickname
	}

	// Handle avatar upload
	if avatar := readAvatar(r); avatar != "" {
		d.Avatar = &avatar
	}
	if v := r.FormValue("country_code"); v != "" {
		d.CountryCode = &v
	}

	if err := h.drivers.Create(r.Context(), d); err != nil {
		if err := h.tmpl.Render(w, "setup", map[string]any{"User": user, "Error": "Failed to create profile"}); err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
		}
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func (h *Driver) NewForm(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	existing, _ := h.drivers.GetByUserID(r.Context(), user.ID)
	if existing != nil {
		http.Redirect(w, r, "/drivers/me/edit", http.StatusFound)
		return
	}
	if err := h.tmpl.Render(w, "driver-form", map[string]any{"User": user}); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

func (h *Driver) CreateProfile(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())

	d := &models.Driver{
		UserID:   &user.ID,
		Name:     r.FormValue("name"),
		JoinedAt: time.Now(),
	}
	if v := r.FormValue("nickname"); v != "" {
		d.Nickname = &v
	}
	if v := r.FormValue("bio"); v != "" {
		d.Bio = &v
	}
	if avatar := readAvatar(r); avatar != "" {
		d.Avatar = &avatar
	}
	if v := r.FormValue("country_code"); v != "" {
		d.CountryCode = &v
	}

	if err := h.drivers.Create(r.Context(), d); err != nil {
		http.Error(w, "Failed to create profile", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/drivers", http.StatusFound)
}

func (h *Driver) EditMyProfile(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	driver, _ := h.drivers.GetByUserID(r.Context(), user.ID)
	if driver == nil {
		http.Redirect(w, r, "/drivers/new", http.StatusFound)
		return
	}
	if err := h.tmpl.Render(w, "driver-form", map[string]any{"User": user, "Driver": driver}); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

func (h *Driver) UpdateMyProfile(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	driver, _ := h.drivers.GetByUserID(r.Context(), user.ID)
	if driver == nil {
		http.Redirect(w, r, "/drivers/new", http.StatusFound)
		return
	}
	driver.Name = r.FormValue("name")
	if v := r.FormValue("nickname"); v != "" {
		// Check nickname uniqueness
		taken, err := h.drivers.NicknameExists(r.Context(), v, driver.ID)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		if taken {
			if err := h.tmpl.Render(w, "driver-form", map[string]any{
				"User":   user,
				"Driver": driver,
				"Error":  "That nickname is already taken. Please choose another.",
			}); err != nil {
				http.Error(w, "Internal error", http.StatusInternalServerError)
			}
			return
		}
		driver.Nickname = &v
	}
	if v := r.FormValue("bio"); v != "" {
		driver.Bio = &v
	}
	if avatar := readAvatar(r); avatar != "" {
		driver.Avatar = &avatar
	}
	if r.FormValue("remove_avatar") == "1" {
		driver.Avatar = nil
	}
	if v := r.FormValue("country_code"); v != "" {
		driver.CountryCode = &v
	} else {
		driver.CountryCode = nil
	}
	if err := h.drivers.Update(r.Context(), driver); err != nil {
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/drivers", http.StatusFound)
}

// Admin operations
func (h *Driver) CreateDriver(w http.ResponseWriter, r *http.Request) {
	d := &models.Driver{
		Name:     r.FormValue("name"),
		JoinedAt: time.Now(),
	}
	if v := r.FormValue("nickname"); v != "" {
		d.Nickname = &v
	}
	if err := h.drivers.Create(r.Context(), d); err != nil {
		http.Error(w, "Failed to create driver", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/drivers", http.StatusFound)
}

func (h *Driver) EditForm(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	user := middleware.UserFromContext(r.Context())
	driver, _ := h.drivers.GetByID(r.Context(), id)
	if err := h.tmpl.Render(w, "driver-form", map[string]any{"User": user, "Driver": driver}); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

func (h *Driver) Update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	driver, _ := h.drivers.GetByID(r.Context(), id)
	if driver == nil {
		http.NotFound(w, r)
		return
	}
	driver.Name = r.FormValue("name")
	if v := r.FormValue("bio"); v != "" {
		driver.Bio = &v
	}
	if avatar := readAvatar(r); avatar != "" {
		driver.Avatar = &avatar
	}
	if err := h.drivers.Update(r.Context(), driver); err != nil {
		http.Error(w, "Failed to update driver", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/drivers", http.StatusFound)
}

func (h *Driver) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err := h.drivers.Delete(r.Context(), id); err != nil {
		http.Error(w, "Failed to delete driver", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/drivers", http.StatusFound)
}

// readAvatar reads the avatar from the form. It first checks for a pre-cropped
// base64 data URI from the client-side cropper, then falls back to the file upload.
// Returns empty string if no avatar provided.
func readAvatar(r *http.Request) string {
	if err := r.ParseMultipartForm(2 << 20); err != nil {
		return ""
	}

	// Prefer pre-cropped data from the client-side cropper
	if cropped := r.FormValue("avatar_cropped"); cropped != "" && len(cropped) > 20 {
		return cropped
	}

	file, _, err := r.FormFile("avatar")
	if err != nil {
		return ""
	}
	defer func() { _ = file.Close() }()

	data, err := io.ReadAll(file)
	if err != nil {
		return ""
	}

	ct := http.DetectContentType(data)

	// Decode image
	var img image.Image
	switch ct {
	case "image/jpeg":
		img, err = jpeg.Decode(bytes.NewReader(data))
	case "image/png":
		img, err = png.Decode(bytes.NewReader(data))
	default:
		// Unsupported format, store as-is but capped at raw size
		return "data:" + ct + ";base64," + base64.StdEncoding.EncodeToString(data)
	}
	if err != nil {
		return ""
	}

	// Resize to fit within 200x200 maintaining aspect ratio
	const maxSize = 200
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w > maxSize || h > maxSize {
		if w > h {
			h = h * maxSize / w
			w = maxSize
		} else {
			w = w * maxSize / h
			h = maxSize
		}
	}

	resized := image.NewRGBA(image.Rect(0, 0, w, h))
	xdraw.BiLinear.Scale(resized, resized.Bounds(), img, bounds, xdraw.Over, nil)

	// Encode as JPEG with quality 80
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, resized, &jpeg.Options{Quality: 80}); err != nil {
		return ""
	}

	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}
