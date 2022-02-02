package main

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"net/smtp"
	"strconv"
)

//go:embed email-templates
var emailTemplates embed.FS

func (app *application) SendMail(from, to, subject, templName string, data interface{}) error {

	templateToRender := fmt.Sprintf("email-templates/%s.html", templName)

	templ, err := template.New("email-html").ParseFS(emailTemplates, templateToRender)

	if err != nil {
		app.errorLog.Println(err)
		return err
	}

	var tplBuffer bytes.Buffer

	if err = templ.ExecuteTemplate(&tplBuffer, "body", data); err != nil {
		app.errorLog.Println(err)
		return err
	}

	htmlMessage := tplBuffer.String()

	stringTemplateToRender := fmt.Sprintf("email-templates/%s.plain.tmpl", templName)
	templ, err = template.New("email-plain").ParseFS(emailTemplates, stringTemplateToRender)

	if err != nil {
		app.errorLog.Println(err)
		return err
	}

	var stringTplBuffer bytes.Buffer

	if err = templ.ExecuteTemplate(&stringTplBuffer, "body", data); err != nil {
		app.errorLog.Println(err)
		return err
	}

	_ = stringTplBuffer.String()

	addr := app.config.smtp.host + ":" + strconv.Itoa(app.config.smtp.port)

	auth := smtp.PlainAuth("", app.config.smtp.username, app.config.smtp.password, app.config.smtp.host)
	err = smtp.SendMail(addr, auth, from, []string{to}, []byte(htmlMessage))
	if err != nil {
		app.errorLog.Println(err)
		return err
	}

	// app.errorLog.Println(htmlMessage, stringMessage)

	/*

		server := mail.NewSMTPClient()

		log.Println(app.config.smtp)
		// SMTP Server
		server.Host = app.config.smtp.host
		server.Port = app.config.smtp.port
		server.Username = app.config.smtp.username
		server.Password = app.config.smtp.password
		// server.Encryption = mail.EncryptionSTARTTLS
		// server.ConnectTimeout = 50 * time.Second
		// server.SendTimeout = 50 * time.Second
		// Variable to keep alive connection
		server.KeepAlive = false

		server.TLSConfig = &tls.Config{InsecureSkipVerify: true}

		smtpClient, err := server.Connect()

		if err != nil {
			app.errorLog.Println(err, "The error place")
			return err
		}

		email := mail.NewMSG()

		err = email.AddTo(to).
			SetFrom(from).
			SetSubject(subject).
			SetBody(mail.TextHTML, htmlMessage).
			// AddAlternative(mail.TextPlain, stringMessage).
			Send(smtpClient)

		if err != nil {
			app.errorLog.Println(err)
			return err
		}

		app.errorLog.Println("Email sent")
	*/

	return nil
}
