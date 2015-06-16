package main

import (
	"github.com/boivie/sec/storage"
	"github.com/boivie/sec/bootstrap"
)

func main() {
	s, _ := storage.New()

	bootstrap.Bootstrap(s)

}
