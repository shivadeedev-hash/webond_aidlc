// Package mail delivers plain-text e-mail over SMTP using only the Go standard
// library (Task 10.8). It supports implicit TLS (port 465, e.g. Gmail),
// STARTTLS (port 587) and plain SMTP for a local catcher such as Mailpit.
// When no SMTP server is configured, LogSender records that a message was not
// delivered without ever writing its body (which may contain one-time tokens).
package mail

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"mime"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

// TLS modes for the SMTP connection.
const (
	TLSImplicit = "tls"      // TLS from the first byte (port 465)
	TLSStart    = "starttls" // plain connect, then mandatory STARTTLS (port 587)
	TLSNone     = "none"     // no encryption; only for a local mail catcher
)

// defaultTimeout bounds one whole SMTP conversation.
const defaultTimeout = 20 * time.Second

// Message is one plain-text e-mail.
type Message struct {
	To      string
	Subject string
	Text    string
}

// Sender delivers a Message.
type Sender interface {
	Send(ctx context.Context, msg Message) error
	// Name identifies the sender in logs ("smtp" or "log").
	Name() string
}

// SMTPConfig configures SMTPSender.
type SMTPConfig struct {
	Host     string
	Port     int
	Username string // empty → no authentication (local catcher)
	Password string
	From     string // "WeBond <no-reply@example.com>" or a bare address
	TLSMode  string // TLSImplicit | TLSStart | TLSNone
	Timeout  time.Duration
}

// Validate checks the configuration; it never echoes the password.
func (c SMTPConfig) Validate() error {
	if strings.TrimSpace(c.Host) == "" {
		return errors.New("SMTP host is required")
	}
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("SMTP port %d is out of range", c.Port)
	}
	if _, err := mail.ParseAddress(c.From); err != nil {
		return fmt.Errorf("SMTP from address %q is invalid: %w", c.From, err)
	}
	switch c.TLSMode {
	case TLSImplicit, TLSStart:
	case TLSNone:
		// Never send credentials over an unencrypted connection.
		if c.Username != "" {
			return errors.New("SMTP_TLS=none cannot be combined with SMTP_USERNAME (credentials would be sent in clear text)")
		}
	default:
		return fmt.Errorf("SMTP TLS mode %q is invalid (use tls, starttls or none)", c.TLSMode)
	}
	if c.Username != "" && c.Password == "" {
		return errors.New("SMTP_PASSWORD is required when SMTP_USERNAME is set")
	}
	return nil
}

// SMTPSender sends mail through an SMTP server.
type SMTPSender struct {
	cfg  SMTPConfig
	from *mail.Address
}

// NewSMTPSender validates cfg and returns a sender.
func NewSMTPSender(cfg SMTPConfig) (*SMTPSender, error) {
	if cfg.Timeout <= 0 {
		cfg.Timeout = defaultTimeout
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	from, _ := mail.ParseAddress(cfg.From) // validated above
	return &SMTPSender{cfg: cfg, from: from}, nil
}

// Name implements Sender.
func (s *SMTPSender) Name() string { return "smtp" }

// Send implements Sender.
func (s *SMTPSender) Send(ctx context.Context, msg Message) error {
	to, err := mail.ParseAddress(msg.To)
	if err != nil {
		return fmt.Errorf("invalid recipient: %w", err)
	}
	raw, err := BuildMessage(s.from, to, msg.Subject, msg.Text, time.Now())
	if err != nil {
		return err
	}

	deadline := time.Now().Add(s.cfg.Timeout)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	addr := net.JoinHostPort(s.cfg.Host, strconv.Itoa(s.cfg.Port))
	tlsCfg := &tls.Config{ServerName: s.cfg.Host, MinVersion: tls.VersionTLS12}
	dialer := &net.Dialer{Deadline: deadline}

	var conn net.Conn
	if s.cfg.TLSMode == TLSImplicit {
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, tlsCfg)
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", addr)
	}
	if err != nil {
		return fmt.Errorf("connect to SMTP server: %w", err)
	}
	_ = conn.SetDeadline(deadline)

	c, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("SMTP handshake: %w", err)
	}
	defer func() { _ = c.Close() }()

	if s.cfg.TLSMode == TLSStart {
		if ok, _ := c.Extension("STARTTLS"); !ok {
			return errors.New("SMTP server does not offer STARTTLS")
		}
		if err := c.StartTLS(tlsCfg); err != nil {
			return fmt.Errorf("SMTP STARTTLS: %w", err)
		}
	}
	if s.cfg.Username != "" {
		// PlainAuth only sends credentials over TLS (or to localhost).
		if err := c.Auth(smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)); err != nil {
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}
	}
	if err := c.Mail(s.from.Address); err != nil {
		return fmt.Errorf("SMTP MAIL FROM: %w", err)
	}
	if err := c.Rcpt(to.Address); err != nil {
		return fmt.Errorf("SMTP RCPT TO: %w", err)
	}
	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("SMTP DATA: %w", err)
	}
	if _, err := w.Write(raw); err != nil {
		_ = w.Close()
		return fmt.Errorf("SMTP write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("SMTP end of data: %w", err)
	}
	return c.Quit()
}

