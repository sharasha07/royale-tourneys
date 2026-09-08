package assert

import "testing"

func Equal[T comparable](t *testing.T, want, got T) {
	t.Helper()

	if want != got {
		t.Errorf("want: %v, got: %v", want, got)
	}
}
