package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/gogo/protobuf/jsonpb"
)

type xDSHandler struct {
	controller *Controller
}

func (h *xDSHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.URL.Path {
	case "/v2/discovery:endpoints":
		h.handleEDS(w, req)
	case "/v2/discovery:listeners":
		h.handleLDS(w, req)
	case "/v2/discovery:clusters":
		h.handleCDS(w, req)
	case "/config":
		h.handleConfig(w, req)
	case "/healthz":
		http.Error(w, "ok", 200)
	default:
		http.Error(w, "not found", 404)
	}
}

// Endpoint Discovery Service
func (h *xDSHandler) handleEDS(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}

	dr, err := readDiscoveryRequest(req)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	if len(dr.ResourceNames) != 1 {
		http.Error(w, "must have 1 resource_names", 400)
		return
	}

	if b, ok := h.controller.epStore.Get(dr.ResourceNames[0]); ok {
		w.Write(b)
	} else {
		http.Error(w, "not found", 404)
	}
}

// Listener Discovery Service
func (h *xDSHandler) handleLDS(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}

	dr, err := readDiscoveryRequest(req)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	if b, ok := h.controller.configStore.GetConfigSnapshot().GetListeners(dr.Node); ok {
		w.Write(b)
	} else {
		http.Error(w, "not found", 404)
	}
}

// Cluster Discovery Service
func (h *xDSHandler) handleCDS(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}

	dr, err := readDiscoveryRequest(req)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	if b, ok := h.controller.configStore.GetConfigSnapshot().GetClusters(dr.Node); ok {
		w.Write(b)
	} else {
		http.Error(w, "not found", 404)
	}
}

func (h *xDSHandler) handleConfig(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(w, "method not allowed", 405)
		return
	}

	status := 200
	lastError := ""
	lastUpdate := h.controller.configStore.lastUpdate

	if h.controller.configStore.lastError != nil {
		status = 500
		lastError = h.controller.configStore.lastError.Error()
	}

	j, _ := json.Marshal(struct {
		LastError  string    `json:"last_error"`
		LastUpdate time.Time `json:"last_update"`
	}{
		lastError,
		lastUpdate,
	})

	w.WriteHeader(status)
	w.Write(j)
}

func readDiscoveryRequest(req *http.Request) (*v2.DiscoveryRequest, error) {
	var dr v2.DiscoveryRequest
	err := jsonpb.Unmarshal(req.Body, &dr)
	return &dr, err
}
