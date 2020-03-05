// Copyright © 2020 Sébastien Gross
//
// Created: 2020-02-25
// Last changed: 2020-03-05 00:26:05
//
// This program is free software. It comes without any warranty, to
// the extent permitted by applicable law. You can redistribute it
// and/or modify it under the terms of the Do What The Fuck You Want
// To Public License, Version 2, as published by Sam Hocevar. See
// http://sam.zoy.org/wtfpl/COPYING for more details.

// Package cwformatter is a simple lorus formatter to display color logs (if
// possible, no color is used if output is piped) to the terminal.
package cwformatter

import (
	"bytes"
	"fmt"
	"os"
	"sync"

	"github.com/logrusorgru/aurora"
	"github.com/mattn/go-isatty"
	"github.com/sirupsen/logrus"
)

// CWFormatter holds the formatter setup.
type CWFormatter struct {
	// time.Format layout suitable string.
	Timeformat string
	// Column used to start fields display if message is not too long.
	FieldsColumn int
	ShowLevel    bool
	UseColor     bool
	// All colors define how to display logs depending on the message
	// level. TimeColor stands for the timestamp. Both KeyColor and
	// ValueColor are used when displaying fields. Command*Color are used
	// when logging command action using special COMMAND_START and
	// COMMAND_RESULT special fields.
	TimeColor          aurora.Color
	PanicColor         aurora.Color
	FatalColor         aurora.Color
	ErrorColor         aurora.Color
	WarnColor          aurora.Color
	InfoColor          aurora.Color
	DebugColor         aurora.Color
	TraceColor         aurora.Color
	KeyColor           aurora.Color
	ValueColor         aurora.Color
	CommandHeaderColor aurora.Color
	CommandColor       aurora.Color
	CommandSucessColor aurora.Color
	CommandFailColor   aurora.Color
	// fieldsHooks is a string map of functions triggered by a special
	// field.
	fieldsHooks map[string]fieldHookFunc
	mu          sync.Mutex
}

// fieldHookFunc defines a function signature suitable to execute a hook
// when a specific CWFormatter field is found.
//
// First argument a pointer to the current CWFormatter type to allow access
// to colors.
//
// Second argument is the a bytes.Buffer pointer to the buffer to which the
// message should be written to.
//
// Third argument is the specific fied value defined as an interface. It's
// up to the function to handle it as needed.
//
// The fourth argument is the aurora.Aurora type which can be used to
// colorize the message. The aurora type has been computed in the Format
// function to handle correctly if the message should be color-printed or
// not to allow a safe use of Colorize method.
type fieldHookFunc func(*CWFormatter, *bytes.Buffer, interface{}, aurora.Aurora)

func commandStart(f *CWFormatter, b *bytes.Buffer, msg interface{}, au aurora.Aurora) {
	fmt.Fprint(b, au.Colorize("Running", f.CommandHeaderColor))
	b.WriteByte(' ')
	fmt.Fprint(b, au.Colorize(msg, f.CommandColor))
}

func commandResult(f *CWFormatter, b *bytes.Buffer, msg interface{}, au aurora.Aurora) {
	b.WriteByte(' ')
	fmt.Fprint(b, au.Colorize("==>", f.CommandHeaderColor))
	b.WriteByte(' ')
	if msg == 0 {
		fmt.Fprint(b, au.Colorize("OK", f.CommandSucessColor))
	} else {
		fmt.Fprint(b, au.Colorize("Failed", f.CommandFailColor))
		b.WriteByte(' ')
		fmt.Fprint(b, au.Colorize(
			fmt.Sprintf("(exit code %d)", msg), f.CommandHeaderColor))
	}
}

// NewFormatter returns a new CWFormatter with default setup. All fields can
// be updated afterwards.
func NewFormatter() (f *CWFormatter) {
	f = &CWFormatter{
		Timeformat:         "2006-01-02 15:04:05",
		TimeColor:          aurora.Gray(15, nil).Color(),
		FieldsColumn:       70,
		UseColor:           true,
		PanicColor:         aurora.RedFg | aurora.BoldFm,
		FatalColor:         aurora.RedFg | aurora.BoldFm,
		ErrorColor:         aurora.RedFg,
		WarnColor:          aurora.YellowFg | aurora.BrightFg | aurora.BoldFm,
		InfoColor:          aurora.CyanFg | aurora.BrightFg,
		DebugColor:         aurora.MagentaFg,
		TraceColor:         aurora.Gray(15, nil).Color(),
		KeyColor:           aurora.Gray(15, nil).Color(),
		ValueColor:         aurora.Gray(19, nil).Color(),
		CommandHeaderColor: aurora.Gray(10, nil).Color(),
		CommandColor:       aurora.Gray(15, nil).Color(),
		CommandSucessColor: aurora.GreenFg | aurora.BoldFm,
		CommandFailColor:   aurora.RedFg | aurora.BoldFm,
		fieldsHooks:        map[string]fieldHookFunc{},
	}
	f.AddHook("COMMAND_START", commandStart)
	f.AddHook("COMMAND_RESULT", commandResult)
	return
}

