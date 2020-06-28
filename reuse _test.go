package qio

import (
	"testing"
)

func TestVersionSurported(t *testing.T) {
	versions := []struct {
		a string
		b bool
	}{
		{"4.15.0", true},
		{"3.9.1", true},
		{"2.6.1", false},
	}
	for _, version := range versions {
		if version.b != versionSurported(version.a) {
			t.Fatalf("execpt %t ,got %t", version.b, versionSurported(version.a))
		}
	}
}

func TestKenelSuported(t *testing.T) {
	t.Errorf("%v", reuseSuported())
}
