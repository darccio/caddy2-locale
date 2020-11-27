package locale

import (
	"fmt"
	"net/http"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"golang.org/x/text/language"
)

func init() {
	caddy.RegisterModule(Middleware{})
	httpcaddyfile.RegisterHandlerDirective("locale", parseCaddyfile)
}

func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var m Middleware
	err := m.UnmarshalCaddyfile(h.Dispenser)
	return m, err
}

type Middleware struct {
	locales []language.Tag
	matcher language.Matcher
}

// CaddyModule returns the Caddy module information.
func (Middleware) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID: "http.handlers.accept_language",
		New: func() caddy.Module {
			return new(Middleware)
		},
	}
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (m Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	if _, err := r.Cookie("Detected-Language"); err == nil {
		return next.ServeHTTP(w, r)
	}

	accept := r.Header.Get("Accept-Language")
	tag, _ := language.MatchStrings(m.matcher, accept)

	cookie := &http.Cookie{
		Name:    "Detected-Language",
		Value:   tag.String(),
		Expires: time.Now().Add(24 * time.Hour),
	}
	http.SetCookie(w, cookie)

	if r.URL.Path != "/" || tag == m.locales[0] {
		return next.ServeHTTP(w, r)
	}

	base, _ := tag.Base()
	redirect := fmt.Sprintf("/%s", base.String())
	http.Redirect(w, r, redirect, http.StatusTemporaryRedirect)

	return nil
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler.
func (m *Middleware) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		args := d.RemainingArgs()
		m.Init(args)

		break
	}

	return nil
}

func (m *Middleware) Init(langs []string) {
	m.locales = m.initLocales(langs)
	m.matcher = m.initMatcher(langs)
}

func (m *Middleware) initMatcher(langs []string) language.Matcher {
	return language.NewMatcher(m.locales)
}

func (m *Middleware) initLocales(langs []string) []language.Tag {
	locales := make([]language.Tag, len(langs))

	for i, v := range langs {
		tag, err := language.Parse(v)
		if err != nil {
			continue
		}

		locales[i] = tag
	}

	return locales
}

func (m *Middleware) Locales() []language.Tag {
	return m.locales
}

var (
	_ caddyhttp.MiddlewareHandler = (*Middleware)(nil)
	_ caddyfile.Unmarshaler       = (*Middleware)(nil)
)
