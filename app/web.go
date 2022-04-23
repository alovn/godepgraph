package app

import (
	"net/http"
	"strconv"

	"github.com/alovn/godepgraph/web"
)

func Serve(path, addr, findPkgName string, isShowStdLib, isShowThirdLib, isReverse bool) error {
	http.HandleFunc("/graph", graphHandler(path, findPkgName, isShowStdLib, isShowThirdLib, isReverse))
	return web.Serve(addr)
}

func graphHandler(path, findPkgName string, isShowStdLib, isShowThirdLib, isReverse bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		showStd := isShowStdLib
		pkgName := findPkgName
		showThird := isShowThirdLib
		reverse := isReverse && pkgName != ""
		isInit := r.URL.Query().Get("init") == "true"
		mod := r.URL.Query().Get("mod") == "true"
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
			reverse = query == "true" && (pkgName != "" || mod)
		}
		w.Header().Add("x-pkg", pkgName)
		w.Header().Add("x-reverse", strconv.FormatBool(isReverse))
		if err := OutputGraphFormat(w, path, pkgName, showStd, showThird, reverse, mod); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
		}
	}
}
