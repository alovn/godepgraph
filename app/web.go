package app

import (
	"io/fs"
	"net/http"
	"strconv"
)

func Serve(root, addr, showPkgName string, showStdLib, showThirdLib, reverse bool) error {
	distFS, err := fs.Sub(distFiles, "dist")
	if err != nil {
		return err
	}
	mux := http.DefaultServeMux
	mux.Handle("/", http.StripPrefix("/", http.FileServer(http.FS(distFS))))
	mux.HandleFunc("/graph", graphHandler(root, showPkgName, showStdLib, showThirdLib, reverse))
	return http.ListenAndServe(addr, mux)
}

func graphHandler(root, showPkgName string, showStdLib, showThirdLib, reverse bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		showStd := showStdLib
		pkgName := showPkgName
		showThird := showThirdLib
		isReverse := reverse && pkgName != ""
		isInit := r.URL.Query().Get("init") == "true"
		if query := r.URL.Query().Get("std"); query != "" {
			showStd = query == "true"
		}
		if query := r.URL.Query().Get("third"); query != "" {
			showThird = query == "true"
		}
		if query := r.URL.Query().Get("pkg"); !isInit {
			pkgName = query
		}
		if query := r.URL.Query().Get("reverse"); query != "" {
			isReverse = query == "true" && pkgName != ""
		}
		w.Header().Add("x-pkg", pkgName)
		w.Header().Add("x-reverse", strconv.FormatBool(isReverse))
		if err := OutputGraphFormat(w, root, pkgName, showStd, showThird, isReverse); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
		}
	}
}
