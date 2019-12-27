package lisgo

const maxSliceLen = 1 << 24

//http://cavaliercoder.com/blog/optimized-abs-for-int64-in-go.html
func abs(n int32) int32 {
	y := n >> 31
	return (n ^ y) - y
}

func pad8(n uint32) uint32 {
	return padx(n, 8)
}

func pad4(n uint32) uint32 {
	return padx(n, 4)
}

func padx(n uint32, x uint32) uint32 {
	a := n % x
	if a == 0 {
		return n
	}
	return x - a + n
}
