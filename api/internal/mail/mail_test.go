package mail

import (
	"bufio"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"net/mail"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestValidate(t *testing.T) {
	ok := SMTPConfig{Host: "smtp.example.com", Port: 465, From: "WeBond <a@example.com>", TLSMode: TLSImplicit, Username: "a@example.com", Password: "x"}
	if err := ok.Validate(); err != nil {
		t.Fatalf("valid config rejected: %v", err)
	}
	cases := map[string]SMTPConfig{
		"no host":           {Port: 25, From: "a@example.com", TLSMode: TLSNone},
		"bad port":          {Host: "h", Port: 0, From: "a@example.com", TLSMode: TLSNone},
		"bad from":          {Host: "h", Port: 25, From: "not an address", TLSMode: TLSNone},
		"bad tls":           {Host: "h", Port: 25, From: "a@example.com", TLSMode: "ssl3"},
		"creds over plain":  {Host: "h", Port: 25, From: "a@example.com", TLSMode: TLSNone, Username: "u", Password: "p"},
		"user, no password": {Host: "h", Port: 587, From: "a@example.com", TLSMode: TLSStart, Username: "u"},
	}
	for name, c := range cases {
		if err := c.Validate(); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestValidate_NeverEchoesPassword(t *testing.T) {
	c := SMTPConfig{Host: "h", Port: 25, From: "bad", TLSMode: TLSNone, Password: "super-secret-pw"}
	if err := c.Validate(); err == nil || strings.Contains(err.Error(), "super-secret-pw") {
		t.Fatalf("error must exist and not contain the password: %v", err)
	}
}

func TestBuildMessage_HeadersAndEncoding(t *testing.T) {
	from := &mail.Address{Name: "WeBond", Address: "no-reply@example.com"}
	to := &mail.Address{Address: "investor@example.com"}
	raw, err := BuildMessage(from, to, "ยืนยันอีเมล\r\nBcc: attacker@evil.test", "line1\nline2", time.Date(2026, 11, 12, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	s := string(raw)
	head, body, found := strings.Cut(s, "\r\n\r\n")
	if !found {
		t.Fatal("no header/body separator")
	}
	for _, line := range strings.Split(head, "\r\n") {
		if strings.HasPrefix(strings.ToLower(line), "bcc:") {
			t.Fatalf("header injection succeeded: %q", line)
		}
	}
	if !strings.Contains(head, "Subject: =?utf-8?q?") {
		t.Fatalf("subject not RFC 2047 encoded: %s", head)
	}
	if !strings.Contains(head, "Content-Type: text/plain; charset=UTF-8") || !strings.Contains(head, "Message-ID: <") {
		t.Fatalf("missing headers: %s", head)
	}
	if !strings.Contains(body, "line1\r\nline2") {
		t.Fatalf("body line endings not CRLF: %q", body)
	}
}

func TestMaskAddress(t *testing.T) {
	if got := MaskAddress("investor@example.com"); got != "i***@example.com" {
		t.Fatalf("got %q", got)
	}
	if got := MaskAddress("nope"); got != "***" {
		t.Fatalf("got %q", got)
	}
}

// fakeSMTP is a minimal plain SMTP server that records one message.
type fakeSMTP struct {
	ln       net.Listener
	mu       sync.Mutex
	mailFrom string
	rcptTo   string
	data     string
}

func startFakeSMTP(t *testing.T) *fakeSMTP {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	f := &fakeSMTP{ln: ln}
	t.Cleanup(func() { _ = ln.Close() })
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		r := bufio.NewReader(conn)
		w := func(s string) { _, _ = conn.Write([]byte(s + "\r\n")) }
		w("220 fake ESMTP")
		for {
			line, err := r.ReadString('\n')
			if err != nil {
				return
			}
			cmd := strings.TrimSpace(line)
			upper := strings.ToUpper(cmd)
			switch {
			case strings.HasPrefix(upper, "EHLO"), strings.HasPrefix(upper, "HELO"):
				w("250 fake")
			case strings.HasPrefix(upper, "MAIL FROM:"):
				f.mu.Lock()
				f.mailFrom = cmd[len("MAIL FROM:"):]
				f.mu.Unlock()
				w("250 ok")
			case strings.HasPrefix(upper, "RCPT TO:"):
				f.mu.Lock()
				f.rcptTo = cmd[len("RCPT TO:"):]
				f.mu.Unlock()
				w("250 ok")
			case upper == "DATA":
				w("354 go ahead")
				var b strings.Builder
				for {
					l, err := r.ReadString('\n')
					if err != nil {
						return
					}
					if l == ".\r\n" {
						break
					}
					b.WriteString(l)
				}
				f.mu.Lock()
				f.data = b.String()
				f.mu.Unlock()
				w("250 queued")
			case upper == "QUIT":
				w("221 bye")
				return
			default:
				w("250 ok")
			}
		}
	}()
	return f
}

func TestSMTPSender_PlainDelivery(t *testing.T) {
	srv := startFakeSMTP(t)
	_, portStr, _ := net.SplitHostPort(srv.ln.Addr().String())
	port, _ := strconv.Atoi(portStr)

	s, err := NewSMTPSender(SMTPConfig{Host: "127.0.0.1", Port: port, From: "WeBond <no-reply@example.com>", TLSMode: TLSNone})
	if err != nil {
		t.Fatalf("new sender: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.Send(ctx, Message{To: "investor@example.com", Subject: "Verify", Text: "http://localhost:5173/verify-email?token=abc"}); err != nil {
		t.Fatalf("send: %v", err)
	}
	srv.mu.Lock()
	defer srv.mu.Unlock()
	if !strings.Contains(srv.mailFrom, "no-reply@example.com") || !strings.Contains(srv.rcptTo, "investor@example.com") {
		t.Fatalf("envelope wrong: from=%q to=%q", srv.mailFrom, srv.rcptTo)
	}
	if !strings.Contains(srv.data, "verify-email?token=3Dabc") && !strings.Contains(srv.data, "verify-email?token=abc") {
		t.Fatalf("body missing link: %q", srv.data)
	}
}

func TestSMTPSender_StartTLSRequired(t *testing.T) {
	srv := startFakeSMTP(t) // offers no STARTTLS
	_, portStr, _ := net.SplitHostPort(srv.ln.Addr().String())
	port, _ := strconv.Atoi(portStr)
	s, err := NewSMTPSender(SMTPConfig{Host: "127.0.0.1", Port: port, From: "a@example.com", TLSMode: TLSStart})
	if err != nil {
		t.Fatalf("new sender: %v", err)
	}
	err = s.Send(context.Background(), Message{To: "b@example.com", Subject: "s", Text: "t"})
	if err == nil || !strings.Contains(err.Error(), "STARTTLS") {
		t.Fatalf("expected STARTTLS refusal, got %v", err)
	}
}

// TestSMTPSender_ImplicitTLSVerifiesCertificate: with TLS (port 465 style) the
// server certificate must be verified; an untrusted self-signed server is
// refused before any credentials are sent.
func TestSMTPSender_ImplicitTLSVerifiesCertificate(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "127.0.0.1"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	ln, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{
		Certificates: []tls.Certificate{{Certificate: [][]byte{der}, PrivateKey: key}},
		MinVersion:   tls.VersionTLS12,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	gotCredentials := make(chan bool, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		_, _ = conn.Write([]byte("220 fake\r\n"))
		buf := make([]byte, 512)
		n, _ := conn.Read(buf) // fails if the client aborted the handshake
		gotCredentials <- strings.Contains(string(buf[:n]), "AUTH")
	}()

	_, portStr, _ := net.SplitHostPort(ln.Addr().String())
	port, _ := strconv.Atoi(portStr)
	s, err := NewSMTPSender(SMTPConfig{Host: "127.0.0.1", Port: port, From: "a@example.com",
		TLSMode: TLSImplicit, Username: "a@example.com", Password: "app-password"})
	if err != nil {
		t.Fatalf("new sender: %v", err)
	}
	err = s.Send(context.Background(), Message{To: "b@example.com", Subject: "s", Text: "t"})
	if err == nil || !strings.Contains(err.Error(), "certificate") {
		t.Fatalf("expected a certificate verification error, got %v", err)
	}
	select {
	case sent := <-gotCredentials:
		if sent {
			t.Fatal("credentials were sent to an untrusted server")
		}
	case <-time.After(2 * time.Second):
	}
}

func TestSMTPSender_RejectsBadRecipient(t *testing.T) {
	s, _ := NewSMTPSender(SMTPConfig{Host: "127.0.0.1", Port: 1, From: "a@example.com", TLSMode: TLSNone})
	if err := s.Send(context.Background(), Message{To: "not-an-address", Subject: "s", Text: "t"}); err == nil {
		t.Fatal("expected invalid recipient error")
	}
}
