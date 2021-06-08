package services

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const (
	webHostFlag  = "host"
	webPortFlag  = "port"
	webSourceURL = "source-url"
)

type Web struct {
	host      string
	port      int
	ln        net.Listener
	cl        *http.Client
	wp        *WriterPool
	sp        *StatPool
	sourceURL string
}

func NewWeb(c *cli.Context, wp *WriterPool, sp *StatPool) *Web {
	return &Web{
		host:      c.String(webHostFlag),
		port:      c.Int(webPortFlag),
		sourceURL: c.String(webSourceURL),
		wp:        wp,
		sp:        sp,
	}
}

func RegisterWebFlags(f []cli.Flag) []cli.Flag {
	return append(f,
		cli.StringFlag{
			Name:   webHostFlag,
			Usage:  "listening host",
			Value:  "",
			EnvVar: "WEB_HOST",
		},
		cli.IntFlag{
			Name:   webPortFlag,
			Usage:  "http listening port",
			Value:  8080,
			EnvVar: "WEB_PORT",
		},
		cli.StringFlag{
			Name:   webSourceURL,
			Usage:  "source url",
			Value:  "",
			EnvVar: "SOURCE_URL",
		},
	)
}

func (s *Web) getSourceURL(r *http.Request) string {
	if s.sourceURL != "" {
		return s.sourceURL
	}
	return r.Header.Get("X-Source-Url")
}

func (s *Web) Serve() error {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	ln, err := net.Listen("tcp", addr)
	s.ln = ln
	if err != nil {
		return errors.Wrap(err, "Failed to web listen to tcp connection")
	}
	m := http.NewServeMux()
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		su := s.getSourceURL(r)

		u, err := url.Parse(su)

		if err != nil {
			log.WithError(err).Errorf("Failed to parse url=%v", su)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		u.RawQuery = ""
		u.Path = ""

		id := r.URL.Query().Get("download-id")
		if id == "" {
			log.Errorf("Failed to find download-id url=%v", r.URL.String())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		t := &http.Transport{
			MaxIdleConns:        500,
			MaxIdleConnsPerHost: 500,
			MaxConnsPerHost:     500,
			IdleConnTimeout:     90 * time.Second,
		}
		proxy := httputil.NewSingleHostReverseProxy(u)
		proxy.Transport = t
		wi := s.wp.Get(id, w)
		proxy.ServeHTTP(wi, r)
	})
	log.Infof("Serving Web at %v", addr)
	return http.Serve(s.ln, m)
}

func (s *Web) Close() {
	log.Info("Closing Web")
	defer func() {
		log.Info("Web closed")
	}()
	if s.ln != nil {
		s.ln.Close()
	}
}
