package resource

import (
	"os"
	"time"

	imgpkg "github.com/vmware-tanzu/carvel-imgpkg/pkg/imgpkg/cmd"
)

func writeCACertsToFile(certs []string) ([]string, error) {
	if len(certs) == 0 {
		return nil, nil
	}

	paths := make([]string, len(certs))
	for i, c := range certs {
		f, err := os.CreateTemp("", "")
		if err != nil {
			return nil, err
		}
		defer f.Close()

		f.WriteString(c)
		paths[i] = f.Name()
	}
	return paths, nil
}

func toRegistryFlags(source Source, certs []string) imgpkg.RegistryFlags {
	return imgpkg.RegistryFlags{
		CACertPaths: certs,
		VerifyCerts: true,
		Insecure:    source.Insecure,

		Username: source.Username,
		Password: source.Password,
		Anon:     false,

		RetryCount: 5,

		ResponseHeaderTimeout: 30 * time.Second,
	}
}
