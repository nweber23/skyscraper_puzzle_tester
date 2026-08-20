package main

import "fmt"

// version is overwritten at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	fmt.Println("rush01-tester", version)
}
