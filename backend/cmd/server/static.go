package main

import (
	"io/fs"
	"net/http"
	"path"
)

// staticFiles serves the built frontend from dir. Unlike a plain
// http.FileServer, it answers 404 instead of listing a directory's contents.
func staticFiles(dir string) http.Handler {
	return http.FileServer(noListing{http.Dir(dir)})
}

// noListing hides directories that have no index.html. http.FileServer would
// otherwise list their contents.
type noListing struct {
	fs http.FileSystem
}

func (n noListing) Open(name string) (http.File, error) {
	f, err := n.fs.Open(name)
	if err != nil {
		return nil, err
	}
	if info, err := f.Stat(); err == nil && info.IsDir() {
		index, err := n.fs.Open(path.Join(name, "index.html"))
		if err != nil {
			f.Close()
			return nil, fs.ErrNotExist
		}
		index.Close()
	}
	return f, nil
}
