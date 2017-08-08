package redirect_back

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/qor/qor/utils"
	"github.com/qor/session"
	"github.com/qor/session/manager"
)

var returnToKey utils.ContextKey = "redirect_back_return_to"

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
	if config.SessionManager == nil {
		config.SessionManager = manager.SessionManager
	}

	if config.FallbackPath == "" {
		config.FallbackPath = "/"
	}

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
	returnTo := req.Context().Value(returnToKey)

	if returnTo != nil {
		http.Redirect(w, req, fmt.Sprint(returnTo), http.StatusSeeOther)
		return
	}

	http.Redirect(w, req, redirectBack.config.FallbackPath, http.StatusSeeOther)
}

// Middleware returns a RedirectBack middleware instance that record return_to path
func (redirectBack *RedirectBack) Middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if !redirectBack.IgnorePath(req) {
			returnTo := redirectBack.config.SessionManager.Get(req, "return_to")
			req = req.WithContext(context.WithValue(req.Context(), returnToKey, returnTo))

			if returnTo != req.URL.String() {
				redirectBack.config.SessionManager.Add(req, "return_to", req.URL.String())
			}
		}

		handler.ServeHTTP(w, req)
	})
}
