package dialer

import (
	"context"
	"crypto/tls"
	"github.com/exelban/cheks/types"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptrace"
	"time"
)

// Dialer - the request maker structure
type Dialer struct {
	sem chan int
}

// New - creates a new dialer with maxConn semaphore
func New(maxConn int) *Dialer {
	return &Dialer{
		sem: make(chan int, maxConn),
	}
}

// Dial - make a http request to the provided host
func (d *Dialer) Dial(ctx context.Context, h *types.Host) types.HttpResponse {
	d.sem <- 1
	defer func() {
		<-d.sem
	}()

	resp := make(chan types.HttpResponse, 1)
	go func() {
		resp <- d.call(ctx, h)
	}()

	return <-resp
}

func (d *Dialer) call(ctx context.Context, h *types.Host) (response types.HttpResponse) {
	req, err := http.NewRequest(h.Method, h.URL, nil)
	if err != nil {
		log.Printf("[ERROR] prepare request %v", err)
		return
	}

	var start, connect, dns, tlsHandshake time.Time
	req = req.WithContext(httptrace.WithClientTrace(ctx, &httptrace.ClientTrace{
		DNSStart:             func(dsi httptrace.DNSStartInfo) { dns = time.Now() },
		DNSDone:              func(ddi httptrace.DNSDoneInfo) { response.DNS = time.Since(dns) },
		TLSHandshakeStart:    func() { tlsHandshake = time.Now() },
		TLSHandshakeDone:     func(cs tls.ConnectionState, err error) { response.TLSHandshake = time.Since(tlsHandshake) },
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
		Timeout: h.TimeoutInterval,
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[ERROR] make request %v", err)
		return
	}
	response.Code = resp.StatusCode

	b, err := ioutil.ReadAll(resp.Body)
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
