package resource

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cppforlife/go-cli-ui/ui"
	imgpkg "github.com/vmware-tanzu/carvel-imgpkg/pkg/imgpkg/cmd"
)

func RunGet(req InRequest, dest string) (*InResponse, error) {
	var err error
	bundle := fmt.Sprintf("%v@%v", req.Source.Repository, req.Version.Digest)

	certs, err := writeCACertsToFile(req.Source.CaCerts)
	if err != nil {
		return nil, fmt.Errorf("writing certs to file: %v", err)
	}

	ui := ui.NewWriterUI(os.Stderr, os.Stderr, nil)
	cmd := imgpkg.NewPullOptions(ui)
	cmd.RegistryFlags = toRegistryFlags(req.Source, certs)
	cmd.BundleFlags = imgpkg.BundleFlags{
		Bundle: bundle,
	}
	cmd.OutputPath = filepath.Join(dest, "output")

	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("running command: %v", err)
	}

	err = saveVersionInfo(dest, req.Version, req.Source.Repository)
	if err != nil {
		return nil, fmt.Errorf("saving version info: %v", err)
	}

	res := &InResponse{
		Version: req.Version,
	}
	return res, nil
}

func saveVersionInfo(dest string, version Version, repo string) error {
	err := ioutil.WriteFile(filepath.Join(dest, "tag"), []byte(version.Tag), 0644)
	if err != nil {
		return fmt.Errorf("write image tag: %w", err)
	}

	err = ioutil.WriteFile(filepath.Join(dest, "digest"), []byte(version.Digest), 0644)
	if err != nil {
		return fmt.Errorf("write image digest: %w", err)
	}

	err = ioutil.WriteFile(filepath.Join(dest, "repository"), []byte(repo), 0644)
	if err != nil {
		return fmt.Errorf("write image repository: %w", err)
	}

	return nil
}
