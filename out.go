package resource

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/google/go-containerregistry/pkg/name"
	imgpkg "github.com/vmware-tanzu/carvel-imgpkg/pkg/imgpkg/cmd"
)

var regex = regexp.MustCompile("Pushed '(.+)'")

func RunPut(req OutRequest, dest string) (*OutResponse, error) {
	var err error
	tag := "latest"
	bundle := req.Source.Repository
	if req.Source.Tag != "" {
		tag = req.Source.Tag
	}
	if req.Params.Tag != "" {
		tag = req.Params.Tag
	}
	bundle = fmt.Sprintf("%v:%v", bundle, tag)

	certs, err := writeCACertsToFile(req.Source.CaCerts)
	if err != nil {
		return nil, fmt.Errorf("writing certs to file: %v", err)
	}

	reader, writer := io.Pipe()
	// parse output and extract digest from image
	done := make(chan error)
	var digest string
	go func() {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			parts := regex.FindStringSubmatch(scanner.Text())
			if len(parts) == 2 {
				image, err := name.NewDigest(parts[1])
				if err != nil {
					done <- err
				}

				digest = image.DigestStr()
				done <- nil
			}
		}
		done <- fmt.Errorf("failed to parse image digest from output")
	}()

	out := io.MultiWriter(os.Stderr, writer)
	ui := ui.NewWriterUI(out, out, nil)

	cmd := imgpkg.NewPushOptions(ui)
	cmd.RegistryFlags = toRegistryFlags(req.Source, certs)
	cmd.BundleFlags = imgpkg.BundleFlags{Bundle: bundle}
	cmd.FileFlags = imgpkg.FileFlags{Files: []string{filepath.Join(dest, req.Params.Path)}}

	err = cmd.Run()
	if err != nil {
		return nil, err
	}
	writer.Close() // have to manually close the pipe

	err = <-done
	if err != nil {
		return nil, err
	}

	res := &OutResponse{
		Version{
			Digest: digest,
			Tag:    tag,
		},
	}
	return res, nil
}
