package main

import (
	"encoding/json"
	"fmt"
	"os"

	resource "github.com/chenbh/imgpkg-resource"
)

func main() {
	path := os.Args[1]
	req := resource.InRequest{}

	decoder := json.NewDecoder(os.Stdin)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Decoding request: %v", err.Error())
		os.Exit(1)
	}

	res, err := resource.RunGet(req, path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Running get: %v", err.Error())
		os.Exit(1)
	}

	encoder := json.NewEncoder(os.Stdout)
	err = encoder.Encode(&res)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Encoding result: %v", err.Error())
		os.Exit(1)
	}
}
