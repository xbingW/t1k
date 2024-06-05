package main

import (
	"net/http"
	"os"

	"github.com/xbingW/t1k/detector"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		d := detector.NewDetector(detector.Config{
			Addr: os.Getenv("DETECTOR_ADDR"),
		})
		res, err := d.DetectorRequest(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !res.Allowed() {
			http.Error(w, "blocked", http.StatusForbidden)
			return
		}
		_, _ = w.Write([]byte("allowed"))
	})
	_ = http.ListenAndServe(":80", nil)
}