// Format is the function used by logrus to format a message. See
// logrus.Format for details.
//
// The message is displayed using colors unless the terminal current output
// does not support it (such as io.Writer or a piped output) or if UseColor
// is set to false. A new line is added to each message. So do not format
// yours with a new line but with a full stop.
//
// Fields are displayed at the end of the message or at FieldsColumn column
// depending on which comes first. Note that fields are not sorted and may
// be displayed in a random order.
//
// There are 2 special field items COMMAND_START and COMMAND_RESULT that can
// be used to log exected command (and their results). These special fields
// should not be used with other ones.
//
//
//    // How to log a success command
//    l.WithFields(logrus.Fields{"COMMAND_START": "ls -al /"}).Info, "")
//    l.WithFields(logrus.Fields{"COMMAND_RESULT": 0}).Info,"")
//    // Result:
//    //   Running ls -al /
//    //    ==> OK
//
//    // How to log a failed command
//    l.WithFields(logrus.Fields{"COMMAND_START": "ls -al /inexistant"}).Info, "")
//    l.WithFields(logrus.Fields{"COMMAND_RESULT": 2}).Error,"")
//    // Result:
//    //   Running ls -al /inexistant
//    //    ==> Failed (exit code 2)
//
// Other fields that trigger a specific action can be added or removed using
// AddHook or DeleteHook function.
func (f *CWFormatter) Format(entry *logrus.Entry) (b []byte, err error) {

	isTerm := false
	if file, ok := (entry.Logger.Out).(*os.File); ok {
		isTerm = isatty.IsTerminal(file.Fd())
	}

	// Line length
	l := 0

	au := aurora.NewAurora(f.UseColor && isTerm)

	buf := &bytes.Buffer{}

	fmt.Fprintf(buf, "%s",
		au.Colorize(entry.Time.Format(f.Timeformat), f.TimeColor))
	l += len(f.Timeformat)
	if l > 0 {
		buf.WriteByte(' ')
		l++
	}

	// Separate function to reduce gocyclo complexity
	color := f.color(entry.Level)

	fmt.Fprintf(buf, "%s", au.Colorize(entry.Message, color))
	l += len(entry.Message)

	i := 0
	for k, v := range entry.Data {
		if i == 0 && l < f.FieldsColumn && entry.Message != "" {
			buf.Write(bytes.Repeat([]byte{' '}, f.FieldsColumn-l))
			buf.WriteByte('|')
		}
		if i > 0 || entry.Message != "" {
			buf.WriteByte(' ')
		}

		if fct := f.fieldsHooks[k]; fct != nil {
			fct(f, buf, v, au)
			//break
		} else {
			fmt.Fprintf(buf, "%s=%#v",
				au.Colorize(k, f.KeyColor),
				au.Colorize(v, f.ValueColor))
		}
		i++
	}
	//	buf.WriteByte('-')
	buf.WriteByte('\n')
	b = buf.Bytes()

	return
}

// color returns the defined color for loglevel l.
func (f *CWFormatter) color(l logrus.Level) (color aurora.Color) {
	switch l {
	case logrus.PanicLevel:
		color = f.PanicColor
	case logrus.FatalLevel:
		color = f.FatalColor
	case logrus.ErrorLevel:
		color = f.ErrorColor
	case logrus.WarnLevel:
		color = f.WarnColor
	case logrus.InfoLevel:
		color = f.InfoColor
	case logrus.DebugLevel:
		color = f.DebugColor
	case logrus.TraceLevel:
		color = f.TraceColor
	}
	return
}

// AddHook adds or replaces fct function to the fieldsHooks map for given
// field. One an only one function can be defined for one field.
func (f *CWFormatter) AddHook(field string, fct fieldHookFunc) {
	f.mu.Lock()
	f.fieldsHooks[field] = fct
	f.mu.Unlock()
}

// DeleteHook deletes hook function for given field. If field was not set,
// nothing happens.
func (f *CWFormatter) DeleteHook(field string) {
	_, ok := f.fieldsHooks[field]
	if ok {
		f.mu.Lock()
		delete(f.fieldsHooks, field)
		f.mu.Unlock()
	}
}
