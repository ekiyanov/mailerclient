package mailerclient

import (
	"html/template"
	"io"
	"log"
	"os"
	"strconv"
	"sync"

	gomail "gopkg.in/mail.v2"
)

var _mc *MailClient
var _mu sync.Mutex

type MailClient struct {
	from      string
	templates map[string]*template.Template
	dialer    *gomail.Dialer
}

func NewMailClient() *MailClient {
	username := os.Getenv("MAIL_USERNAME")
	password := os.Getenv("MAIL_PASSWORD")
	host := os.Getenv("MAIL_HOST")
	portS := os.Getenv("MAIL_PORT")
	port, _ := strconv.ParseInt(portS, 10, 64)
	if port == 0 {
		port = 587
	}

	dialer := gomail.NewDialer(host, int(port), username, password)

	return &MailClient{
		dialer: dialer,
	}
}

func SharedMailClient() *MailClient {
	_mu.Lock()
	defer _mu.Unlock()

	if _mc == nil {
		_mc = NewMailClient()
	}

	return _mc
}

func Send(to, subject, template string, templateContext interface{}) {
	SharedMailClient().Send(to, subject, template, templateContext)
}

func (mc *MailClient) Send(to, subject string, templateName string, templateContext interface{}) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("Recovered:", err)
		}
	}()

	if mc.from == "" {
		mc.from = os.Getenv("MAIL_FROM")
		if mc.from == "" {
			panic("MAIL_FROM is not set")
		}
	}

	message := gomail.NewMessage()
	message.SetHeader("From", mc.from)
	message.SetHeader("To", to)
	message.SetHeader("Subject", subject)

	if mc.templates == nil {
		mc.templates = map[string]*template.Template{}
	}
	tmpl := mc.templates[templateName]

	if tmpl == nil {
		tmpl = template.Must(template.ParseFiles("/templates/" + templateName))
		mc.templates[templateName] = tmpl
	}

	message.SetBodyWriter("text/html", func(w io.Writer) error {
		return tmpl.Execute(w, templateContext)
	})

	if err := mc.dialer.DialAndSend(message); err != nil {
		log.Println(err)
	}
}
