package resp

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	respTypes "github.com/furui/gochunk/pkg/types"
)

var (
	// ErrClosedConnection is thrown when the connection is closed
	ErrClosedConnection = errors.New("closed connection")
)

// Scanner provides an interface for scanning in RESP from IO
type Scanner struct {
	r          io.Reader
	read       *bufio.Reader
	err        error
	scanCalled bool
	done       bool
	t          respTypes.Type
}

// NewScanner returns a new scanner
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{
		r:    r,
		read: bufio.NewReader(r),
	}
}

// Scan the next type in until connection is closed
func (s *Scanner) Scan() bool {
	if s.done {
		return false
	}
	s.err = nil
	s.t, s.err = s.scanType()
	if s.err == io.EOF || s.err == io.ErrClosedPipe || s.err == io.ErrUnexpectedEOF {
		s.done = true
		return false
	}
	return true
}

// Err returns an error except io.EOF
func (s *Scanner) Err() error {
	if s.err == io.EOF {
		return nil
	}
	return s.err
}

// Type returns the last type read
func (s *Scanner) Type() respTypes.Type {
	return s.t
}

func (s *Scanner) scanType() (respTypes.Type, error) {
	val, err := s.read.ReadSlice('\n')
	if err != nil {
		return nil, err
	}
	if len(val) == 0 {
		return nil, ErrClosedConnection
	}
	// Trim off the CR
	if val[len(val)-2] != '\r' {
		s.read.Reset(s.r)
		return nil, fmt.Errorf("Invalid CRLF, expected \"\\r\\n\"")
	}
	val = val[:len(val)-2]

	switch val[0] {
	case '+':
		t := respTypes.SimpleString(string(val[1:]))
		return &t, nil
	case '-':
		t := respTypes.Error(string(val[1:]))
		return &t, nil
	case ':':
		i, err := strconv.ParseInt(string(val[1:]), 10, 64)
		if err != nil {
			s.read.Reset(s.r)
			return nil, err
		}
		t := respTypes.Integer(i)
		return &t, nil
	case '$':
		i, err := strconv.ParseInt(string(val[1:]), 10, 64)
		if err != nil {
			s.read.Reset(s.r)
			return nil, err
		}
		if i < 0 {
			return &respTypes.BulkString{Data: nil}, nil
		}
		buf := make([]byte, i, i)
		if i > 0 {
			n, err := io.ReadFull(s.read, buf)
			if err != nil {
				return nil, err
			}
			if int64(n) < i {
				s.read.Reset(s.r)
				return nil, fmt.Errorf("Expected %d bytes but got %d bytes", i, n)
			}
		}
		crlf := make([]byte, 2, 2)
		n, err := io.ReadFull(s.read, crlf)
		if err != nil {
			return nil, err
		}
		if n != 2 {
			s.read.Reset(s.r)
			return nil, fmt.Errorf("Unexpected early end of string")
		}
		if bytes.Compare(crlf, []byte{'\r', '\n'}) != 0 {
			s.read.Reset(s.r)
			return nil, fmt.Errorf("Terminating CRLF not found")
		}
		return &respTypes.BulkString{Data: buf}, nil
	case '*':
		n, err := strconv.ParseInt(string(val[1:]), 10, 64)
		if err != nil {
			return nil, err
		}
		if n < 0 {
			s.read.Reset(s.r)
			return nil, fmt.Errorf("Number of indexes must be zero or positive")
		}
		t := []respTypes.Type{}
		for i := int64(0); i < n; i++ {
			v, err := s.scanType()
			if err != nil {
				return nil, err
			}
			t = append(t, v)
		}
		return &respTypes.Array{Contents: t}, nil
	default:
		r := csv.NewReader(strings.NewReader(string(val)))
		r.Comma = ' '
		record, err := r.Read()
		if err != nil {
			return nil, err
		}
		t := []respTypes.Type{}
		for _, s := range record {
			t = append(t, &respTypes.BulkString{Data: []byte(s)})
		}
		return &respTypes.Array{Contents: t}, nil
	}
}
