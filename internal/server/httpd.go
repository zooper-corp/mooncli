package server

import (
	"encoding/json"
	"fmt"
	"github.com/NYTimes/gziphandler"
	"github.com/zooper-corp/mooncli/config"
	"log"
	"net/http"
	"strings"
	"time"
)

func (c *ChainData) HandleInfo(w http.ResponseWriter, r *http.Request) {
	info := c.GetInfo()
	handleJsonResponse(w, info)
}

func (c *ChainData) HandleHealth(w http.ResponseWriter, r *http.Request) {
	info := c.GetInfo()
	update := time.Unix(int64(info.Update.TsSecs), 0)
	delta := time.Now().Unix() - update.Unix()
	if delta > int64(c.maxUpdateDelta.Seconds()) {
		w.WriteHeader(500)
		_, _ = w.Write([]byte(fmt.Sprintf("error, %v > %v", delta, c.maxUpdateDelta.Seconds())))
	} else {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(fmt.Sprintf("ok, %v < %v", delta, c.maxUpdateDelta.Seconds())))
	}
}

func (c *ChainData) HandleCollators(w http.ResponseWriter, r *http.Request) {
	stats := c.GetCollators()
	handleJsonResponse(w, stats)
}

func (c *ChainData) HandleDelegations(w http.ResponseWriter, r *http.Request) {
	p := strings.Split(r.URL.Path, "/")
	if len(p) == 3 {
		address := p[2]
		stats := c.GetDelegations(address)
		if len(stats.Delegations) > 0 {
			handleJsonResponse(w, stats)
			return
		} else {
			http.Error(w, fmt.Sprintf("Delegator '%v' not found", address), 404)
			return
		}
	} else {
		http.Error(w, "Invalid arguments", 400)
	}
}

func (c *ChainData) HandleCollator(w http.ResponseWriter, r *http.Request) {
	p := strings.Split(r.URL.Path, "/")
	if len(p) == 3 {
		address := p[2]
		stats := c.GetCollator(address)
		if len(stats.Collators) > 0 {
			handleJsonResponse(w, stats)
			return
		} else {
			http.Error(w, fmt.Sprintf("Collator '%v' not found", address), 404)
			return
		}
	} else {
		http.Error(w, "Invalid arguments", 400)
	}
}

func handleJsonResponse(w http.ResponseWriter, data any) {
	// Cors
	w.Header().Set("Access-Control-Allow-Origin", "*")
	// Minimum caching
	w.Header().Set("Cache-Control", "max-age:90, stale-if-error=600")
	// Json
	js, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Write
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, err = w.Write(js)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func ServeChainData(config config.HttpConfig) {
	chainData, err := NewChainData(
		config.ChainConfig,
		config.UpdateInterval*3/2,
	)
	if err != nil {
		panic(err)
	}
	// First update
	err = chainData.Update()
	if err != nil {
		panic(err)
	}
	// Start update routine
	ticker := time.NewTicker(config.UpdateInterval)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				_ = chainData.Update()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
	// Live probes
	http.HandleFunc("/healthz", chainData.HandleHealth)
	// Generic info page with last update data
	http.Handle("/info", gziphandler.GzipHandler(http.HandlerFunc(chainData.HandleInfo)))
	// Main stats (for a collator or all)
	http.Handle("/collators/", gziphandler.GzipHandler(http.HandlerFunc(chainData.HandleCollator)))
	http.Handle("/collators", gziphandler.GzipHandler(http.HandlerFunc(chainData.HandleCollators)))
	// Delegations
	http.Handle("/delegations/", gziphandler.GzipHandler(http.HandlerFunc(chainData.HandleDelegations)))
	// Start engine
	log.Fatal(http.ListenAndServe(config.Addr, nil))
}
