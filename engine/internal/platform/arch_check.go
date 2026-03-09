package platform

// Compile-time 64-bit assertion.
// ^uint(0)>>63 evaluates to 1 on 64-bit and 0 on 32-bit.
// Subtracting 1 gives 0 (valid array length) on 64-bit, and underflows
// to an impossibly large value on 32-bit, producing a compile error.
var _ [^uint(0)>>63 - 1]struct{}
