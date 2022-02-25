package resource

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
)

func RunCheck(req CheckRequest) (CheckResponse, error) {
	var regOpts []name.Option
	if req.Source.Insecure {
		regOpts = append(regOpts, name.Insecure)
	}

	repo, err := name.NewRepository(req.Source.Repository, regOpts...)
	if err != nil {
		return CheckResponse{}, fmt.Errorf("resolve repository: %w", err)
	}

	opts, err := authOptions(repo, req.Source)
	if err != nil {
		return CheckResponse{}, err
	}

	if req.Source.Tag != "" {
		return checkTag(repo.Tag(req.Source.Tag), req.Source, req.Version, opts...)
	} else {
		return checkRepository(repo, req.Source, req.Version, opts...)
	}
}

func authOptions(repo name.Repository, source Source) ([]remote.Option, error) {
	var auth authn.Authenticator
	if source.Username != "" && source.Password != "" {
		auth = &authn.Basic{
			Username: source.Username,
			Password: source.Password,
		}
	} else {
		auth = authn.Anonymous
	}

	tr := http.DefaultTransport.(*http.Transport)
	// a cert was provided
	if len(source.CaCerts) > 0 {
		rootCAs, err := x509.SystemCertPool()
		if err != nil {
			return nil, err
		}
		if rootCAs == nil {
			rootCAs = x509.NewCertPool()
		}

		for _, cert := range source.CaCerts {
			// append our cert to the system pool
			if ok := rootCAs.AppendCertsFromPEM([]byte(cert)); !ok {
				return nil, fmt.Errorf("failed to append registry certificate: %w", err)
			}
		}

		// trust the augmented cert pool in our client
		config := &tls.Config{
			RootCAs: rootCAs,
		}

		tr.TLSClientConfig = config
	}

	scopes := []string{repo.Scope(transport.PullScope)}

	rt, err := transport.New(repo.Registry, auth, tr, scopes)
	if err != nil {
		return nil, fmt.Errorf("initialize transport: %w", err)
	}

	return []remote.Option{remote.WithAuth(auth), remote.WithTransport(rt)}, nil
}

