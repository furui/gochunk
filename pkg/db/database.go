package db

// Database is a structure for accessing a database
type Database interface {
}

type database struct {
}

// NewDatabase returns a database
func NewDatabase(filename string) Database {
	return &database{}
}
