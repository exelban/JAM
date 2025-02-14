package notify

import (
	"crypto/tls"
	"fmt"
	gomail "gopkg.in/mail.v2"
	"log"
	"sync"
	"time"
)

type SMTP struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	To       []string

	dialer     *gomail.Dialer
	sendCloser gomail.SendCloser
	open       bool
	timer      *time.Timer
	last       *time.Time
	once       sync.Once

	sync.Mutex
}

// dial creates a new connection to the SMTP server.
// close closes the connection to the SMTP server.
func (s *SMTP) dial() error {
	sc, err := s.dialer.Dial()
	if err != nil {
		return fmt.Errorf("dialer dial: %w", err)
	}
	s.sendCloser = sc
	s.open = true

	if s.timer != nil {
		s.timer.Stop()
	}
	s.timer = time.AfterFunc(10*time.Second, s.close)

	return nil
}
func (s *SMTP) close() {
	if err := s.sendCloser.Close(); err != nil {
		log.Printf("[ERROR] smpt close: %v", err)
	}
	s.open = false
}

func (s *SMTP) string() string {
	return "smtp"
}

func (s *SMTP) send(str string) error {
	s.once.Do(func() {
		s.dialer = gomail.NewDialer(s.Host, s.Port, s.Username, s.Password)
		s.dialer.TLSConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	})

	s.Lock()
	defer s.Unlock()

	if s.last != nil && time.Since(*s.last) < time.Second {
		s.Unlock()
		time.Sleep(time.Second - time.Since(*s.last))
		s.Lock()
	} else {
		s.last = nil
	}

	if !s.open {
		if err := s.dial(); err != nil {
			return err
		}
	}

	message := gomail.NewMessage()
	message.SetHeader("From", s.From)
	message.SetHeader("To", s.To...)
	message.SetHeader("Subject", "Status page: event triggered")
	message.SetBody("text/plain", str)

	if err := gomail.Send(s.sendCloser, message); err != nil {
		log.Printf("[ERROR] send email: %v", err)
	}

	now := time.Now()
	s.last = &now

	return nil
}
