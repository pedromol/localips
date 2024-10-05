package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"strconv"
	"time"
)

type Response struct {
	Elapsed time.Duration `json:"elapsed"`
	IPs     []string      `json:"ips"`
}

func GetIPs(expected int) []string {
	ips := make([]string, 0)

	for i := 0; i < 100 && len(ips) < expected; i++ {
		c := http.Client{Timeout: time.Duration(10) * time.Second}
		c.CloseIdleConnections()
		res, err := c.Get("https://api.ipify.org")
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		resBody, err := io.ReadAll(res.Body)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}
		res.Body.Close()
		if !slices.Contains(ips, string(resBody)) {
			ips = append(ips, string(resBody))
		}

	}

	return ips
}

func main() {
	ipsvar := os.Getenv("NUMIPS")
	ipsnum := 2
	if ipsvar != "" {
		ipsnum, _ = strconv.Atoi(ipsvar)
	}
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}

	var ips []string
	lastCheck := time.Unix(0, 0)
	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		if time.Since(lastCheck) > time.Minute*10 {
			ips = GetIPs(ipsnum)
			lastCheck = time.Now()
		}
		res := Response{
			Elapsed: time.Since(start),
			IPs:     ips,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)

		if err := json.NewEncoder(w).Encode(res); err != nil {
			slog.Error("Failed to encode response", "error", err)
		}
		slog.Info("Response", "elapsed", res.Elapsed, "ips", res.IPs)
	}))
	http.Handle("/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))

	server := &http.Server{
		Addr:         addr,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
	}
	slog.Info("Starting HTTP server", "addr", server.Addr, "ips", ipsnum)
	slog.Error("Failed to start HTTP server", "error", server.ListenAndServe())
}
