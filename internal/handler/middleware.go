package handler

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"io"
	"metrify/internal/service"
	"net/http"
	"strings"
	"time"
)

func (handler *Handler) WithLogging(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		loggingWriter := service.NewLoggingResponseWriter(w)
		start := time.Now()
		uri := r.RequestURI
		method := r.Method

		h.ServeHTTP(loggingWriter, r)
		end := time.Now()
		duration := end.Sub(start)

		handler.logger.Infoln(
			"uri", uri,
			"method", method,
			"duration", duration,
			"size", loggingWriter.ResponseData.Size,
			"status", loggingWriter.ResponseData.Status,
		)

	}

	return http.HandlerFunc(logFn)
}

func (handler *Handler) WithRequestCompress(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
			return
		}

		dr, err := service.NewCompressReader(r.Body)

		if err != nil {
			http.Error(w, "invalid gzip", http.StatusBadRequest)
			return
		}
		r.Body = dr
		h.ServeHTTP(w, r)
	})
}

func (handler *Handler) WithResponseCompress(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
			return
		}

		cw := service.NewCompressWriter(w)
		w = cw
		defer cw.Close()

		h.ServeHTTP(w, r)
	})
}

func (handler *Handler) WithHashedRequest(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body []byte
		hash := r.Header.Get("HashSHA256")

		if hash == "" {
			h.ServeHTTP(w, r)
			return
		}

		body, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, "failed to read body", http.StatusBadRequest)
			return
		}

		if service.SignData(body, handler.Key) != hash {
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		r.Body = io.NopCloser(bytes.NewReader(body))

		h.ServeHTTP(w, r)
	})
}

func (handler *Handler) WithDecrypt(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		encryption := r.Header.Get("Content-Encryption")
		if encryption != "RSA-PKCS1v15" {
			next.ServeHTTP(w, r)
			return
		}

		body, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, "Can't read request", http.StatusInternalServerError)
			return
		}

		decoded, err := base64.StdEncoding.DecodeString(string(body))
		if err != nil {
			http.Error(w, "Can't base64 decode request", http.StatusBadRequest)
			return
		}

		decrypted, err := rsa.DecryptPKCS1v15(rand.Reader, handler.privKey, decoded)
		if err != nil {
			http.Error(w, "Can't decrypt request", http.StatusBadRequest)
			return
		}

		r.Body = io.NopCloser(bytes.NewReader(decrypted))
		r.ContentLength = int64(len(decrypted))
		r.Header.Del("Content-Encryption")

		next.ServeHTTP(w, r)
	})
}

func (handler *Handler) WithTrustedSubnet(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if handler.TrustedSubnet == "" {
			h.ServeHTTP(w, r)
		}

		if handler.TrustedSubnet != r.Header.Get("X-Forwarded-For") {
			http.Error(w, "Not authorized", http.StatusForbidden)
			return
		}
	})
}
