package main

import "github.com/remko/gemsite"

func main() {
	laddr := "0.0.0.0:1965"
	if err := gemsite.Listen(laddr); err != nil {
		panic(err)
	}
}
