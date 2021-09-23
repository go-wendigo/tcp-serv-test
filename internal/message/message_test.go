package message

import (
	"bytes"
	"io"
	"math"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func TestEncode(t *testing.T) {
	maxLenString := randString(math.MaxUint16)
	type args struct {
		msg string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name:    "case1",
			args:    args{msg: "AB"},
			want:    []byte{0x00, 0x02, 0x41, 0x42},
			wantErr: false,
		},
		{
			name:    "case2",
			args:    args{msg: "ABA"},
			want:    []byte{0x00, 0x03, 0x41, 0x42, 0x41},
			wantErr: false,
		},
		{
			name:    "string max len",
			args:    args{msg: maxLenString},
			want:    append([]byte{0xff, 0xff}, []byte(maxLenString)...),
			wantErr: false,
		},
		{
			name:    "too long string",
			args:    args{msg: randString(math.MaxUint16 + 1)},
			want:    []byte{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Encode(tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Encode() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"case 1", args{data: []byte{0x00, 0x02, 0x41, 0x42}}, "AB", false},
		{"case 2", args{data: []byte{0x00, 0x03, 0x41, 0x42, 0x41}}, "ABA", false},
		{"wrong length", args{data: []byte{0x00, 0x02, 0x41, 0x42, 0x41, 0x42}}, "", true},
		{"empty content message", args{data: []byte{0x00, 0x00}}, "", false},
		{"empty data", args{data: []byte{}}, "", true},
		{"one byte", args{data: []byte{0x03}}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Decode(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Decode() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func randString(l int) string {
	rand.Seed(time.Now().UnixNano())
	var letter = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]rune, l)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
}

func TestRead(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			"message ok",
			args{r: bytes.NewReader([]byte{0, 1, 'A'})},
			[]byte{0, 1, 'A'},
			false,
		},
		{
			"wrong msg format",
			args{r: bytes.NewReader([]byte{0, 4, 'A'})},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Read(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Read() got = %v, want %v", got, tt.want)
			}
		})
	}
}
