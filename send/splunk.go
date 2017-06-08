package send

import (
	"fmt"
	"os"

	"github.com/fuyufjh/splunk-hec-go"
	"github.com/mongodb/grip/level"
	"github.com/mongodb/grip/message"
)

const (
	splunkServerURL   = "GRIP_SPLUNK_SERVER_URL"
	splunkClientToken = "GRIP_SPLUNK_CLIENT_TOKEN"
)

type splunkLogger struct {
	info   SplunkConnectionInfo
	client splunkClient
	*Base
}

// SplunkConnectionInfo stores all information needed to connect
// to a splunk server to send log messsages.
type SplunkConnectionInfo struct {
	ServerURL string
	Token     string
}

// GetSplunkConnectionInfo builds a SplunkConnectionInfo structure
// reading default values from the following environment variables:
//
// 		GRIP_SPLUNK_SERVER_URL
//		GRIP_SPLUNK_CLIENT_TOKEN
func GetSplunkConnectionInfo() SplunkConnectionInfo {
	return SplunkConnectionInfo{
		ServerURL: os.Getenv(splunkServerURL),
		Token:     os.Getenv(splunkClientToken),
	}
}

func (s *splunkLogger) Send(m message.Composer) {
	g, isGroup := m.(*message.GroupComposer)
	if isGroup {
		batch := make([]*hec.Event, 0)
		for _, c := range g.Messages() {
			if s.level.ShouldLog(c) {
				batch = append(batch, hec.NewEvent(c.Raw()))
			}
		}
		if err := s.client.WriteBatch(batch); err != nil {
			s.errHandler(err, m)
		}
	} else {
		if s.level.ShouldLog(m) {
			event := hec.NewEvent(m.Raw())
			if err := s.client.WriteEvent(event); err != nil {
				s.errHandler(err, m)
			}
		}
	}
}

func newSplunkLoggerNoClient(name string, info SplunkConnectionInfo, l LevelInfo, client splunkClient) (Sender, error) {
	s := &splunkLogger{
		info:   info,
		client: client,
		Base:   NewBase(name),
	}

	if err := s.client.Create(info.ServerURL, info.Token); err != nil {
		return nil, err
	}

	if err := s.SetLevel(l); err != nil {
		return nil, err
	}

	return s, nil
}

// NewSplunkLogger constructs a new Sender implementation that sends
// messages to a Splunk event collector using the credentials specified
// in the SplunkConnectionInfo struct.
func NewSplunkLogger(name string, info SplunkConnectionInfo, l LevelInfo) (Sender, error) {
	return newSplunkLoggerNoClient(name, info, l, &splunkClientImpl{})
}

// MakeSplunkLogger constructs a new Sender implementation that reads
// the hostname, username, and password from environment variables:
//
// 		GRIP_SPLUNK_SERVER_URL
//		GRIP_SPLUNK_CLIENT_TOKEN
func MakeSplunkLogger(name string) (Sender, error) {
	info := GetSplunkConnectionInfo()
	if info.ServerURL == "" {
		return nil, fmt.Errorf("environment variable %s not defined, cannot create splunk client",
			info.ServerURL)
	}
	if info.Token == "" {
		return nil, fmt.Errorf("environment variable %s not defined, cannot create slack client",
			info.Token)
	}
	return NewSplunkLogger(name, info, LevelInfo{level.Trace, level.Trace})
}

////////////////////////////////////////////////////////////////////////
//
// interface wrapper for the splunk client so that we can mock things out
//
////////////////////////////////////////////////////////////////////////

type splunkClient interface {
	Create(string, string) error
	WriteEvent(*hec.Event) error
	WriteBatch([]*hec.Event) error
}

type splunkClientImpl struct {
	hec.HEC
}

func (c *splunkClientImpl) Create(serverURL string, token string) error {
	c.HEC = hec.NewClient(serverURL, token)
	return nil
}
