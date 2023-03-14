package main

import (
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type ReverseProxy struct {
	demoUrl *url.URL
}

func (rp *ReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.Host = rp.demoUrl.Host
	r.URL.Host = rp.demoUrl.Host
	r.URL.Scheme = rp.demoUrl.Scheme
	r.RequestURI = ""

	log.Info().
		Str("host", r.Host).
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Str("remoteAddr", r.RemoteAddr).
		Msg("Received request")

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		log.Error().Err(err).Msg("Failed to proxy request")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)

	// Log the outgoing response
	log.Info().
		Str("host", r.Host).
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Int("statusCode", resp.StatusCode).
		Msg("Completed request")
}

func main() {
	log.Info().Msg("gouvernail starting up...")

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	demoUrl, err := url.Parse("http://127.0.0.1:55000/")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse demo url")
	}

	rp := &ReverseProxy{demoUrl}

	http.ListenAndServe(":8080", rp)
}
