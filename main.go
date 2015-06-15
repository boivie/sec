package main

import (
	"github.com/boivie/sec/storage"
	"fmt"
)

func main() {
	s, _ := storage.New()

	fmt.Printf("Got %d\n", s.Add(storage.Record{"hello", 1, []byte("helloworld")}))
	fmt.Printf("Got %d\n", s.GetLastRecordNbr("hello"))
}
