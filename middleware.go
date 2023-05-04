package napoleon

import "net/http"

func (n *Napoleon) SessionLoad(next http.Handler) http.Handler {
	return n.Session.LoadAndSave(next)
}
