package testlib

// IntPtr returns an int pointer with a given value.
// Tests cases needed inline int pointer declaration can use this.
func IntPtr(n int) *int {
	return &n
}
