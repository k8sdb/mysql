package mail

import (
	"net/mail"
	"os"
	"testing"
)

const (
	TMP = "/tmp/mail.jpg"
)

var (
	subject = Subject("中山大学")
	header  = Header{
		From:    mail.Address{"mikespook", "mikespook#gmail.com"},
		To:      []mail.Address{{"X", "x#nowhere"}, {"Y", "y#nowhere"}},
		Cc:      []mail.Address{{"Z", "z#nowhere"}},
		Subject: subject,
	}
	body = &Body{
		Charset:     "GBK",
		ContentType: "text/html",
		Content:     []byte("<strong>Hello</strong>"),
	}
	attachment = FileAttach(TMP)
)

func TestSubject(t *testing.T) {
	if subject.String() != "=?UTF-8?B?5Lit5bGx5aSn5a2m?=" {
		t.Error("Subject encoding faild")
	}
}

func TestHeader(t *testing.T) {
	list := header.ToList()
	if len(list) != 3 {
		t.Error("The length of To list should be 3")
	}
	t.Logf("%s", header.Bytes())
}

func TestMail(t *testing.T) {
	if err := create(TMP); err != nil {
		t.Log(err)
	}
	defer remove(TMP)
	m := NewMail()
	m.Header = header
	m.Append(body)
	m.Append(attachment)
	if b, err := m.Bytes(); err != nil {
		t.Log(err)
	} else {
		t.Logf("%s", b)
	}
}

func create(name string) error {
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	defer f.Close()
	f.WriteString("Hello world")
	return nil
}

func remove(name string) {
	os.Remove(name)
}
