package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultWaitTime      = 0
	defaultHeadersLength = 400
	defaultBodyLength    = 400
)

var (
	log = slog.New(slog.NewJSONHandler(os.Stdout, nil))
)

var sampleHeaders = map[string]string{
	"ETag":                   "489bbe95-4221-49a8-90c9-0050ffe752b5",
	"X-Config-Id":            "87428fc522803d31065e7bce3cf03fe475096631e5e07bbd7a0fde60c4cf25c7",
	"X-Device-Id":            "123456",
	"X-Server-Pool":          "my-pool.server.pauloavelar.com",
	"X-Random-Seed":          "AKQUW9912X",
	"Cache-Control":          "no-cache",
	"X-Custom-Header":        "custom-value",
	"X-Request-ID":           "ABCDEFGHIJKLMNOPQRSTUV",
	"X-Forwarded-For":        "192.168.1.1",
	"X-Forwarded-Host":       "pauloavelar.com",
	"X-Random-Date":          "Mon, 02 Jan 2006 15:04:05 MST",
	"Server":                 "Apache/2.4.41 (Unix)",
	"Set-Cookie":             "sessionid=123456789; Path=/; Secure; HttpOnly",
	"X-Content-Type-Options": "nosniff",
	"X-Frame-Options":        "DENY",
	"X-Powered-By":           "PHP/7.4.9",
	"X-XSS-Protection":       "1; mode=block",
	"X-Custom-Header-1":      "custom_value=1",
	"X-Custom-Header-2":      "custom_value=2",
	"X-Custom-Header-3":      "custom_value=3",
	"X-Custom-Header-4":      "custom_value=4",
	"X-Custom-Header-5":      "custom_value=5",
	"X-Custom-Header-6":      "custom_value=6",
	"X-Custom-Header-7":      "custom_value=7",
	"X-Custom-Header-8":      "custom_value=8",
	"X-Custom-Header-9":      "custom_value=9",
	"X-Custom-Header-0":      "custom_value=0",
}

func main() {
	log.Info("starting server")

	server := buildServer()

	err := server.ListenAndServe()
	if err != nil {
		log.Error("server failed", slog.Any("cause", err))
	}
}

func buildServer() *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/request", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		waitTime, ok := parseIntQuery(q, "time", defaultWaitTime)
		if !ok || waitTime < 0 || waitTime > 5000 {
			respondBadRequest(w)
			return
		}

		headersLen, ok := parseIntQuery(q, "headers", defaultHeadersLength)
		if !ok || headersLen < 0 || headersLen > 1000 {
			respondBadRequest(w)
			return
		}

		bodyLen, ok := parseIntQuery(q, "body", defaultBodyLength)
		if !ok || bodyLen < 0 || bodyLen > 2000 {
			respondBadRequest(w)
			return
		}

		log.Info(
			"Request received",
			slog.Int("wait_time", waitTime),
			slog.Int("headers_length", headersLen),
			slog.Int("body_length", bodyLen),
		)

		time.Sleep(time.Duration(waitTime) * time.Millisecond)

		var status int
		if bodyLen == 0 {
			status = http.StatusNoContent
		} else {
			status = http.StatusOK
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			headersLen -= 45
		}

		for headersLen > 0 {
			for name, value := range sampleHeaders {
				w.Header().Set(name, value)
				headersLen -= len(name)
				headersLen -= len(value)
			}
		}

		w.WriteHeader(status)

		switch {
		case bodyLen == 0:
			return
		case bodyLen == 1:
			writeBody(w, "1")
		case bodyLen == 2:
			writeBody(w, "12")
		case bodyLen >= 3 && bodyLen < 8:
			body := fmt.Sprintf(`"%q"`, strings.Repeat("3", bodyLen-2))
			writeBody(w, body)
		default:
			body := fmt.Sprintf(`{"a":"%q"}`, strings.Repeat("B", bodyLen-8))
			writeBody(w, body)
		}
	})

	return &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func parseIntQuery(query url.Values, name string, defaultValue int) (int, bool) {
	if !query.Has(name) {
		return defaultValue, true
	}

	v := query.Get(name)
	value, err := strconv.Atoi(v)
	if err != nil {
		log.Error("invalid query value", slog.String("param", name), slog.String("value", v), slog.Any("error", err))
		return 0, false
	}

	return value, true
}

func respondBadRequest(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	writeBody(w, "invalid request params")
}

func writeBody(w http.ResponseWriter, data string) {
	_, err := w.Write([]byte(data))
	if err != nil {
		log.Error("error writing body", slog.Any("cause", err))
	}
}
