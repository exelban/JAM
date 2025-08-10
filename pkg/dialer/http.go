package dialer

import (
	"context"
	"crypto/tls"
	"github.com/exelban/JAM/types"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptrace"
	"strings"
	"time"
)

// httpCall makes a HTTP request to the host
func (d *Dialer) httpCall(ctx context.Context, h *types.Host) (response types.HttpResponse) {
	req, err := http.NewRequest(h.Method, h.URL, nil)
	if err != nil {
		log.Printf("[ERROR] prepare request %v", err)
		return
	}

	var start, connect, dns, tlsHandshake time.Time
	var tlsState *tls.ConnectionState
	req = req.WithContext(httptrace.WithClientTrace(ctx, &httptrace.ClientTrace{
		DNSStart:          func(dsi httptrace.DNSStartInfo) { dns = time.Now() },
		DNSDone:           func(ddi httptrace.DNSDoneInfo) { response.DNS = time.Since(dns) },
		TLSHandshakeStart: func() { tlsHandshake = time.Now() },
		TLSHandshakeDone: func(cs tls.ConnectionState, err error) {
			response.TLSHandshake = time.Since(tlsHandshake)
			tlsState = &cs
		},
		ConnectStart:         func(network, addr string) { connect = time.Now() },
		ConnectDone:          func(network, addr string, err error) { response.Connect = time.Since(connect) },
		GotFirstResponseByte: func() { response.TTFB = time.Since(start) },
	}))

	for key, value := range h.Headers {
		req.Header.Set(key, value)
	}

	client := http.Client{
		Transport: &http.Transport{
			ResponseHeaderTimeout: time.Second * 5,
			DialContext: (&net.Dialer{
				Timeout:   time.Second * 30,
				KeepAlive: time.Second * 30,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       time.Second * 30,
			TLSHandshakeTimeout:   time.Second * 30,
			ExpectContinueTimeout: time.Second * 30,
		},
	}
	if h.TimeoutInterval != nil {
		client.Timeout = *h.TimeoutInterval
	}

	response.Timestamp = time.Now()

	startTime := time.Now()
	resp, err := client.Do(req)
	response.Time = time.Since(startTime)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			response.Code = 522
		} else if opErr, ok := err.(*net.OpError); ok {
			if opErr.Op == "dial" {
				response.Code = 523
			} else if opErr.Op == "read" {
				response.Code = 521
			}
		} else {
			if err.Error() != "" && (strings.Contains(err.Error(), "refused") || strings.Contains(err.Error(), "unreachable")) {
				response.Code = 523
			} else {
				response.Code = http.StatusServiceUnavailable
			}
		}
		return
	}
	defer resp.Body.Close()
	response.Code = resp.StatusCode

	if tlsState != nil && len(tlsState.PeerCertificates) > 0 {
		response.SSLCertExpiry = &tlsState.PeerCertificates[0].NotAfter
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[ERROR] read body %v", err)
		return
	}
	if len(b) < 1024 {
		response.Bytes = b
	}
	response.OK = true

	return
}
