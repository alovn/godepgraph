package app

import (
	"io/fs"
	"net/http"
)

func Serve(root string, addr string, showStdLib bool) error {
	distFS, err := fs.Sub(distFiles, "dist")
	if err != nil {
		return err
	}
	mux := http.DefaultServeMux
	mux.Handle("/", http.StripPrefix("/", http.FileServer(http.FS(distFS))))
	mux.HandleFunc("/graph", graphHandler(root, showStdLib))
	return http.ListenAndServe(addr, mux)
}

func graphHandler(root string, showStdLib bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		showStd := showStdLib
		query := r.URL.Query().Get("std")
		if query != "" {
			showStd = query == "true"
		}
		if err := OutputGraphFormat(w, root, showStd); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
		}
	}
}
