package napoleon

import (
	"net/http"
	"strconv"

	"github.com/justinas/nosurf"
)

func (n *Napoleon) SessionLoad(next http.Handler) http.Handler {
	return n.Session.LoadAndSave(next)
}

func (n *Napoleon) NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	secure, _ := strconv.ParseBool(n.config.cookie.srcure)

	csrfHandler.ExemptGlob("/api/*")

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		Domain:   n.config.cookie.domain,
	})

	return csrfHandler
}
