package web

import (
	"io/fs"
	"net/http"
)

func Serve(addr string) error {
	distFS, err := fs.Sub(distFiles, "dist")
	if err != nil {
		return err
	}
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.FS(distFS))))
	return http.ListenAndServe(addr, http.DefaultServeMux)
}
