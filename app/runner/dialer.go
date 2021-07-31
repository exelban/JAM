package runner

import (
	"context"
	"github.com/exelban/cheks/app/types"
	"log"
	"net/http"
)

// Dialer - the request maker structure
type Dialer struct {
	sem chan int
}

// NewDialer - creates a new dialer with maxConn semaphore
func NewDialer(maxConn int) *Dialer {
	return &Dialer{
		sem: make(chan int, maxConn),
	}
}

// Dial - make a http request to the provided host
func (d *Dialer) Dial(ctx context.Context, h *types.Host) (int, bool) {
	d.sem <- 1
	defer func() {
		<-d.sem
	}()

	code := make(chan int, 1)
	ok := make(chan bool, 1)
	go func() {
		req, err := http.NewRequest(h.Method, h.URL, nil)
		if err != nil {
			code <- 0
			ok <- false
			log.Printf("[ERROR] prepare request %v", err)
			return
		}
		req.WithContext(ctx)

		client := http.Client{
			Timeout: h.TimeoutInterval,
		}
		resp, err := client.Do(req)
		if err != nil {
			code <- 0
			ok <- false
			log.Printf("[ERROR] make request %v", err)
			return
		}

		code <- resp.StatusCode
		ok <- true
	}()

	return <-code, <-ok
}
