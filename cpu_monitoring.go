package main

import "net/http"

type CPUMonitorCtx struct {
}

// HandleGetCPUPercent is used to handle request for getting cpu usage as stream of data.
func HandleGetCPUPercent(ctx *CPUMonitorCtx) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

	})
}

func getCPUUsage() float32 {
	return 0
}
