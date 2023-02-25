package main

import (
	"io"
	"net"
	"net/http"
	"time"
)

type HTTPProcessor struct {
	config *Config
}

func (s *HTTPProcessor) handleTunneling(w http.ResponseWriter, req *http.Request) {
	// Get endpoint config
	_, ok := s.config.getDomainEndpoint(req.URL.Host, req.Method)
	if !ok {
		// Processing of this domain is not allowed
		http.Error(w, "Requested domain is not whitelisted", http.StatusForbidden)
		return
	}

	dest_conn, err := net.DialTimeout("tcp", req.Host, 10*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	client_conn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}
	go transfer(dest_conn, client_conn)
	go transfer(client_conn, dest_conn)
}
func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source)
}

func (s *HTTPProcessor) handleHTTP(w http.ResponseWriter, req *http.Request) {
	// Get endpoint config
	eid, ok := s.config.getDomainEndpoint(req.URL.Host, req.Method)
	if !ok {
		// Processing of this domain is not allowed
		http.Error(w, "Requested domain is not whitelisted", http.StatusForbidden)
		return
	}

	pact, ok := s.config.getProxyAction(eid)
	if !ok {
		// Action is not configured
		http.Error(w, "Action for this domain is not configured", http.StatusForbidden)
		return
	}

	if pact.XForwardedFor {
		// Append X-Forwarded-For
		req.Header["x-forwarded-for"] = []string{req.RemoteAddr}
	}

	// Add extra headers
	for _, hdr := range pact.AddHeaders {
		req.Header[hdr.Key] = append(req.Header[hdr.Key], hdr.Value)
	}

	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	resp.Header["x-proxy-name"] = []string{"Go-proxy 0.001"}
	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func (s *HTTPProcessor) requestHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodConnect {
		s.handleTunneling(w, r)
	} else {
		s.handleHTTP(w, r)
	}
}
