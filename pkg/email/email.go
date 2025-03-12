package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
)

type Cfg struct {
	EnableTLS      bool
	SenderAddress  string
	SenderName     string
	SenderPwd      string
	MailServerHost string
	MailServerPort int
}

type Msg struct {
	ToAddress string
	Title     string
	Content   string
}

func SendEmail(cfg Cfg, msg Msg) error {
	auth := smtp.PlainAuth("", cfg.SenderAddress, cfg.SenderPwd, cfg.MailServerHost)

	to := []string{msg.ToAddress}
	sender := cfg.SenderAddress
	nickname := cfg.SenderName

	subject := msg.Title
	contentType := "Content-Type: text/html; charset=UTF-8"
	body := msg.Content
	msgContent := "To:" + strings.Join(to, ",") + "\r\nFrom: "
	msgContent += nickname + "<" + sender + ">\r\nSubject: " + subject
	msgContent += "\r\n" + contentType + "\r\n\r\n" + body
	return sendMailUsingTLS(cfg.EnableTLS, cfg.MailServerHost, cfg.MailServerPort, auth, sender, to, []byte(msgContent))
}

func dial(enableTLS bool, host string, port int) (*smtp.Client, error) {
	var tlsconfig *tls.Config
	if enableTLS {
		tlsconfig = &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         host,
		}
	}
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", host, port), tlsconfig)
	if err != nil {
		return nil, err
	}
	return smtp.NewClient(conn, host)
}

func sendMailUsingTLS(enableTLS bool, host string, port int, auth smtp.Auth, from string,
	to []string, msg []byte) (err error) {
	c, err := dial(enableTLS, host, port)
	if err != nil {
		return err
	}
	defer c.Close()
	if auth != nil {
		if ok, _ := c.Extension("AUTH"); ok {
			if err = c.Auth(auth); err != nil {
				return err
			}
		}
	}
	if err = c.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}
