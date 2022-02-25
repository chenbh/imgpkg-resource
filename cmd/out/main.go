package main

import (
	"encoding/json"
	"fmt"
	"os"

	resource "github.com/chenbh/imgpkg-resource"
)

func main() {
	stdin := os.Stdin
	stdout := os.Stdout
	stderr := os.Stderr

	path := os.Args[1]
	req := resource.OutRequest{}

	decoder := json.NewDecoder(stdin)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&req)
	if err != nil {
		fmt.Fprintf(stderr, "Decoding request: %v", err.Error())
		os.Exit(1)
	}

	// stdout is used to communicate to concourse, stderr is shown to users
	os.Stdout = stderr
	res, err := resource.RunPut(req, path)
	if err != nil {
		fmt.Fprintf(stderr, "Running put: %v", err.Error())
		os.Exit(1)
	}

	encoder := json.NewEncoder(stdout)
	err = encoder.Encode(&res)
	if err != nil {
		fmt.Fprintf(stderr, "Encoding result: %v", err.Error())
		os.Exit(1)
	}
}
