package email

import (
	"bytes"
	"fmt"
	"net/smtp"
	"os"
	"strings"
	"text/template"

	"github.com/BrunoTeixeira1996/nutgmonitor/internal/logger"
)

type EmailTemplate struct {
	Timestamp string
	Content   string
}

type Smtp struct {
	Host    string
	Port    string
	Address string
}

func (s *Smtp) setSmtpValues() {
	s.Host = "smtp.gmail.com"
	s.Port = "587"
	s.Address = s.Host + ":" + s.Port
}

func GetEnvs() (string, string) {
	senderEmail := os.Getenv("SENDEREMAIL")
	senderPass := os.Getenv("SENDERPASS")

	return senderEmail, senderPass
}

func buildEmail(e *EmailTemplate) (string, error) {
	currentPath, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("[email error] could not get current path: %s\n", err)
	}

	logFile := fmt.Sprintf("%s/logs/log.txt", currentPath)

	logstdoutFile, err := os.ReadFile(logFile)
	if err != nil {
		return "", fmt.Errorf("[email error] could not read file logstdout: %s\n", err)
	}

	buf := bytes.NewBuffer(logstdoutFile)

	for _, s := range strings.Split(buf.String(), "\n") {
		e.Content += fmt.Sprintf("%s<br>", s)
	}

	pF := fmt.Sprintf("%s/email_template.html", currentPath)
	templ, err := template.New("email_template.html").ParseFiles(pF)
	if err != nil {
		return "", fmt.Errorf("[email error] could not parse files: %s\n", err)
	}

	var outTemp bytes.Buffer
	if err := templ.Execute(&outTemp, e); err != nil {
		return "", fmt.Errorf("[email error] could not execute template: %s\n", err)
	}

	return outTemp.String(), nil
}

func SendEmail(body *EmailTemplate) error {
	senderEmail, senderPass := GetEnvs()

	s := &Smtp{}
	s.setSmtpValues()

	recipientEmail := "brunoalexandre3@hotmail.com"
	headers := "Content-Type: text/html; charset=ISO-8859-1\r\n" // used to send HTML

	from := fmt.Sprintf("From: <%s>\r\n", senderEmail)
	to := fmt.Sprintf("To: <%s>\r\n", recipientEmail)
	subject := "Subject: UPS Outage\r\n"

	logger.Log.Printf("[email] building email body\n")
	finalBody, err := buildEmail(body)
	if err != nil {
		return fmt.Errorf("[email error] could not build email template: %s", err)
	}

	msg := headers + from + to + subject + "\r\n" + finalBody + "\r\n"

	auth := smtp.PlainAuth("", senderEmail, senderPass, s.Host)

	if err := smtp.SendMail(s.Address, auth, senderEmail, []string{recipientEmail}, []byte(msg)); err != nil {
		return fmt.Errorf("[email error] could not send email: %s", err)
	}

	logger.Log.Printf("[email info] sent email to %s\n", recipientEmail)

	return nil
}
