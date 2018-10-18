package types

import (
	"bytes"
	"io"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScanner_Type(t *testing.T) {
	ss := SimpleString("OK")
	e := Error("BAD")
	i := Integer(1234)
	tests := []struct {
		name string
		r    io.Reader
		want Type
	}{
		{
			name: "simple string",
			r:    bytes.NewBufferString("+OK\r\n"),
			want: &ss,
		},
		{
			name: "error",
			r:    bytes.NewBufferString("-BAD\r\n"),
			want: &e,
		},
		{
			name: "integer",
			r:    bytes.NewBufferString(":1234\r\n"),
			want: &i,
		},
		{
			name: "bulk string",
			r:    bytes.NewBufferString("$2\r\nOK\r\n"),
			want: &BulkString{Data: []byte("OK")},
		},
		{
			name: "empty bulk string",
			r:    bytes.NewBufferString("$0\r\n\r\n"),
			want: &BulkString{Data: []byte{}},
		},
		{
			name: "null bulk string",
			r:    bytes.NewBufferString("$-1\r\n"),
			want: &BulkString{Data: nil},
		},
		{
			name: "array",
			r:    bytes.NewBufferString("*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"),
			want: &Array{
				Contents: []Type{
					&BulkString{Data: []byte("foo")},
					&BulkString{Data: []byte("bar")},
				},
			},
		},
		{
			name: "empty array",
			r:    bytes.NewBufferString("*0\r\n"),
			want: &Array{
				Contents: []Type{},
			},
		},
		{
			name: "inline command",
			r:    bytes.NewBufferString("foo \"bar\"\r\n"),
			want: &Array{
				Contents: []Type{
					&BulkString{Data: []byte("foo")},
					&BulkString{Data: []byte("bar")},
				},
			},
		},
		{
			name: "inline command long quotes",
			r:    bytes.NewBufferString("foo \"bar test\"\r\n"),
			want: &Array{
				Contents: []Type{
					&BulkString{Data: []byte("foo")},
					&BulkString{Data: []byte("bar test")},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScanner(tt.r)
			res := s.Scan()
			if res == false {
				t.Errorf("s.Scan() returned false && s.Err() = %s", s.Err())
			}
			if s.Err() != nil {
				t.Errorf("s.Err() = %s", s.Err())
			}
			if got := s.Type(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Scanner.Type() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScanner_Err(t *testing.T) {
	tests := []struct {
		name    string
		r       io.Reader
		message string
	}{
		{
			name:    "missing CR",
			r:       bytes.NewBufferString("+OK\n"),
			message: "Invalid CRLF, expected \"\\r\\n\"",
		},
		{
			name:    "invalid value",
			r:       bytes.NewBufferString(":abc\r\n"),
			message: "strconv.ParseInt: parsing \"abc\": invalid syntax",
		},
		{
			name:    "too many bytes",
			r:       bytes.NewBufferString("$3\r\nfoobar\r\n"),
			message: "Terminating CRLF not found",
		},
		{
			name:    "too few bytes",
			r:       bytes.NewBufferString("$6\r\nfoo\r\n"),
			message: "unexpected EOF",
		},
		{
			name:    "invalid length",
			r:       bytes.NewBufferString("$abc\r\n"),
			message: "strconv.ParseInt: parsing \"abc\": invalid syntax",
		},
		{
			name:    "negative index length",
			r:       bytes.NewBufferString("*-1\r\n\r\n"),
			message: "Number of indexes must be zero or positive",
		},
		{
			name:    "invalid index length",
			r:       bytes.NewBufferString("*abc\r\n"),
			message: "strconv.ParseInt: parsing \"abc\": invalid syntax",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScanner(tt.r)
			s.Scan()
			assert.Error(t, s.Err())
			assert.Equal(t, tt.message, s.Err().Error())
		})
	}
}

func TestScanner_ErrReset(t *testing.T) {
	tests := []struct {
		name  string
		r     *bytes.Buffer
		line1 string
		line2 string
	}{
		{
			name:  "after invalid value",
			r:     bytes.NewBufferString(""),
			line1: ":abc\r\n",
			line2: ":123\r\n",
		},
		{
			name:  "garbage at end",
			r:     bytes.NewBufferString(""),
			line1: ":abc\r\n:*+-",
			line2: ":123\r\n",
		},
		{
			name:  "bad array",
			r:     bytes.NewBufferString(""),
			line1: "*2\r\n:abc\r\n+123\r\n",
			line2: ":123\r\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScanner(tt.r)
			tt.r.WriteString(tt.line1)
			s.Scan()
			assert.Error(t, s.Err())
			tt.r.WriteString(tt.line2)
			s.Scan()
			assert.NoError(t, s.Err())
		})
	}
}
