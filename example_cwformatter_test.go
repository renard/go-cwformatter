// +build example

package cwformatter

// Simple test to display an example on the screen
//
// Run this file using: go test -tags example
func init() {
	Example_invoke()
}

// Example on how to invoke CWFormatter
func Example_invoke() {
	l := log.New()
	l.Out = os.Stdout
	f := NewFormatter()
	l.Formatter = f

	l.SetLevel(log.TraceLevel)
	// l.SetFormatter(new(ColorFormatter))
	l.Trace("Trace: Something very low level.")
	l.Debug("Debug: Useful debugging information.")
	l.Info("Info: Something noteworthy happened!")
	l.Warn("Warn: You should probably take a look at this.")
	l.Error("Error: Something failed but I'm not quitting.")

	l.WithFields(log.Fields{
		"event": "event",
		"topic": "topic",
	}).Trace("Example with fields")

	l.Warning("Commands have to be run by caller.")

	l.Debug("Example of a bogus command with COMMAND_START/COMMAD_RESULT.")
	l.WithFields(log.Fields{
		"COMMAND_START": "ls -al /bogus",
	}).Info("")
	l.WithFields(log.Fields{
		"COMMAND_RESULT": 2,
	}).Error("")

	l.Debug("Example of a successful command.")
	l.WithFields(log.Fields{
		"COMMAND_START": "ls -al /",
	}).Info("")
	l.WithFields(log.Fields{
		"COMMAND_RESULT": 0,
	}).Info("")

}
