package main

import (
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/urfave/negroni"
)

const (
	dir       = http.Dir("public")
	indexFile = "index.html"
)

func main() {
	n := negroni.New(negroni.NewRecovery(), negroni.NewLogger())
	n.Use(&static{})
	n.UseHandler(http.NotFoundHandler())

	n.Run(":8080")
}

type static struct {
}

func tryFile(file string) (http.File, os.FileInfo, error) {
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

	f, fi, err := tryFile(file)
	if err != nil {
		// try .html
		addIndex = true
		file += ".html"
		f, fi, err = tryFile(file)
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

		file = path.Join(file, indexFile)
		f, err = dir.Open(file)
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
