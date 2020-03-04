package cwformatter

import (
	"bytes"
	"testing"

	log "github.com/sirupsen/logrus"
)

type runFunc func(...interface{})

var l = log.New()

func runTest(t *testing.T, name string, fct runFunc, input, expect string) {
	b := &bytes.Buffer{}
	l.Out = b
	f := NewFormatter()
	f.Timeformat = ""
	f.FieldsColumn = 30
	l.Formatter = f

	fct(input)

	bs := b.String()
	lbs := len(bs)
	if lbs > 0 {
		lbs--
	}
	if bs[:lbs] != expect {
		t.Errorf("Test %s failed: -%s- instead of -%s-", name, bs[:lbs], expect)
	}
}

func TestSimpleLog(t *testing.T) {

	runTest(t, "Simple Error", l.Error, "Some log", "Some log")
	runTest(t, "Simple Warn", l.Warn, "Some log", "Some log")
	runTest(t, "Simple Info", l.Info, "Some log", "Some log")
	runTest(t, "Simple Debug", l.Debug, "Some log", "")
	runTest(t, "Simple Trace", l.Trace, "Some log", "")

	runTest(t, "Error with fields", l.WithFields(log.Fields{
		"f1": "v1",
		//		"f1": "v1",
	}).Error, "Some log",
		`Some log                      | f1="v1"`)

	// runTest(t, "Start command and field", l.WithFields(log.Fields{
	// 	"COMMAND_START": "ls -al /inexistant",
	// 	"event":         "event",
	// }).Info, "", `Running ls -al /inexistant event="event"`)

	runTest(t, "Start command", l.WithFields(log.Fields{
		"COMMAND_START": "ls -al /",
	}).Info, "", `Running ls -al /`)

	runTest(t, "Success command", l.WithFields(log.Fields{
		"COMMAND_RESULT": 0,
	}).Info, "", ` ==> OK`)

	runTest(t, "Failed command", l.WithFields(log.Fields{
		"COMMAND_RESULT": 1,
	}).Info, "", ` ==> Failed (exit code 1)`)

}

func TestHooks(t *testing.T) {
	b := &bytes.Buffer{}
	l.Out = b
	f := NewFormatter()
	f.Timeformat = ""
	f.FieldsColumn = 30
	l.Formatter = f

	f.DeleteHook("COMMAND_START")
	l.WithFields(log.Fields{"COMMAND_START": "ls"}).Info("Message")
	bs := b.String()
	lbs := len(bs)
	if lbs > 0 {
		lbs--
	}
	expect := `Message                       | COMMAND_START="ls"`
	if bs[:lbs] != expect {
		t.Errorf("Test failed: -%s- instead of -%s-", bs[:lbs], expect)
	}
	b.Reset()

	f.AddHook("COMMAND_START", commandStart)
	l.WithFields(log.Fields{"COMMAND_START": "ls"}).Info("Message")
	bs = b.String()
	lbs = len(bs)
	if lbs > 0 {
		lbs--
	}
	expect = "Message                       | Running ls"
	if bs[:lbs] != expect {
		t.Errorf("Test failed: -%s- instead of -%s-", bs[:lbs], expect)
	}

}

func BenchmarkSimple(b *testing.B) {
	// b := &bytes.Buffer{}
	// l.Out = b
	f := NewFormatter()
	f.Timeformat = ""
	f.FieldsColumn = 30
	l.Formatter = f
	for i := 0; i < b.N; i++ {
		l.Info("Message")
	}
}

func BenchmarkFields(b *testing.B) {
	// b := &bytes.Buffer{}
	// l.Out = b
	f := NewFormatter()
	f.Timeformat = ""
	f.FieldsColumn = 30
	l.Formatter = f
	for i := 0; i < b.N; i++ {
		l.WithFields(log.Fields{
			"f1": "v1",
			"f2": "v2",
		}).Info("Message")
	}
}

func BenchmarkCommandStart(b *testing.B) {
	// b := &bytes.Buffer{}
	// l.Out = b
	f := NewFormatter()
	f.Timeformat = ""
	f.FieldsColumn = 30
	l.Formatter = f
	for i := 0; i < b.N; i++ {
		l.WithFields(log.Fields{
			"COMMAND_START": "ls",
		}).Info("Message")
	}
}

func BenchmarkCommandResultOK(b *testing.B) {
	// b := &bytes.Buffer{}
	// l.Out = b
	f := NewFormatter()
	f.Timeformat = ""
	f.FieldsColumn = 30
	l.Formatter = f
	for i := 0; i < b.N; i++ {
		l.WithFields(log.Fields{
			"COMMAND_RESULT": 0,
		}).Info("Message")
	}
}

func BenchmarkCommandResultFail(b *testing.B) {
	// b := &bytes.Buffer{}
	// l.Out = b
	f := NewFormatter()
	f.Timeformat = ""
	f.FieldsColumn = 30
	l.Formatter = f
	for i := 0; i < b.N; i++ {
		l.WithFields(log.Fields{
			"COMMAND_RESULT": 1,
		}).Info("Message")
	}
}
