package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"net/http"
	"os"

	"github.com/karthub/karthub/internal/config"
	"github.com/karthub/karthub/internal/database"
	"github.com/karthub/karthub/internal/repository"
	xdraw "golang.org/x/image/draw"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	db, err := database.Open(cfg.Database.Path)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer db.Close()

	ctx := context.Background()
	driverRepo := repository.NewDriverRepository(db)

	drivers, err := driverRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("listing drivers: %w", err)
	}

	optimized := 0
	for i := range drivers {
		d := &drivers[i]
		if d.Avatar == nil || *d.Avatar == "" {
			continue
		}

		result := optimizeAvatar(*d.Avatar)
		if result == "" {
			fmt.Printf("  [skip] %s (already small or unsupported)\n", d.Name)
			continue
		}

		before := len(*d.Avatar)
		after := len(result)
		d.Avatar = &result
		if err := driverRepo.Update(ctx, d); err != nil {
			fmt.Printf("  [error] %s: %v\n", d.Name, err)
			continue
		}
		fmt.Printf("  [ok] %s: %dKB → %dKB\n", d.Name, before/1024, after/1024)
		optimized++
	}

	fmt.Printf("\nDone. Optimized %d/%d avatars.\n", optimized, len(drivers))
	return nil
}

func optimizeAvatar(dataURI string) string {
	const prefix = "base64,"
	idx := bytes.Index([]byte(dataURI), []byte(prefix))
	if idx < 0 {
		return ""
	}
	raw, err := base64.StdEncoding.DecodeString(dataURI[idx+len(prefix):])
	if err != nil {
		return ""
	}

	ct := http.DetectContentType(raw)
	var img image.Image
	switch ct {
	case "image/jpeg":
		img, err = jpeg.Decode(bytes.NewReader(raw))
	case "image/png":
		img, err = png.Decode(bytes.NewReader(raw))
	default:
		return ""
	}
	if err != nil {
		return ""
	}

	const maxSize = 200
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	if w <= maxSize && h <= maxSize {
		return ""
	}

	if w > h {
		h = h * maxSize / w
		w = maxSize
	} else {
		w = w * maxSize / h
		h = maxSize
	}

	resized := image.NewRGBA(image.Rect(0, 0, w, h))
	xdraw.BiLinear.Scale(resized, resized.Bounds(), img, bounds, xdraw.Over, nil)

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, resized, &jpeg.Options{Quality: 80}); err != nil {
		return ""
	}

	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}