// BuildMessage renders an RFC 5322 plain-text UTF-8 message. Header values are
// stripped of CR/LF (header injection) and the subject is RFC 2047 encoded so
// Thai text is safe.
func BuildMessage(from, to *mail.Address, subject, text string, now time.Time) ([]byte, error) {
	id, err := messageID(from.Address)
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	header := func(k, v string) { b.WriteString(k + ": " + v + "\r\n") }
	header("From", from.String())
	header("To", to.String())
	header("Subject", mime.QEncoding.Encode("utf-8", oneLine(subject)))
	header("Date", now.Format(time.RFC1123Z))
	header("Message-ID", id)
	header("MIME-Version", "1.0")
	header("Content-Type", "text/plain; charset=UTF-8")
	header("Content-Transfer-Encoding", "quoted-printable")
	b.WriteString("\r\n")

	qp := quotedprintable.NewWriter(&b)
	if _, err := qp.Write([]byte(strings.ReplaceAll(text, "\n", "\r\n"))); err != nil {
		return nil, fmt.Errorf("encode body: %w", err)
	}
	if err := qp.Close(); err != nil {
		return nil, fmt.Errorf("encode body: %w", err)
	}
	return b.Bytes(), nil
}

// oneLine removes CR/LF so a value can never start a new header.
func oneLine(s string) string {
	return strings.NewReplacer("\r", " ", "\n", " ").Replace(s)
}

// messageID builds a unique Message-ID in the sender's domain.
func messageID(fromAddr string) (string, error) {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("message id: %w", err)
	}
	domain := "webond.local"
	if at := strings.LastIndex(fromAddr, "@"); at >= 0 && at < len(fromAddr)-1 {
		domain = fromAddr[at+1:]
	}
	return "<" + hex.EncodeToString(buf) + "@" + domain + ">", nil
}

// LogSender is used when no SMTP server is configured. It only logs that a
// message was not delivered: recipient (masked) and subject, never the body.
type LogSender struct{ Logger *slog.Logger }

// Name implements Sender.
func (s LogSender) Name() string { return "log" }

// Send implements Sender.
func (s LogSender) Send(_ context.Context, msg Message) error {
	logger := s.Logger
	if logger == nil {
		logger = slog.Default()
	}
	logger.Info("email not delivered: no SMTP server configured (set SMTP_HOST)",
		"to", MaskAddress(msg.To), "subject", msg.Subject)
	return nil
}

// MaskAddress hides most of the local part: "investor@example.com" → "i***@example.com".
func MaskAddress(addr string) string {
	at := strings.LastIndex(addr, "@")
	if at <= 0 {
		return "***"
	}
	return addr[:1] + "***" + addr[at:]
}
