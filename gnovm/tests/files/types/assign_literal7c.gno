package main

import "fmt"

func main() {
	str := "hello"
	str[0] = 'y'
	fmt.Println(str[0])
	fmt.Printf("%c\n", str[0])
}

// Error:
// main/files/types/assign_literal7c.gno:7:2: cannot assign to str<VPBlock(1,0)>[(const (0 int))]

// TypeCheckError:
// main/files/types/assign_literal7c.gno:7:2: cannot assign to str[0] (neither addressable nor a map index expression)
