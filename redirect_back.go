package redirect_back

import (
	"net/http"
	"strings"

	"github.com/qor/session"
)

// Config redirect back config
type Config struct {
	SessionManager  session.ManagerInterface
	FallbackPath    string
	IgnoredPaths    []string
	IgnoredPrefixes []string
	IgnoreFunc      func(*http.Request) bool
}

// New initialize redirect back instance
func New(config *Config) *RedirectBack {
	redirectBack := &RedirectBack{config: config}
	redirectBack.compile()
	return redirectBack
}

// RedirectBack redirect back struct
type RedirectBack struct {
	config          *Config
	ignoredPathsMap map[string]bool
}

func (redirectBack *RedirectBack) compile() {
	redirectBack.ignoredPathsMap = map[string]bool{}

	for _, pth := range redirectBack.config.IgnoredPaths {
		redirectBack.ignoredPathsMap[pth] = true
	}
}

// IgnorePath check path is ignored or not
func (redirectBack *RedirectBack) IgnorePath(req *http.Request) bool {
	if redirectBack.ignoredPathsMap[req.URL.Path] {
		return true
	}

	for _, prefix := range redirectBack.config.IgnoredPrefixes {
		if strings.HasPrefix(req.URL.Path, prefix) {
			return true
		}
	}

	if redirectBack.config.IgnoreFunc != nil {
		return redirectBack.config.IgnoreFunc(req)
	}

	return false
}

// RedirectBack redirect back to last visited page
func (redirectBack *RedirectBack) RedirectBack(w http.ResponseWriter, req *http.Request) {
	if returnTo := redirectBack.config.SessionManager.Pop(req, "return_to"); returnTo != "" {
		http.Redirect(w, req, returnTo, http.StatusSeeOther)
	}

	if redirectBack.config.FallbackPath != "" {
		http.Redirect(w, req, redirectBack.config.FallbackPath, http.StatusSeeOther)
	}

	http.Redirect(w, req, "/", http.StatusSeeOther)
}

// Middleware returns a RedirectBack middleware instance that record return_to path
func (redirectBack *RedirectBack) Middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if !redirectBack.IgnorePath(req) {
			redirectBack.config.SessionManager.Add(req, "return_to", req.URL.String())
		}

		handler.ServeHTTP(w, req)
	})
}
