package ingress

import "net/http"

type RedirectingMiddleware struct{}

var redirectingMiddleware = RedirectingMiddleware{}

func (m *RedirectingMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.TLS == nil {
			redirectToHttps(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func redirectToHttps(w http.ResponseWriter, r *http.Request) {
	targetUrl := "https://" + r.Host + r.URL.RequestURI()
	http.Redirect(w, r, targetUrl, http.StatusPermanentRedirect) // #nosec G710: HTTPS redirect intentionally preserves the incoming same-host URL
}
