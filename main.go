package main

import (
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/urfave/cli"
	"github.com/urfave/negroni"
)

type config struct {
	Fallback bool
	Index    string
	Dir      http.Dir
}

func main() {
	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.BoolFlag{Name: "fallback"},
		cli.StringFlag{Name: "index", Value: "index.html"},
		cli.StringFlag{Name: "dir", Value: "www"},
	}
	app.Action = func(c *cli.Context) error {
		run(config{
			Fallback: c.Bool("fallback"),
			Index:    c.String("index"),
			Dir:      http.Dir(c.String("dir")),
		})
		return nil
	}

	app.Run(os.Args)
}

func run(cfg config) {
	n := negroni.New(negroni.NewRecovery())
	n.Use(&static{cfg})
	if cfg.Fallback {
		n.Use(&fallback{cfg})
	}
	n.UseHandler(http.NotFoundHandler())

	n.Run(":80")
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

func (s *static) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if r.Method != "GET" && r.Method != "HEAD" {
		next(rw, r)
		return
	}
	file := r.URL.Path
	addIndex := false

	f, fi, err := tryFile(file, s.cfg.Dir)
	if err != nil {
		// try .html
		addIndex = true
		file += ".html"
		f, fi, err = tryFile(file, s.cfg.Dir)
		if err != nil {
			next(rw, r)
			return
		}
	}
	defer f.Close()

	// try to serve index file
	if !addIndex && fi.IsDir() {
		// redirect if missing trailing slash
		if !strings.HasSuffix(r.URL.Path, "/") {
			http.Redirect(rw, r, r.URL.Path+"/", http.StatusFound)
			return
		}

		file = path.Join(file, s.cfg.Index)
		f, err = s.cfg.Dir.Open(file)
		if err != nil {
			next(rw, r)
			return
		}
		defer f.Close()

		fi, err = f.Stat()
		if err != nil || fi.IsDir() {
			next(rw, r)
			return
		}
	}

	http.ServeContent(rw, r, file, fi.ModTime(), f)
}

type fallback struct {
	cfg config
}

func (s *fallback) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if r.Method != "GET" && r.Method != "HEAD" {
		next(rw, r)
		return
	}

	f, fi, err := tryFile(s.cfg.Index, s.cfg.Dir)
	if err != nil {
		next(rw, r)
		return
	}
	defer f.Close()

	http.ServeContent(rw, r, s.cfg.Index, fi.ModTime(), f)
}
