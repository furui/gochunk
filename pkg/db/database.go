package db

// Database is a structure for accessing a database
type Database interface {
	Close() error
}

type database struct {
}

// NewDatabase returns a database
func NewDatabase(filename string) Database {
	return &database{}
}

func (d *database) Close() error {
	return nil
}
