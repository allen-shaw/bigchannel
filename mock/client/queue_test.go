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

func TestSlice2(t *testing.T) {
	data := make([]int, 0, 100)
	data = append(data, 1, 2, 3, 4, 5)

	data2 := data[:50]
	data3 := data[:len(data)]
	fmt.Println(len(data))
	fmt.Println("data2:", data2)
	fmt.Println("data3:", data3)
}
