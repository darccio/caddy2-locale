package locale_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	locale "github.com/imdario/caddy2-locale"
)

var (
	langs = []string{"en", "ca", "es", "hu", "ru", "fr", "pt", "eo", "oc"}
)

func TestInitLocales(t *testing.T) {
	m := new(locale.Middleware)
	m.Init(langs)

	locales := m.Locales()
	for i, tag := range locales {
		if tag.String() != langs[i] {
			t.Errorf("expected %q, got %q", langs[i], tag.String())
		}
	}
}

type stub struct {
}

func (s *stub) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	w.WriteHeader(http.StatusOK)
	return nil
}

func TestRedirect(t *testing.T) {
	m := new(locale.Middleware)
	m.Init(langs)

	testCases := []struct {
		name     string
		path     string
		accept   string
		expected string
		init     func(*http.Request)
	}{
		{
			name:     "front page request",
			path:     "/",
			accept:   "ca-ES,ca;q=0.9",
			expected: "/ca",
		},
		{
			name:   "front page request with matching default language",
			path:   "/",
			accept: "en,ca-ES,ca;q=0.9",
		},
		{
			name: "Accept-Language missing",
			path: "/",
		},
		{
			name: "Detected-Language cookie present",
			path: "/",
			init: func(r *http.Request) {
				r.Header.Set("Set-Cookie", "Detected-Language=hu")
			},
		},
		{
			name:   "non-front page request",
			path:   "/privacy",
			accept: "ca-ES,ca;q=0.9",
		},
	}

	for _, tC := range testCases {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, tC.path, nil)

		if tC.accept != "" {
			r.Header.Set("Accept-Language", tC.accept)
		}

		if tC.init != nil {
			tC.init(r)
		}

		t.Run(tC.name, func(t *testing.T) {
			if err := m.ServeHTTP(w, r, &stub{}); err != nil {
				t.Error(err)
			}

			if w.Header().Get("Location") != tC.expected {
				t.Errorf("expected %q, got %q", tC.expected, w.Header().Get("Location"))
			}
		})
	}
}
