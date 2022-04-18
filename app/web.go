package app

import (
	"net/http"
)

func Serve(root string, addr string, showStdLib bool) error {
	mux := http.DefaultServeMux
	mux.HandleFunc("/graph", graphHandler(root, showStdLib))
	return http.ListenAndServe(addr, mux)
}

func graphHandler(root string, showStdLib bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := OutputGraphFormat(w, root, showStdLib); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
		}
	}
}
