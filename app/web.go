package app

import (
	"io/fs"
	"net/http"
)

func Serve(root string, addr, showPkgName string, showStdLib, showThirdLib bool) error {
	distFS, err := fs.Sub(distFiles, "dist")
	if err != nil {
		return err
	}
	mux := http.DefaultServeMux
	mux.Handle("/", http.StripPrefix("/", http.FileServer(http.FS(distFS))))
	mux.HandleFunc("/graph", graphHandler(root, showPkgName, showStdLib, showThirdLib))
	return http.ListenAndServe(addr, mux)
}

func graphHandler(root, showPkgName string, showStdLib, showThirdLib bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		showStd := showStdLib
		showPkg := showPkgName
		showThird := showThirdLib
		if query := r.URL.Query().Get("std"); query != "" {
			showStd = query == "true"
		}
		if query := r.URL.Query().Get("third"); query != "" {
			showThird = query == "true"
		}
		if pkg := r.URL.Query().Get("pkg"); pkg != "" {
			showPkg = pkg
		}
		w.Header().Add("x-pkg", showPkg)
		if err := OutputGraphFormat(w, root, showPkg, showStd, showThird); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
		}
	}
}
