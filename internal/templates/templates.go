package templates

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

//go:embed html/*.html html/**/*.html
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

type Templates struct {
	pages    map[string]*template.Template
	partials *template.Template
}

func New() *Templates {
	funcMap := template.FuncMap{
		"add":      func(a, b int) int { return a + b },
		"sub":      func(a, b int) int { return a - b },
		"safeHTML": func(s string) template.HTML { return template.HTML(s) },
		"safeURL":  func(s string) template.URL { return template.URL(s) },
		"isStaff": func(role string) bool {
			return role == "admin" || role == "organizer"
		},
		"flag": func(code string) string {
			if len(code) != 2 {
				return ""
			}
			r0 := rune(code[0]) - 'A' + 0x1F1E6
			r1 := rune(code[1]) - 'A' + 0x1F1E6
			return string([]rune{r0, r1})
		},
		"deref": func(s *string) string {
			if s == nil {
				return ""
			}
			return *s
		},
		"fmtDate": func(s string) string {
			t, err := time.Parse(time.RFC3339, s)
			if err != nil {
				t, err = time.Parse("2006-01-02", s)
			}
			if err != nil {
				return s
			}
			return t.Format("Mon, 2 Jan 2006")
		},
		"isoDate": func(s string) string {
			t, err := time.Parse(time.RFC3339, s)
			if err != nil {
				t, err = time.Parse("2006-01-02", s)
			}
			if err != nil {
				return s
			}
			return t.Format("2006-01-02")
		},
	}

	// Parse base layout and shared partials (nav, etc.)
	base := template.Must(
		template.New("base").Funcs(funcMap).ParseFS(templateFS, "html/base.html", "html/nav.html"),
	)

	// Parse each page template by cloning the base
	pages := make(map[string]*template.Template)

	entries, _ := fs.ReadDir(templateFS, "html/pages")
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		clone := template.Must(base.Clone())
		pages[name] = template.Must(clone.ParseFS(templateFS, "html/pages/"+entry.Name()))
	}

	// Partials are available standalone (for HTMX responses)
	partials := template.Must(
		template.New("").Funcs(funcMap).ParseFS(templateFS, "html/*.html", "html/**/*.html"),
	)

	return &Templates{pages: pages, partials: partials}
}

// RenderPage renders a full page (base layout + page content).
// The page name corresponds to the filename without extension in html/pages/.
func (t *Templates) RenderPage(w http.ResponseWriter, page string, data any) error {
	tmpl, ok := t.pages[page]
	if !ok {
		return fmt.Errorf("template %q not found", page)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return tmpl.ExecuteTemplate(w, "base", data)
}

// Render is an alias for RenderPage for backward compat.
func (t *Templates) Render(w http.ResponseWriter, page string, data any) error {
	return t.RenderPage(w, page, data)
}

// RenderPartial renders a named partial template (for HTMX responses).
func (t *Templates) RenderPartial(w io.Writer, name string, data any) error {
	return t.partials.ExecuteTemplate(w, name, data)
}

func StaticFS() fs.FS {
	sub, _ := fs.Sub(staticFS, "static")
	return sub
}
