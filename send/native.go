package send

import (
	"fmt"
	"log"
	"os"

	"github.com/tychoish/grip/level"
	"github.com/tychoish/grip/message"
)

type nativeLogger struct {
	logger *log.Logger
	*base
}

// NewFileLogger creates a Sender implementation that writes log
// output to a file. Returns an error but falls back to a standard
// output logger if there's problems with the file. Internally using
// the go standard library logging system.
func NewFileLogger(name, filePath string, l LevelInfo) (Sender, error) {
	s, err := MakeFileLogger(filePath)
	if err != nil {
		return nil, err
	}

	return setup(s, name, l)
}

// MakeFileLogger creates a file-based logger, writing output to
// the specified file. The Sender instance is not configured: Pass to
// Journaler.SetSender or call SetName before using.
func MakeFileLogger(filePath string) (Sender, error) {
	s := &nativeLogger{base: newBase("")}

	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("error opening logging file, %s", err.Error())
	}

	s.reset = func() {
		prefix := fmt.Sprintf("[%s]", s.Name())
		s.logger = log.New(f, prefix, log.LstdFlags)
	}

	s.closer = func() error {
		return f.Close()
	}

	return s, nil
}

// NewNativeLogger creates a new Sender interface that writes all
// loggable messages to a standard output logger that uses Go's
// standard library logging system.
func NewNativeLogger(name string, l LevelInfo) (Sender, error) {
	return setup(MakeNative(), name, l)
}

// MakeNative returns an unconfigured native standard-out logger. You
// *must* call SetName on this instance before using it. (Journaler's
// SetSender will typically do this.)
func MakeNative() Sender {
	s := &nativeLogger{
		base: newBase(""),
	}
	s.level = LevelInfo{level.Trace, level.Trace}

	s.reset = func() {
		prefix := fmt.Sprintf("[%s]", s.Name())
		s.logger = log.New(os.Stdout, prefix, log.LstdFlags)
	}

	// we don't call reset here because name isn't set yet, and
	// SetName/SetSender will always call it. The potential for a nil
	// pointer is not 0

	return s
}

func MakeErrorLogger() Sender {
	s := &nativeLogger{
		base: newBase(""),
	}
	s.level = LevelInfo{level.Trace, level.Trace}

	s.reset = func() {
		prefix := fmt.Sprintf("[%s]", s.Name())
		s.logger = log.New(os.Stderr, prefix, log.LstdFlags)
	}

	return s
}

func NewErrorLogger(name string, l LevelInfo) (Sender, error) {
	return setup(MakeErrorLogger(), name, l)
}

func (s *nativeLogger) Type() SenderType { return Native }
func (s *nativeLogger) Send(m message.Composer) {
	if s.level.ShouldLog(m) {
		s.logger.Printf("[p=%s]: %s", m.Priority(), m.Resolve())
	}
}