func checkRepository(repo name.Repository, source Source, from *Version, opts ...remote.Option) (CheckResponse, error) {
	tags, err := remote.List(repo, opts...)
	if err != nil {
		return CheckResponse{}, fmt.Errorf("list repository tags: %w", err)
	}

	bareTag := "latest"

	versionTags := map[*semver.Version]name.Tag{}
	tagDigests := map[string]string{}
	digestVersions := map[string]*semver.Version{}

	var cursorVer *semver.Version
	var latestTag string

	if from != nil {
		// assess the 'from' tag first so we can skip lower version numbers
		sort.Slice(tags, func(i, j int) bool {
			return tags[i] == from.Tag
		})
	}

	var constraint *semver.Constraints

	for _, identifier := range tags {
		var ver *semver.Version
		if identifier == bareTag {
			latestTag = identifier
		} else {
			verStr := identifier

			ver, err = semver.NewVersion(verStr)
			if err != nil {
				// not a version
				continue
			}

			if constraint != nil && !constraint.Check(ver) {
				// semver constraint not met
				continue
			}

			pre := ver.Prerelease()
			if pre != "" {
				// contains additional variant
				if strings.Contains(pre, "-") {
					continue
				}

				if !strings.HasPrefix(pre, "alpha") &&
					!strings.HasPrefix(pre, "beta") &&
					!strings.HasPrefix(pre, "rc") {
					// additional variant, not a prerelease segment
					continue
				}
			}

			if cursorVer != nil && (cursorVer.GreaterThan(ver) || cursorVer.Equal(ver)) {
				// optimization: don't bother fetching digests for lesser (or equal but
				// less specific, i.e. 6.3 vs 6.3.0) version tags
				continue
			}
		}

		tagRef := repo.Tag(identifier)

		digest, found, err := headOrGet(tagRef, opts...)
		if err != nil {
			return CheckResponse{}, fmt.Errorf("get tag digest: %w", err)
		}

		if !found {
			continue
		}

		tagDigests[identifier] = digest.String()

		if ver != nil {
			versionTags[ver] = tagRef

			existing, found := digestVersions[digest.String()]

			shouldSet := !found
			if found {
				if existing.Prerelease() == "" && ver.Prerelease() != "" {
					// favor final version over prereleases
					shouldSet = false
				} else if existing.Prerelease() != "" && ver.Prerelease() == "" {
					// favor final version over prereleases
					shouldSet = true
				} else if strings.Count(ver.Original(), ".") > strings.Count(existing.Original(), ".") {
					// favor more specific semver tag (i.e. 3.2.1 over 3.2, 1.0.0-rc.2 over 1.0.0-rc)
					shouldSet = true
				}
			}

			if shouldSet {
				digestVersions[digest.String()] = ver
			}
		}

		if from != nil && identifier == from.Tag && digest.String() == from.Digest {
			// if the 'from' version exists and has the same digest, treat its
			// version as a cursor in the tags, only considering newer versions
			//
			// note: the 'from' version will always be the first one hit by this loop
			cursorVer = ver
		}
	}

	var tagVersions TagVersions
	for digest, version := range digestVersions {
		tagVersions = append(tagVersions, TagVersion{
			TagName: versionTags[version].TagStr(),
			Digest:  digest,
			Version: version,
		})
	}

	sort.Sort(tagVersions)

	response := CheckResponse{}

	for _, ver := range tagVersions {
		response = append(response, Version{
			Tag:    ver.TagName,
			Digest: ver.Digest,
		})
	}

	if latestTag != "" {
		digest := tagDigests[latestTag]

		_, existsAsSemver := digestVersions[digest]
		if !existsAsSemver && constraint == nil {
			response = append(response, Version{
				Tag:    latestTag,
				Digest: digest,
			})
		}
	}

	return response, nil
}

type TagVersion struct {
	TagName string
	Digest  string
	Version *semver.Version
}

type TagVersions []TagVersion

func (vs TagVersions) Len() int           { return len(vs) }
func (vs TagVersions) Less(i, j int) bool { return vs[i].Version.LessThan(vs[j].Version) }
func (vs TagVersions) Swap(i, j int)      { vs[i], vs[j] = vs[j], vs[i] }

func checkTag(tag name.Tag, source Source, version *Version, opts ...remote.Option) (CheckResponse, error) {
	digest, found, err := headOrGet(tag, opts...)
	if err != nil {
		return CheckResponse{}, fmt.Errorf("get remote image: %w", err)
	}

	response := CheckResponse{}
	if version != nil && found && version.Digest != digest.String() {
		digestRef := tag.Repository.Digest(version.Digest)

		_, found, err := headOrGet(digestRef, opts...)
		if err != nil {
			return CheckResponse{}, fmt.Errorf("get remote image: %w", err)
		}

		if found {
			response = append(response, Version{
				Tag:    tag.TagStr(),
				Digest: version.Digest,
			})
		}
	}

	if found {
		response = append(response, Version{
			Tag:    tag.TagStr(),
			Digest: digest.String(),
		})
	}

	return response, nil
}

func headOrGet(ref name.Reference, imageOpts ...remote.Option) (v1.Hash, bool, error) {
	v1Desc, err := remote.Head(ref, imageOpts...)
	if err != nil {
		if checkMissingManifest(err) {
			return v1.Hash{}, false, nil
		}

		remoteDesc, err := remote.Get(ref, imageOpts...)
		if err != nil {
			if checkMissingManifest(err) {
				return v1.Hash{}, false, nil
			}

			return v1.Hash{}, false, err
		}

		return remoteDesc.Digest, true, nil
	}

	return v1Desc.Digest, true, nil
}

func checkMissingManifest(err error) bool {
	if rErr, ok := err.(*transport.Error); ok {
		return rErr.StatusCode == http.StatusNotFound
	}

	return false
}
