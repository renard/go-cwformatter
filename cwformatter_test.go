package cwformatter

import (
	"bytes"
	"os"
	"testing"

	"github.com/logrusorgru/aurora"
	log "github.com/sirupsen/logrus"
)

type runFunc func(...interface{})

var l = log.New()

func initLogger() *log.Logger {
	f := NewFormatter()
	f.Timeformat = ""
	f.FieldsColumn = 30
	l.Formatter = f

	return l
}

func doLog(l *log.Logger, fct runFunc, input string) string {
	b := &bytes.Buffer{}
	l.Out = b

	fct(input)

	bs := b.String()
	lbs := len(bs)
	if lbs > 0 {
		lbs--
	}
	return bs[:lbs]
}

func TestLevels(t *testing.T) {
	l := initLogger()
	tests := []struct {
		n    string
		f    runFunc
		in   string
		want string
	}{
		{"Trace", l.Trace, "Message", ""},
		{"Debug", l.Debug, "Message", ""},
		{"Info", l.Info, "Message", "Message"},
		{"Warn", l.Warn, "Message", "Message"},
		{"Error", l.Error, "Message", "Message"},
		//{"Fatal", l.Fatal, "Message", "Message"},
		//{"Panic", l.Panic, "Message", "Message"},

		{"Error with fields", l.WithFields(log.Fields{"f1": "v1"}).Error,
			"Some log", `Some log                      | f1="v1"`},

		{"Start command", l.WithFields(log.Fields{"COMMAND_START": "ls -al /"}).Info,
			"", `Running ls -al /`},

		{"Success command", l.WithFields(log.Fields{"COMMAND_RESULT": 0}).Info,
			"", ` ==> OK`},
		{"Failed command", l.WithFields(log.Fields{"COMMAND_RESULT": 1}).Error,
			"", ` ==> Failed (exit code 1)`},
	}
	for _, test := range tests {
		t.Run(test.n, func(t *testing.T) {
			got := doLog(l, test.f, test.in)
			if got != test.want {
				t.Errorf("got %q want %q", got, test.want)
			}
		})
	}
}

func TestColors(t *testing.T) {
	f := NewFormatter()
	tests := []struct {
		n string
		l log.Level
		c aurora.Color
	}{
		{"Panic", log.PanicLevel, f.PanicColor},
		{"Fatal", log.FatalLevel, f.FatalColor},
		{"Error", log.ErrorLevel, f.ErrorColor},
		{"Warn", log.WarnLevel, f.WarnColor},
		{"Info", log.InfoLevel, f.InfoColor},
		{"Debug", log.DebugLevel, f.DebugColor},
		{"Trace", log.TraceLevel, f.TraceColor},
	}
	for _, test := range tests {
		t.Run(test.n, func(t *testing.T) {
			got := f.color(test.l)
			if got != test.c {
				t.Errorf("got %q want %q", got, test.c)
			}
		})
	}
}

func TestTerminal(t *testing.T) {
	f := NewFormatter()
	f.FieldsColumn = 30
	l.Out = os.Stderr
	l.Formatter = f
	l.Warn("Foo")
}

func TestHooks(t *testing.T) {
	f := NewFormatter()
	f.Timeformat = ""
	f.FieldsColumn = 30
	l.Formatter = f

	tests := []struct {
		n    string
		f    runFunc
		in   string
		want string
		pre  func()
	}{
		{"Without COMMAND_START", l.WithFields(log.Fields{"COMMAND_START": "ls"}).Info,
			"Message", `Message                       | COMMAND_START="ls"`,
			func() { f.DeleteHook("COMMAND_START") }},

		{"Without COMMAND_START", l.WithFields(log.Fields{"COMMAND_START": "ls"}).Info,
			"", `COMMAND_START="ls"`,
			func() { f.DeleteHook("COMMAND_START") }},

		{"With COMMAND_START", l.WithFields(log.Fields{"COMMAND_START": "ls"}).Info,
			"Message", `Message                       | Running ls`,
			func() { f.AddHook("COMMAND_START", commandStart) }},

		{"With COMMAND_START", l.WithFields(log.Fields{"COMMAND_START": "ls"}).Info,
			"", `Running ls`,
			func() { f.AddHook("COMMAND_START", commandStart) }},
	}

	for _, test := range tests {
		t.Run(test.n, func(t *testing.T) {
			test.pre()
			got := doLog(l, test.f, test.in)
			if got != test.want {
				t.Errorf("got %q want %q", got, test.want)
			}
		})
	}
}

func BenchmarkLevels(b *testing.B) {
	l := initLogger()
	tests := []struct {
		n string
		f runFunc
	}{
		{"Trace", l.Trace},
		{"Error", l.Error},
		{"Fields Trace", l.WithFields(log.Fields{"f1": "v1", "f2": "v2"}).Trace},
		{"Fields Error", l.WithFields(log.Fields{"f1": "v1", "f2": "v2"}).Error},
		{"CMD_START", l.WithFields(log.Fields{"COMMAND_START": "ls"}).Info},
		{"CMD_RES OK", l.WithFields(log.Fields{"COMMAND_RESULT": 0}).Info},
		{"CMD_RES ERR", l.WithFields(log.Fields{"COMMAND_RESULT": 1}).Error},
	}
	for _, test := range tests {
		b.Run(test.n, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				doLog(l, test.f, "Message")
			}
		})
	}
}
