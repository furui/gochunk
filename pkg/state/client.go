package state

import "errors"

var (
	// ErrNoPassSet is thrown when authing with no pass set
	ErrNoPassSet = errors.New("Client sent AUTH, but no password set")
)

// Client stores the client's state
type Client interface {
	Database() int
	SetDatabase(int)
	Closed() bool
	SetClosed(bool)
	Authenticated() bool
	Authenticate(string) (bool, error)
	SetAuthRequired(string)
	SetRemoteAddr(string)
	RemoteAddr() string
}

type client struct {
	database     int
	closed       bool
	authed       bool
	authRequired string
	remoteAddr   string
}

func (s *client) SetRemoteAddr(addr string) {
	s.remoteAddr = addr
}

func (s *client) RemoteAddr() string {
	return s.remoteAddr
}

func (s *client) Authenticated() bool {
	return s.authed
}

func (s *client) Authenticate(passwd string) (bool, error) {
	if len(s.authRequired) == 0 {
		s.authed = true
		return true, ErrNoPassSet
	}
	if s.authRequired == passwd {
		s.authed = true
		return true, nil
	}
	s.authed = false
	return false, nil
}

func (s *client) SetAuthRequired(passwd string) {
	if len(passwd) > 0 {
		s.authed = false
	} else {
		s.authed = true
	}
	s.authRequired = passwd
}

func (s *client) Database() (db int) {
	return s.database
}

func (s *client) SetDatabase(db int) {
	s.database = db
}

func (s *client) Closed() bool {
	return s.closed
}

func (s *client) SetClosed(b bool) {
	s.closed = b
}

// NewClient returns a new client state
func NewClient() Client {
	return &client{
		database: 0,
		closed:   false,
	}
}
