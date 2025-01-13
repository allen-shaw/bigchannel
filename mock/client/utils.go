package main

import (
	"fmt"

	"github.com/google/uuid"
)

func UUID() []byte {
	id, err := uuid.NewUUID()
	if err != nil {
		panic(fmt.Errorf("new uuid %w", err))
	}
	uid, _ := id.MarshalBinary()
	return uid
}
