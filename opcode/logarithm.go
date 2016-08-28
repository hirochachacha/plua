package opcode

/*
** converts an integer to a "floating point byte", represented as
** (eeeeexxx), where the real value is (1xxx) * 2^(eeeee - 1) if
** eeeee != 0 and (xxx) otherwise.
 */

func IntToLog(x int) int {
	if x < 8 {
		return x
	}

	var e int /* exponent */
	for x >= 0x10 {
		x = (x + 1) >> 1
		e++
	}
	return ((e + 1) << 3) | (x - 8)
}

// converts back
func LogToInt(x int) int {
	e := (x >> 3) & 0x1f
	if e == 0 {
		return x
	}
	return ((x & 7) + 8) << uint(e-1)
}
