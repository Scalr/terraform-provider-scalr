# Terraform provider for Scalr
[![Build status](https://github.com/Scalr/terraform-provider-scalr/workflows/default/badge.svg)](https://github.com/Scalr/terraform-provider-scalr/actions) [![Coverage status](https://coveralls.io/repos/github/Scalr/terraform-provider-scalr/badge.svg?branch=develop)](https://coveralls.io/github/Scalr/terraform-provider-scalr?branch=develop)

The Scalr Terraform provider is used to manage resources within [Scalr](https://scalr.com), a cost-effective, drop-in replacement for Terraform Cloud with feature parity and better GitOps support. This provider allows you to automate the creation of workspaces, variables, VCS providers, and much more.

## Using the provider
### Requirements
- [Terraform](https://www.terraform.io/downloads.html) >= 0.12.x
Download the latest provider build for your OS and architecture
from the [releases page](https://github.com/Scalr/terraform-provider-scalr/releases).
Extract the archive to get the provider binary.

Follow the instructions on the [official documentation page](https://docs.scalr.io/docs/provider_overview) to learn how to use it.
## Developing the provider
### Requirements
- [Terraform](https://www.terraform.io/downloads.html) >= 0.15.x
- [Go](https://golang.org/doc/install) >= 1.19
- [jq](https://stedolan.github.io/jq/) >= 1.0

### Setup
If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed
on your machine (version 1.19+ is *required*).

Clone the repository:
```sh
$ git clone git@github.com:Scalr/terraform-provider-scalr
```

Enter the provider directory and build the provider:
```sh
$ cd $GOPATH/src/github.com/Scalr/terraform-provider-scalr
$ make build
```
If you are on macOS and wish to cross-compile the provider for GNU/Linux you can use `make build-linux` instead.
Note that the behaviour of the linker [has changed in Go 1.15](https://golang.org/doc/go1.15#linker).
If you are not using the makefile to build you will need to update your build flags to match it.

You should have the `terraform-provider-scalr` binary in your current working directory.

### Running

Tag current commit with some valid semantic version (`7.7.7` for example) or use `VER` variable as an argument for `make`.
Create a terraform configuration file (`main.tf`) with following content:
```
provider scalr {}

terraform {
    required_providers {
        scalr = {
            source = "registry.scalr.io/scalr/scalr"
            version = "7.7.7"
        }
    }
}
```
Export `SCALR_HOSTNAME` and `SCALR_TOKEN` environment variables.
Execute `VER=7.7.7 make install`. This will build the provider binary and install it to the user's provider directory (see `GNUMakefile`).
The `terraform init` now will find the correct provider version.

### Debugging provider

Start provider process by your IDE or debugger with `-debug` flag set.

```shell
./terraform-provider-scalr -debug=true
```

It will print output like the following:

```text
Provider started, to attach Terraform set the TF_REATTACH_PROVIDERS env var:

        TF_REATTACH_PROVIDERS='{"registry.scalr.io/scalr/scalr":{"Protocol":"grpc","Pid":30636,"Test":true,"Addr":{"Network":"unix","String":"/var/folders/8g/knswsjzs623b63ln4m9wxf180000gn/T/plugin4288941537"}}}'
```

Either export it, or prefix every Terraform command with it.
- Terraform will not start the provider process; instead, it will attach to your process
- the provider will no longer be restarted once per walk of the Terraform graph; instead the same provider process will be reused until the command is completed

For more info refer to the [documentation](https://developer.hashicorp.com/terraform/plugin/debugging#debugger-based-debugging).

### Using a local copy of go-scalr
This provider uses [go-scalr](https://github.com/Scalr/go-scalr) to call the Scalr API.
New code uses the v2 client (`github.com/scalr/go-scalr/v2`), generated from the Scalr OpenAPI spec;
the v1 client is frozen. To add a missing API to v2, see [CONTRIBUTING.md](CONTRIBUTING.md#how-to-contribute-to-the-scalr-provider).

For development purposes you can make the provider use your local copy of go-scalr v2
(with `go-scalr` checked out next to this repository):
```sh
go mod edit -replace github.com/scalr/go-scalr/v2=../go-scalr/v2
```
Remove this link before pushing:
```sh
go mod edit -dropreplace github.com/scalr/go-scalr/v2 && go mod tidy
```

To update the go-scalr v2 version, pin a merged commit:
```sh
go get github.com/scalr/go-scalr/v2@<commit-sha> && go mod tidy
```

### Testing
#### Unit tests
```sh
$ make test
```
#### Acceptance tests
Acceptance tests create real objects. Run them against a disposable test environment with the fatmouse seed data, never production.

Put the test environment settings into `.scalr-testenv.env` (git-ignored; AI agents use it too):
```sh
cp .scalr-testenv.env.sample .scalr-testenv.env
chmod 600 .scalr-testenv.env
# set SCALR_HOSTNAME and SCALR_TOKEN, and optional credentials for tests that need them
```

Run selected tests by a name regexp:
```sh
make testacc-one RUN='^TestAccScalrWorkspace_(basic|update)$'
```

The full suite (`make testacc`) takes up to 45 minutes and reads `SCALR_HOSTNAME` and `SCALR_TOKEN` from the environment; CI runs it on every PR.

### CI

#### Acceptance tests

The required parameters for creating a container will be determined automatically based on existing branches and pull requests in other repositories. No more need to add `[API_BRANCH]`
