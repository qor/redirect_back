# Redirect Back

A Golang HTTP Handler that redirect back to last URL saved in session

## Usage

```go
import (
	"github.com/qor/redirect_back"
	"github.com/qor/session/manager"
)

func main() {
	redirectBack := redirect_back.New(
		IgnoredPaths:    []string{"/login"}, // Will ignore requests that has those paths when set return path
		IgnoredPrefixes: []string{"/auth"},  // Will ignore requests that has those prefixes when set return path
		IgnoreFunc:      func(req *http.Request) bool { // Will ignore requests if `IgnoreFunc` returns true
			if user, ok := req.Context().Value("current_user").(*User); ok {
				return user.IsAdmin && strings.HasPrefix(req.URL.Path, "/admin")
			}
		},
		SessionManager:  manager.SessionManager,
		FallbackPath: "/",
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/redirectBack", redirectBackHandler)

	// Wrap your application's handlers or router with redirect back's middleware
	http.ListenAndServe(":7000", redirectBack.Middleware(mux))
}

func homeHandler(w http.ResponseWriter, req *http.Request) {
	w.Writer([]byte("home"))
}

func redirectBackHandler(w http.ResponseWriter, req *http.Request) {
	// redirect to return path or the default one
	redirectBack.RedirectBack(w, req)
}
```
