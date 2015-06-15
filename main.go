package main

import (
	"github.com/boivie/sec/storage"
	"fmt"
)

func main() {
	s, _ := storage.New()

	fmt.Printf("Got %v\n", s.Add(storage.Record{"hello", 1, []byte("helloworld")}))
	fmt.Printf("Got %v\n", s.GetLastRecordNbr("hello"))

	written, _ := s.Append([]storage.Record{storage.Record{"hello", 1, []byte("helloworld")}})
	fmt.Printf("Got %v\n", written)
	data, _ := s.GetOne("hello", 1)
	fmt.Printf("Got %v\n", data)
}
