package types

import (
	"bufio"
	"fmt"
)

// Type is a RESP type
type Type interface {
	Bytes() []byte
	Stream(*bufio.Writer) (int, error)
	Value() interface{}
}

// Error is a RESP error type
type Error string

// Bytes representation
func (e *Error) Bytes() []byte {
	return []byte(fmt.Sprintf("-%s\r\n", *e))
}

// Stream the bytes to the writer
func (e *Error) Stream(w *bufio.Writer) (int, error) {
	return w.Write(e.Bytes())
}

// Error returns the error message
func (e *Error) Error() string {
	return string(*e)
}

// Value of the type
func (e *Error) Value() interface{} {
	return error(e)
}

// SimpleString is a RESP simple string type
type SimpleString string

// Bytes representation
func (s *SimpleString) Bytes() []byte {
	return []byte(fmt.Sprintf("+%s\r\n", *s))
}

// Stream the bytes to the writer
func (s *SimpleString) Stream(w *bufio.Writer) (int, error) {
	return w.Write(s.Bytes())
}

// Value of the type
func (s *SimpleString) Value() interface{} {
	return string(*s)
}

// Integer is a RESP integer type
type Integer int64

// Bytes returns the bytes representation
func (i *Integer) Bytes() []byte {
	return []byte(fmt.Sprintf(":%d\r\n", *i))
}

// Stream the bytes to the writer
func (i *Integer) Stream(w *bufio.Writer) (int, error) {
	return w.Write(i.Bytes())
}

// Value of the type
func (i *Integer) Value() interface{} {
	return int64(*i)
}

// BulkString is a RESP bulk string type
type BulkString struct {
	Data []byte
}

// Bytes returns the bytes representation
func (s *BulkString) Bytes() []byte {
	if s.Data == nil {
		return []byte("$-1\r\n")
	}
	return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(s.Data), s.Data))
}

// Stream the bytes to the writer
func (s *BulkString) Stream(w *bufio.Writer) (int, error) {
	l := len(s.Data)
	if s.Data == nil {
		return w.Write([]byte("$-1\r\n"))
	}
	nn1, err := w.Write([]byte(fmt.Sprintf("$%d\r\n", l)))
	if err != nil {
		return nn1, err
	}
	nn2, err := w.Write(s.Data)
	if err != nil {
		return nn1 + nn2, err
	}
	nn3, err := w.Write([]byte("\r\n"))
	return nn1 + nn2 + nn3, err
}

// Value of the type
func (s *BulkString) Value() interface{} {
	return s.Data
}

// Array is a RESP array type
type Array struct {
	Contents []Type
}

// Bytes returns the bytes representation
func (a *Array) Bytes() []byte {
	output := []byte(fmt.Sprintf("*%d\r\n", len(a.Contents)))
	for _, t := range a.Contents {
		output = append(output[:], t.Bytes()[:]...)
	}
	return output
}

// Stream the bytes to the writer
func (a *Array) Stream(w *bufio.Writer) (int, error) {
	total, err := w.Write([]byte(fmt.Sprintf("*%d\r\n", len(a.Contents))))
	if err != nil {
		return total, err
	}
	for _, t := range a.Contents {
		nn, err := t.Stream(w)
		if err != nil {
			return total + nn, err
		}
		total += nn
	}
	return total, nil
}

// Value of the type
func (a *Array) Value() interface{} {
	return a.Contents
}
