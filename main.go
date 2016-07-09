package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"strings"

	"github.com/justinas/alice"
	"github.com/urfave/cli"
)

type config struct {
	Fallback bool
	Index    string
	Dir      http.Dir
	Port     int
	Script   bool
	Key      string
}

func main() {
	app := cli.NewApp()
	app.Name = "serve-http"
	app.Version = "1.0.0"
	app.Author = "Thanatat Tamtan"
	app.Email = "acoshift@gmail.com"
	app.Usage = "a very small http service for serve static contents"
	app.UsageText = "serve-http [--fallback] [--index=index.html] [--dir=www] [--port=80] [--script] [--key=APIKEY]"
	app.Flags = []cli.Flag{
		cli.BoolFlag{Name: "fallback", Usage: "enable to serve index file if current file not found"},
		cli.StringFlag{Name: "index", Value: "index.html"},
		cli.StringFlag{Name: "dir", Value: "www"},
		cli.IntFlag{Name: "port", Value: 80},
		cli.BoolFlag{Name: "script", Usage: "enable to run .sh as script"},
		cli.StringFlag{Name: "key", Usage: "authenticate key ?key="},
	}
	app.Action = func(c *cli.Context) error {
		run(config{
			Fallback: c.Bool("fallback"),
			Index:    c.String("index"),
			Dir:      http.Dir(c.String("dir")),
			Port:     c.Int("port"),
			Script:   c.Bool("script"),
			Key:      c.String("key"),
		})
		return nil
	}

	app.Run(os.Args)
}

func run(cfg config) {
	m := alice.New(recovery)

	// key auth middleware
	if len(cfg.Key) != 0 {
		m = m.Append(authKey(cfg))
	}

	// serve static
	m = m.Append(serveStatic(cfg))

	// serve fallback
	if cfg.Fallback {
		m = m.Append(serveFallback(cfg))
	}

	http.ListenAndServe(":"+strconv.Itoa(cfg.Port), m.Then(http.NotFoundHandler()))
}

type static struct {
	cfg config
}

func tryFile(file string, dir http.Dir) (http.File, os.FileInfo, error) {
	f, err := dir.Open(file)
	if err != nil {
		return nil, nil, err
	}

	fi, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, nil, err
	}
	return f, fi, nil
}

func serveStatic(cfg config) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" && r.Method != "HEAD" {
				h.ServeHTTP(w, r)
				return
			}
			file := r.URL.Path

			if cfg.Script && path.Ext(file) == ".sh" {
				out, err := exec.Command(path.Join(string(cfg.Dir), file)).Output()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.Write(out)
				return
			}

			addIndex := false

			f, fi, err := tryFile(file, cfg.Dir)
			if err != nil {
				// try .html
				addIndex = true
				file += ".html"
				f, fi, err = tryFile(file, cfg.Dir)
				if err != nil {
					h.ServeHTTP(w, r)
					return
				}
			}
			defer f.Close()

			// try to serve index file
			if !addIndex && fi.IsDir() {
				// redirect if missing trailing slash
				if !strings.HasSuffix(r.URL.Path, "/") {
					http.Redirect(w, r, r.URL.Path+"/", http.StatusFound)
					return
				}

				file = path.Join(file, cfg.Index)
				f, err = cfg.Dir.Open(file)
				if err != nil {
					h.ServeHTTP(w, r)
					return
				}
				defer f.Close()

				fi, err = f.Stat()
				if err != nil || fi.IsDir() {
					h.ServeHTTP(w, r)
					return
				}
			}

			http.ServeContent(w, r, file, fi.ModTime(), f)
		})
	}
}

func serveFallback(cfg config) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" && r.Method != "HEAD" {
				h.ServeHTTP(w, r)
				return
			}

			f, fi, err := tryFile(cfg.Index, cfg.Dir)
			if err != nil {
				h.ServeHTTP(w, r)
				return
			}
			defer f.Close()

			http.ServeContent(w, r, cfg.Index, fi.ModTime(), f)
		})
	}
}

func authKey(cfg config) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			k := r.URL.Query().Get("key")
			if len(k) != 0 && k == cfg.Key {
				h.ServeHTTP(w, r)
				return
			}
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		})
	}
}

func recovery(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				if w.Header().Get("Content-Type") == "" {
					w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				}

				w.WriteHeader(http.StatusInternalServerError)
				stack := make([]byte, 8192)
				stack = stack[:runtime.Stack(stack, false)]

				f := "PANIC: %s\n%s"

				fmt.Fprintf(w, f, err, stack)
			}
		}()

		h.ServeHTTP(w, r)
	})
}
