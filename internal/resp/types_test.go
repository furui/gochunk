package resp

import (
	"bufio"
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError_Bytes(t *testing.T) {
	e := Error("test")
	tests := []struct {
		name string
		s    *Error
		want []byte
	}{
		{
			name: "normal",
			s:    &e,
			want: []byte("-test\r\n"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := tt.s
			if got := e.Bytes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Error.Bytes() = %v, want %v", got, tt.want)
			}
			if val, ok := e.Value().(error); ok {
				assert.Error(t, val)
				assert.Equal(t, "test", val.Error())
			} else {
				t.Errorf("Error.Value() type = %T, want error", val)
			}
		})
	}
}

func TestError_Stream(t *testing.T) {
	e := Error("test")
	type args struct {
		w *bufio.Writer
	}
	tests := []struct {
		name    string
		s       *Error
		args    args
		want    int
		wantErr bool
	}{
		{
			name:    "normal",
			s:       &e,
			args:    args{w: bufio.NewWriter(bytes.NewBuffer([]byte{}))},
			want:    7,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := tt.s
			got, err := e.Stream(tt.args.w)
			if (err != nil) != tt.wantErr {
				t.Errorf("Error.Stream() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Error.Stream() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSimpleString_Bytes(t *testing.T) {
	test := SimpleString("test")
	tests := []struct {
		name string
		s    *SimpleString
		want []byte
	}{
		{
			name: "normal",
			s:    &test,
			want: []byte("+test\r\n"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Bytes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SimpleString.Bytes() = %v, want %v", got, tt.want)
			}
			if val, ok := tt.s.Value().(string); ok {
				assert.Equal(t, string(*tt.s), val)
			} else {
				t.Errorf("SimpleString.Value() type = %T, want string", val)
			}
		})
	}
}

func TestSimpleString_Stream(t *testing.T) {
	test := SimpleString("test")
	type args struct {
		w *bufio.Writer
	}
	tests := []struct {
		name    string
		s       *SimpleString
		args    args
		want    int
		wantErr bool
	}{
		{
			name:    "normal",
			s:       &test,
			args:    args{w: bufio.NewWriter(bytes.NewBuffer([]byte{}))},
			want:    7,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.Stream(tt.args.w)
			if (err != nil) != tt.wantErr {
				t.Errorf("SimpleString.Stream() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SimpleString.Stream() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInteger_Bytes(t *testing.T) {
	integer := Integer(1234)
	negativeInteger := Integer(-1234)
	tests := []struct {
		name string
		i    *Integer
		want []byte
	}{
		{
			name: "normal",
			i:    &integer,
			want: []byte(":1234\r\n"),
		},
		{
			name: "negative",
			i:    &negativeInteger,
			want: []byte(":-1234\r\n"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.i.Bytes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Integer.Bytes() = %v, want %v", got, tt.want)
			}
			if val, ok := tt.i.Value().(int64); ok {
				assert.NotZero(t, val)
				assert.Equal(t, int64(*tt.i), val)
			} else {
				t.Errorf("Integer.Value() type = %T, want int64", val)
			}
		})
	}
}

func TestInteger_Stream(t *testing.T) {
	integer := Integer(1234)
	negativeInteger := Integer(-1234)
	type args struct {
		w *bufio.Writer
	}
	tests := []struct {
		name    string
		i       *Integer
		args    args
		want    int
		wantErr bool
	}{
		{
			name:    "normal",
			i:       &integer,
			args:    args{w: bufio.NewWriter(bytes.NewBuffer([]byte{}))},
			want:    7,
			wantErr: false,
		},
		{
			name:    "negative",
			i:       &negativeInteger,
			args:    args{w: bufio.NewWriter(bytes.NewBuffer([]byte{}))},
			want:    8,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.i.Stream(tt.args.w)
			if (err != nil) != tt.wantErr {
				t.Errorf("Integer.Stream() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Integer.Stream() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBulkString_Bytes(t *testing.T) {
	type fields struct {
		Data []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   []byte
	}{
		{
			name:   "normal",
			fields: fields{Data: []byte("test")},
			want:   []byte("$4\r\ntest\r\n"),
		},
		{
			name:   "empty",
			fields: fields{Data: []byte("")},
			want:   []byte("$0\r\n\r\n"),
		},
		{
			name:   "null",
			fields: fields{Data: nil},
			want:   []byte("$-1\r\n"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &BulkString{
				Data: tt.fields.Data,
			}
			if got := s.Bytes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BulkString.Bytes() = %v, want %v", got, tt.want)
			}
			if val, ok := s.Value().([]byte); ok {
				assert.Equal(t, tt.fields.Data, val)
			} else {
				t.Errorf("BulkString.Value() type = %T, want []byte", val)
			}
		})
	}
}

func TestBulkString_Stream(t *testing.T) {
	type fields struct {
		Data []byte
	}
	type args struct {
		w *bufio.Writer
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		{
			name:    "normal",
			fields:  fields{Data: []byte("test")},
			args:    args{w: bufio.NewWriter(bytes.NewBuffer([]byte{}))},
			want:    10,
			wantErr: false,
		},
		{
			name:    "empty",
			fields:  fields{Data: []byte("")},
			args:    args{w: bufio.NewWriter(bytes.NewBuffer([]byte{}))},
			want:    6,
			wantErr: false,
		},
		{
			name:    "null",
			fields:  fields{Data: nil},
			args:    args{w: bufio.NewWriter(bytes.NewBuffer([]byte{}))},
			want:    5,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &BulkString{
				Data: tt.fields.Data,
			}
			got, err := s.Stream(tt.args.w)
			if (err != nil) != tt.wantErr {
				t.Errorf("BulkString.Stream() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("BulkString.Stream() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestArray_Bytes(t *testing.T) {
	integer := Integer(1234)
	simpleString := SimpleString("bar")
	type fields struct {
		Contents []Type
	}
	tests := []struct {
		name   string
		fields fields
		want   []byte
	}{
		{
			name: "two bulk strings",
			fields: fields{
				Contents: []Type{
					&BulkString{Data: []byte("foo")},
					&BulkString{Data: []byte("bar")},
				},
			},
			want: []byte("*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"),
		},
		{
			name: "empty",
			fields: fields{
				Contents: []Type{},
			},
			want: []byte("*0\r\n"),
		},
		{
			name: "mixed",
			fields: fields{
				Contents: []Type{
					&BulkString{Data: []byte("foo")},
					&integer,
					&simpleString,
				},
			},
			want: []byte("*3\r\n$3\r\nfoo\r\n:1234\r\n+bar\r\n"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Array{
				Contents: tt.fields.Contents,
			}
			if got := a.Bytes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Array.Bytes() = %v, want %v", got, tt.want)
			}
			if val, ok := a.Value().([]Type); ok {
				assert.Equal(t, tt.fields.Contents, val)
			} else {
				t.Errorf("Array.Value() type = %T, want []Type", val)
			}
		})
	}
}

func TestArray_Stream(t *testing.T) {
	integer := Integer(1234)
	simpleString := SimpleString("bar")
	type fields struct {
		Contents []Type
	}
	type args struct {
		w *bufio.Writer
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "two bulk strings",
			fields: fields{
				Contents: []Type{
					&BulkString{Data: []byte("foo")},
					&BulkString{Data: []byte("bar")},
				},
			},
			args:    args{w: bufio.NewWriter(bytes.NewBuffer([]byte{}))},
			want:    22,
			wantErr: false,
		},
		{
			name:    "empty",
			fields:  fields{Contents: []Type{}},
			args:    args{w: bufio.NewWriter(bytes.NewBuffer([]byte{}))},
			want:    4,
			wantErr: false,
		},
		{
			name: "mixed",
			fields: fields{
				Contents: []Type{
					&BulkString{Data: []byte("foo")},
					&integer,
					&simpleString,
				},
			},
			args:    args{w: bufio.NewWriter(bytes.NewBuffer([]byte{}))},
			want:    26,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Array{
				Contents: tt.fields.Contents,
			}
			got, err := a.Stream(tt.args.w)
			if (err != nil) != tt.wantErr {
				t.Errorf("Array.Stream() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Array.Stream() = %v, want %v", got, tt.want)
			}
		})
	}
}
