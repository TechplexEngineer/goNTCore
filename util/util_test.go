package util

import (
	"bytes"
	"io"
	"testing"
)

func TestEncodeULeb128(t *testing.T) {
	type args struct {
		value uint32
	}
	tests := []struct {
		name       string
		args       args
		wantWriter []byte
		wantErr    bool
	}{
		// examples taken from dwarf spec page 140
		// http://dwarfstd.org/doc/Dwarf3.pdf
		{
			name:       "encode 2",
			args:       args{2},
			wantWriter: []byte{2},
			wantErr:    false,
		},
		{
			name:       "encode 127",
			args:       args{127},
			wantWriter: []byte{127},
			wantErr:    false,
		},
		{
			name:       "encode 128",
			args:       args{128},
			wantWriter: []byte{0 + 0x80, 1},
			wantErr:    false,
		},
		{
			name:       "encode 129",
			args:       args{129},
			wantWriter: []byte{1 + 0x80, 1},
			wantErr:    false,
		},
		{
			name:       "encode 130",
			args:       args{130},
			wantWriter: []byte{2 + 0x80, 1},
			wantErr:    false,
		},
		{
			name:       "encode 12857",
			args:       args{12857},
			wantWriter: []byte{57 + 0x80, 100},
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := &bytes.Buffer{}
			err := EncodeULeb128(tt.args.value, writer)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncodeULeb128() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if bytes.Compare(writer.Bytes(), tt.wantWriter) != 0 {
				t.Errorf("EncodeULeb128() gotWriter = %#x, want %#x", writer.Bytes(), tt.wantWriter)
			}
		})
	}
}

func TestDecodeULeb128(t *testing.T) {
	type args struct {
		reader io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    uint32
		wantErr bool
	}{
		// examples taken from dwarf spec page 140
		// http://dwarfstd.org/doc/Dwarf3.pdf
		{
			name:    "decode 2",
			args:    args{bytes.NewReader([]byte{2})},
			want:    2,
			wantErr: false,
		},
		{
			name:    "decode 127",
			args:    args{bytes.NewBuffer([]byte{127})},
			want:    127,
			wantErr: false,
		},
		{
			name:    "decode 128",
			args:    args{bytes.NewBuffer([]byte{0 + 0x80, 1})},
			want:    128,
			wantErr: false,
		},
		{
			name:    "decode 129",
			args:    args{bytes.NewBuffer([]byte{1 + 0x80, 1})},
			want:    129,
			wantErr: false,
		},
		{
			name:    "decode 130",
			args:    args{bytes.NewBuffer([]byte{2 + 0x80, 1})},
			want:    130,
			wantErr: false,
		},
		{
			name:    "decode 12857",
			args:    args{bytes.NewBuffer([]byte{57 + 0x80, 100})},
			want:    12857,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeULeb128(tt.args.reader)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeULeb128() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DecodeULeb128() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSanitizeKey(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"no change",
			args{"/smartdashboard"},
			"/smartdashboard",
		},
		{
			"add leading",
			args{"smartdashboard"},
			"/smartdashboard",
		},
		{
			"remove trailing",
			args{"/smartdashboard/"},
			"/smartdashboard",
		},
		{
			"many levels",
			args{"/smartdashboard/this/is/a/Test/"},
			"/smartdashboard/this/is/a/Test",
		},
		{
			"many trailing",
			args{"/smartdashboard/this/is/a/Test////"},
			"/smartdashboard/this/is/a/Test",
		},
		{
			"many leading",
			args{"////smartdashboard/this/is/a/Test"},
			"/smartdashboard/this/is/a/Test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SanitizeKey(tt.args.key); got != tt.want {
				t.Errorf("SanitizeKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
