package main

import (
	"encoding/json"
	"fmt"
	"os"

	resource "github.com/chenbh/imgpkg-resource"
)

func main() {
	req := resource.CheckRequest{}

	decoder := json.NewDecoder(os.Stdin)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Decoding request: %v", err.Error())
		os.Exit(1)
	}

	res, err := resource.RunCheck(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Running check: %v", err.Error())
		os.Exit(1)
	}

	encoder := json.NewEncoder(os.Stdout)
	err = encoder.Encode(&res)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Encoding result: %v", err.Error())
		os.Exit(1)
	}
}
