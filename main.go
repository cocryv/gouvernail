package main

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type ReverseProxy struct {
	demoUrl *url.URL
	cache   map[string]cacheItem
}

type cacheItem struct {
	data       []byte
	expiration time.Time
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

	if _, ok := (rp.cache[rp.demoUrl.Host]); ok && time.Now().Before(rp.cache[rp.demoUrl.Host].expiration) {
		log.Info().Msg("Cache hit")

		w.WriteHeader(http.StatusOK)
		// set Forwarded header
		w.Header().Set("X-Forwarded-For", r.RemoteAddr)

		w.Write(rp.cache[rp.demoUrl.Host].data)

	} else {
		resp, err := http.DefaultClient.Do(r)
		if err != nil {
			log.Error().Err(err).Msg("Failed to proxy request")
			w.WriteHeader(http.StatusBadGateway)
			return
		}

		if resp.StatusCode >= 400 {
			log.Error().
				Str("host", r.Host).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("remoteAddr", r.RemoteAddr).
				Int("statusCode", resp.StatusCode).
				Msg("Upstream server returned an error")

			w.WriteHeader(resp.StatusCode)
			return
		}

		log.Info().Msg("Cache miss")

		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Error().Err(err).Msg("Failed to read response body")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		rp.cache[rp.demoUrl.Host] = cacheItem{
			data:       respBody,
			expiration: time.Now().Add(1 * time.Minute),
		}

		w.Header().Set("X-Forwarded-For", r.RemoteAddr)
		w.WriteHeader(resp.StatusCode)
		w.Write(respBody)
	}

	// Log the outgoing response
	log.Info().
		Str("host", r.Host).
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Msg("Completed request")
}

func main() {
	log.Info().Msg("gouvernail starting up...")

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// demoUrl is a URL to a demo server that we will proxy requests to

	demoUrl, err := url.Parse("http://127.0.0.1:55000/")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse demo url")
	}

	cache := make(map[string]cacheItem)

	rp := &ReverseProxy{demoUrl, cache}

	http.ListenAndServe(":8080", rp)
}
