package send

import (
	"os"
	"testing"

	"github.com/mongodb/grip/level"
	"github.com/mongodb/grip/message"
	"github.com/stretchr/testify/suite"
)

type SplunkSuite struct {
	info   SplunkConnectionInfo
	client splunkClient
	suite.Suite
}

func TestSplunkSuite(t *testing.T) {
	suite.Run(t, new(SplunkSuite))
}

func (s *SplunkSuite) SetupSuite() {}

func (s *SplunkSuite) SetupTest() {
	s.info = SplunkConnectionInfo{}
	s.client = &splunkClientMock{}
}

func (s *SplunkSuite) TestEnvironmentVariableReader() {
	serverVal := "serverURL"
	tokenVal := "token"

	defer os.Setenv(splunkServerURL, os.Getenv(splunkServerURL))
	defer os.Setenv(splunkClientToken, os.Getenv(splunkClientToken))

	s.NoError(os.Setenv(splunkServerURL, serverVal))
	s.NoError(os.Setenv(splunkClientToken, tokenVal))

	info := GetSplunkConnectionInfo()

	s.Equal(serverVal, info.ServerURL)
	s.Equal(tokenVal, info.Token)
}

func (s *SplunkSuite) TestNewConstructor() {
	sender, err := NewSplunkLogger("name", s.info, LevelInfo{level.Debug, level.Info})
	s.NoError(err)
	s.NotNil(sender)
}

func (s *SplunkSuite) TestNewConstructorFailsWhenClientCreateFails() {
	s.client = &splunkClientMock{failCreate: true}

	sender, err := newSplunkLoggerNoClient("name", s.info, LevelInfo{level.Debug, level.Info}, s.client)
	s.Error(err)
	s.Nil(sender)

	s.client = &splunkClientMock{}
}

func (s *SplunkSuite) TestAutoConstructor() {
	serverVal := "serverURL"
	tokenVal := "token"

	defer os.Setenv(splunkServerURL, os.Getenv(splunkServerURL))
	defer os.Setenv(splunkClientToken, os.Getenv(splunkClientToken))

	s.NoError(os.Setenv(splunkServerURL, serverVal))
	s.NoError(os.Setenv(splunkClientToken, tokenVal))

	sender, err := MakeSplunkLogger("name")
	s.NoError(err)
	s.NotNil(sender)
}

func (s *SplunkSuite) TestAutoConstructorFailsWhenEnvVarFails() {
	serverVal := ""
	tokenVal := ""

	defer os.Setenv(splunkServerURL, os.Getenv(splunkServerURL))
	defer os.Setenv(splunkClientToken, os.Getenv(splunkClientToken))

	s.NoError(os.Setenv(splunkServerURL, serverVal))
	s.NoError(os.Setenv(splunkClientToken, tokenVal))

	sender, err := MakeSplunkLogger("name")
	s.Error(err)
	s.Nil(sender)

	serverVal = "serverVal"

	s.NoError(os.Setenv(splunkServerURL, serverVal))
	sender, err = MakeSplunkLogger("name")
	s.Error(err)
	s.Nil(sender)
}

func (s *SplunkSuite) TestSendMethod() {
	sender, err := newSplunkLoggerNoClient("name", s.info, LevelInfo{level.Debug, level.Info}, s.client)
	s.NoError(err)
	s.NotNil(sender)

	mock, ok := s.client.(*splunkClientMock)
	s.True(ok)
	s.Equal(mock.numSent, 0)

	m := message.NewDefaultMessage(level.Debug, "hello")
	sender.Send(m)
	s.Equal(mock.numSent, 0)

	m = message.NewDefaultMessage(level.Alert, "")
	sender.Send(m)
	s.Equal(mock.numSent, 0)

	m = message.NewDefaultMessage(level.Alert, "world")
	sender.Send(m)
	s.Equal(mock.numSent, 1)
}

func (s *SplunkSuite) TestSendMethodWithError() {
	sender, err := newSplunkLoggerNoClient("name", s.info, LevelInfo{level.Debug, level.Info}, s.client)
	s.NoError(err)
	s.NotNil(sender)

	mock, ok := s.client.(*splunkClientMock)
	s.True(ok)
	s.Equal(mock.numSent, 0)
	s.False(mock.failSend)

	m := message.NewDefaultMessage(level.Alert, "world")
	sender.Send(m)
	s.Equal(mock.numSent, 1)

	mock.failSend = true
	sender.Send(m)
	s.Equal(mock.numSent, 1)
}

func (s *SplunkSuite) TestBatchSendMethod() {
	sender, err := newSplunkLoggerNoClient("namne", s.info, LevelInfo{level.Debug, level.Info}, s.client)
	s.NoError(err)
	s.NotNil(sender)

	mock, ok := s.client.(*splunkClientMock)
	s.True(ok)
	s.Equal(mock.numSent, 0)

	m1 := message.NewDefaultMessage(level.Alert, "hello")
	m2 := message.NewDefaultMessage(level.Debug, "hello")
	m3 := message.NewDefaultMessage(level.Alert, "")
	m4 := message.NewDefaultMessage(level.Alert, "hello")

	g := message.MakeGroupComposer(m1, m2, m3, m4)

	sender.Send(g)
	s.Equal(mock.numSent, 2)
}

func (s *SplunkSuite) TestBatchSendMethodWithEror() {
	sender, err := newSplunkLoggerNoClient("name", s.info, LevelInfo{level.Debug, level.Info}, s.client)
	s.NoError(err)
	s.NotNil(sender)

	mock, ok := s.client.(*splunkClientMock)
	s.True(ok)
	s.Equal(mock.numSent, 0)
	s.False(mock.failSend)

	m1 := message.NewDefaultMessage(level.Alert, "hello")
	m2 := message.NewDefaultMessage(level.Debug, "hello")
	m3 := message.NewDefaultMessage(level.Alert, "")
	m4 := message.NewDefaultMessage(level.Alert, "hello")

	g := message.MakeGroupComposer(m1, m2, m3, m4)

	sender.Send(g)
	s.Equal(mock.numSent, 2)

	mock.failSend = true
	sender.Send(g)
	s.Equal(mock.numSent, 2)
}
