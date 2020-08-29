package testutil

// IntPtr returns an int pointer with a given value.
// Tests cases needed inline int pointer declaration can use this.
func IntPtr(n int) *int {
	return &n
}

// StrPtr returns a string pointer with a given value.
// Tests cases needed inline string pointer declaration can use this.
func StrPtr(s string) *string {
	return &s
}
