package lisgo

const maxSliceLen = 1 << 24

//http://cavaliercoder.com/blog/optimized-abs-for-int64-in-go.html
func abs(n int32) int32 {
	y := n >> 31
	return (n ^ y) - y
}

func pad4(n uint32) uint32 {
	a := n % 4
	if a == 0 {
		return n
	}
	return 4 - a + n
}