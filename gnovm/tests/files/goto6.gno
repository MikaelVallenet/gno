package main

func main() {
	for {
		goto here
	nowhere:
		panic("should not happen")
	there:
		println("there")
		return
	here:
		println("here")
		switch 1 {
		case 1:
			goto there
		default:
			panic("should not happen")
		}
	}
}

// Output:
// here
// there

// TypeCheckError:
// main/goto6.gno:6:2: label nowhere declared and not used
