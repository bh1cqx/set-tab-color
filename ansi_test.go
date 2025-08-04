package main

import (
	"testing"
)

func TestHexToRGB(t *testing.T) {
	tests := []struct {
		name    string
		hex     string
		wantR   int
		wantG   int
		wantB   int
		wantErr bool
	}{
		{
			name:  "valid hex without #",
			hex:   "ff8800",
			wantR: 255,
			wantG: 136,
			wantB: 0,
		},
		{
			name:  "valid hex with #",
			hex:   "#ff8800",
			wantR: 255,
			wantG: 136,
			wantB: 0,
		},
		{
			name:  "black color",
			hex:   "000000",
			wantR: 0,
			wantG: 0,
			wantB: 0,
		},
		{
			name:  "white color",
			hex:   "ffffff",
			wantR: 255,
			wantG: 255,
			wantB: 255,
		},
		{
			name:  "red color",
			hex:   "ff0000",
			wantR: 255,
			wantG: 0,
			wantB: 0,
		},
		{
			name:    "invalid length",
			hex:     "ff88",
			wantErr: true,
		},
		{
			name:    "invalid hex characters",
			hex:     "gghhii",
			wantErr: true,
		},
		{
			name:    "too short",
			hex:     "abc",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotR, gotG, gotB, err := hexToRGB(tt.hex)
			if (err != nil) != tt.wantErr {
				t.Errorf("hexToRGB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if gotR != tt.wantR || gotG != tt.wantG || gotB != tt.wantB {
					t.Errorf("hexToRGB() = (%d, %d, %d), want (%d, %d, %d)", gotR, gotG, gotB, tt.wantR, tt.wantG, tt.wantB)
				}
			}
		})
	}
}

func TestColorText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		hexColor string
		want     string
	}{
		{
			name:     "basic red text",
			text:     "hello",
			hexColor: "ff0000",
			want:     "\033[38;2;255;0;0mhello\033[0m",
		},
		{
			name:     "white text",
			text:     "world",
			hexColor: "ffffff",
			want:     "\033[38;2;255;255;255mworld\033[0m",
		},
		{
			name:     "orange text",
			text:     "orange",
			hexColor: "ff8800",
			want:     "\033[38;2;255;136;0morange\033[0m",
		},
		{
			name:     "invalid hex returns uncolored text",
			text:     "test",
			hexColor: "invalid",
			want:     "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := colorText(tt.text, tt.hexColor)
			if got != tt.want {
				t.Errorf("colorText() = %q, want %q", got, tt.want)
			}
		})
	}
}
