package main

func main() {
	x := 11
	y := 1.0 << x // no const
	println(y)
}

// Error:
// main/shift_d5.gno:5:2-15: operator << not defined on: Float64Kind

// TypeCheckError:
// main/shift_d5.gno:5:7: invalid operation: shifted operand 1.0 (type float64) must be integer
