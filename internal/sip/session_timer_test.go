package sip

import "testing"

func TestParseSessionExpires(t *testing.T) {
	tests := []struct {
		input     string
		interval  uint32
		refresher string
	}{
		{"1800", 1800, ""},
		{"1800;refresher=uac", 1800, "uac"},
		{"1800;refresher=uas", 1800, "uas"},
		{"  900 ; refresher=uac ", 900, "uac"},
		{"", 0, ""},
		{"0", 0, ""},
		{"abc", 0, ""},
		{"1800;foo=bar;refresher=uac", 1800, "uac"},
	}

	for _, tt := range tests {
		interval, refresher := ParseSessionExpires(tt.input)
		if interval != tt.interval || refresher != tt.refresher {
			t.Errorf("ParseSessionExpires(%q) = (%d, %q), want (%d, %q)",
				tt.input, interval, refresher, tt.interval, tt.refresher)
		}
	}
}

func TestParseMinSE(t *testing.T) {
	tests := []struct {
		input string
		want  uint32
	}{
		{"90", 90},
		{"120;refresher=uas", 120},
		{"", 0},
		{"abc", 0},
	}

	for _, tt := range tests {
		got := ParseMinSE(tt.input)
		if got != tt.want {
			t.Errorf("ParseMinSE(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestFormatSessionExpires(t *testing.T) {
	got := FormatSessionExpires(1800, "uac")
	want := "1800;refresher=uac"
	if got != want {
		t.Errorf("FormatSessionExpires(1800, uac) = %q, want %q", got, want)
	}
}
