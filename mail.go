package mail

import (
	"bufio"
	"fmt"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/zxdev/unit"
)

// Mail mailer struct
type Mail struct {
	To               []string
	Subject          string
	Message          string
	user, pass, smtp string // unit file settings
	unit             string // path to unit file
	alert            bool   // alert flag
}

// NewMailer *Mail configurator
func NewMailer(user, pass, smtp string) *Mail {
	return &Mail{
		user: user,
		pass: pass,
		smtp: smtp,
	}
}

// NewMail *Mail unit based configurator and stores
// the unit file for futrue use; requires path,section ordering
func NewMail(path string, a ...string) *Mail {

	var mail = new(Mail)
	var u unit.Unit

	mail.unit = path // store unit file path
	u.Parse(path, a...)

	mail.user = u["user"]
	mail.pass = u["pass"]
	mail.smtp = u["smtp"]

	return mail
}

// reset mail
func (mail *Mail) reset() {

	mail.alert = false
	mail.To = []string{}
	mail.Subject = ""
	mail.Message = ""

}

// Alert sets the alert boolean flag
func (mail *Mail) Alert() *Mail { mail.alert = true; return mail }

// NewLine character
func (mail *Mail) NewLine() string { return "\r\n" }

// FromUnit populates mail.To from a unit file
//
// [section]
//
// mail=address,address
func (mail *Mail) FromUnit(path *string, a ...string) *Mail {

	if len(a) > 0 {

		var u unit.Unit
		if path == nil {
			u.Parse(mail.unit, a...)
		} else {
			u.Parse(*path, a...)
		}

		mail.To = strings.Split(u["mail"], ",")
	}

	return mail
}

// FromFile populates mail.To from \n delimited file
func (mail *Mail) FromFile(path string) *Mail {

	f, err := os.Open(path)
	if err == nil {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			mail.To = append(mail.To, scanner.Text())
		}
	}

	return mail
}

// Send with auto reset
//
// To:      @email, FromFile, unit [section]
//
// Subject: nil, string and/or use .Alert()
//
// Message: string or []string for multiple lines
func (mail *Mail) Send(to, subject, message interface{}) bool {

	if len(mail.pass) == 0 {
		return false
	}

	defer mail.reset()

	// configure mail.To
	switch t := to.(type) {
	case nil: // maybe set already externally
	case []string: // proper format
	case string: // various support options
		switch {
		case strings.Contains(t, "@"): // email address
			mail.To = strings.Split(t, ",")
		case strings.Contains(t, string(filepath.Separator)): // file load
			mail.FromFile(t)
		default:
			mail.FromUnit(&mail.unit, t) // unit file section
		}
	}
	if len(mail.To) == 0 {
		return false
	}
	for i := range mail.To {
		mail.To[i] = strings.TrimSpace(mail.To[i])
	}

	// configure mail.Subject
	switch sbj := subject.(type) {
	case string:
		mail.Subject = sbj
	default:
		name, _ := os.Hostname()
		mail.Subject = fmt.Sprintf("%s: message", name)
	}
	if mail.alert {
		mail.Subject = "ALERT: " + mail.Subject
	}

	// configure mail.Message
	switch msg := message.(type) {
	case string:
		mail.Message = msg
	case []string:
		mail.Message = strings.Join(msg, mail.NewLine())
	default:
		if !mail.alert {
			mail.Message = time.Now().UTC().Format(time.RFC3339)[:19]
		}
	}

	return smtp.SendMail(
		mail.smtp+":587", smtp.PlainAuth("", mail.user, mail.pass, mail.smtp), mail.user, mail.To,
		[]byte(fmt.Sprintf("To: %s%sSubject:%s %s%s%s%s",
			strings.Join(mail.To, ","), mail.NewLine(),
			mail.Subject, time.Now().UTC().Format(time.RFC3339)[:19], mail.NewLine(),
			mail.Message, mail.NewLine())),
	) == nil

}
