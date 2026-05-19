package provider

import (
	"reflect"
	"testing"
)

func TestEmptyIfNil(t *testing.T) {
	t.Parallel()

	t.Run("nil string slice becomes empty", func(t *testing.T) {
		var in []string
		got := emptyIfNil(in)
		if got == nil {
			t.Fatal("expected non-nil slice")
		}
		if len(got) != 0 {
			t.Fatalf("expected empty slice, got %v", got)
		}
	})

	t.Run("non-nil slice is returned as-is", func(t *testing.T) {
		in := []string{"a", "b"}
		got := emptyIfNil(in)
		if !reflect.DeepEqual(got, in) {
			t.Fatalf("expected %v, got %v", in, got)
		}
	})

	t.Run("works with int slices", func(t *testing.T) {
		var in []int
		got := emptyIfNil(in)
		if got == nil || len(got) != 0 {
			t.Fatalf("expected non-nil empty int slice, got %v", got)
		}
	})
}

func TestStringOrDefault(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		s    string
		def  string
		want string
	}{
		{"empty returns default", "", "fallback", "fallback"},
		{"non-empty returns value", "value", "fallback", "value"},
		{"both empty returns empty", "", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stringOrDefault(tt.s, tt.def); got != tt.want {
				t.Errorf("stringOrDefault(%q, %q) = %q, want %q", tt.s, tt.def, got, tt.want)
			}
		})
	}
}
