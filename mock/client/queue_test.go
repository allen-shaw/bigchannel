package main

import (
	"fmt"
	"testing"
)

func TestSlice(t *testing.T) {
	s := make([]int, 0)
	for i := 0; i < 10; i++ {
		s = append(s, i)
	}

	fmt.Println(s)

	s1 := s[:3]
	s2 := s[3:]
	s3 := s[2:5]

	fmt.Println(s1)
	fmt.Println(s2)
	fmt.Println(s3)
}
