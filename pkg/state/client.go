package state

// Client stores the client's state
type Client interface {
	Database() int
	SetDatabase(int)
	Closed() bool
	SetClosed(bool)
}

type client struct {
	database int
	closed   bool
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
