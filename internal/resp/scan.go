package resp

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
)

// Scanner provides an interface for scanning in RESP from IO
type Scanner struct {
	r          io.Reader
	read       *bufio.Reader
	err        error
	scanCalled bool
	done       bool
	t          Type
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
	if s.err == io.EOF {
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
func (s *Scanner) Type() Type {
	return s.t
}

func (s *Scanner) scanType() (Type, error) {
	val, err := s.read.ReadSlice('\n')
	if err != nil {
		return nil, err
	}
	// Trim off the CR
	if val[len(val)-2] != '\r' {
		s.read.Reset(s.r)
		return nil, fmt.Errorf("Invalid CRLF, expected \"\\r\\n\"")
	}
	val = val[:len(val)-2]

	switch val[0] {
	case '+':
		t := SimpleString(string(val[1:]))
		return &t, nil
	case '-':
		t := Error(string(val[1:]))
		return &t, nil
	case ':':
		i, err := strconv.ParseInt(string(val[1:]), 10, 64)
		if err != nil {
			s.read.Reset(s.r)
			return nil, err
		}
		t := Integer(i)
		return &t, nil
	case '$':
		i, err := strconv.ParseInt(string(val[1:]), 10, 64)
		if err != nil {
			s.read.Reset(s.r)
			return nil, err
		}
		if i < 0 {
			return &BulkString{Data: nil}, nil
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
		return &BulkString{Data: buf}, nil
	case '*':
		n, err := strconv.ParseInt(string(val[1:]), 10, 64)
		if err != nil {
			return nil, err
		}
		if n < 0 {
			s.read.Reset(s.r)
			return nil, fmt.Errorf("Number of indexes must be zero or positive")
		}
		t := []Type{}
		for i := int64(0); i < n; i++ {
			v, err := s.scanType()
			if err != nil {
				return nil, err
			}
			t = append(t, v)
		}
		return &Array{Contents: t}, nil
	default:
		return nil, fmt.Errorf("Unknown type")
	}
}
