package mss

import "net/http"

func MeasureHTTP(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		m := Measure("http", Data{"url": req.RequestURI, "content-type": req.Header.Get("Content-Type")})
		defer Done(m)

		sw := statusResponseWriter{w, 0}
		h.ServeHTTP(&sw, req)

		m.Data["status"] = sw.Status
	}

	return http.HandlerFunc(fn)
}

type statusResponseWriter struct {
	http.ResponseWriter
	Status int
}

func (w *statusResponseWriter) WriteHeader(status int) {
	w.ResponseWriter.WriteHeader(status)
	w.Status = status
}
