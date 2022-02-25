# imgpkg Resource

Concourse [resource][concourse-resource] for pulling and pushing
[imgpkg][imgpkg] bundles

## Source configuration

- `repository`: *Required* The name of the repository, e.g.
  `dkalinin/test-content`

- `tag`: *Optional* A tag to monitor or push bundles to, e.g. `v1.0.1`

- `username` and `password`: *Optional* A username and password to use when
  authenticating to the registry.

- `insecure`: *Optional* Allow the use of http when interacting with registries

- `ca_certs`: *Optional* An array of PEM-encoded CA certificates:

  ```yaml
  ca_certs:
  - |
    -----BEGIN CERTIFICATE-----
    ...
    -----END CERTIFICATE-----
  - |
    -----BEGIN CERTIFICATE-----
    ...
    -----END CERTIFICATE-----
  ```


## Behaviour

### `check`

Same behaviour as the the Concourse [registry-image-resource][registry-image-resource]:

- If `tag` is configured, then the current digest for that tag

- Otherwise, the latest semver tag

### `get`

Pulls the bundle specified by the digest.

Files populated:

- `./output/`: A directory containing the contents of the bundle.

- `./repository`: A file containing the image's full repository name, e.g.
  `dkalinin/test-content`

- `./tag`: A file containing the tag from the version.

- `./digest`: A file containing the digest from the version, e.g. `sha256:...`.


### `put`

Push the contents of the directory to a bundle.

Parameters:

- `tag`: Tag for the pushed bundle, overwrites the tag specified in the source

- `path`: Path to the directory that will be pushed.

[concourse-resource]: https://concourse-ci.org/resource-types.html
[imgpkg]: https://carvel.dev/imgpkg/
[registry-image-resource]: https://github.com/concourse/registry-image-resource#check-with-tag-discover-new-digests-for-the-tag
