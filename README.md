# Terraform provider for Scalr
[![Build status](https://github.com/Scalr/terraform-provider-scalr/workflows/default/badge.svg)](https://github.com/Scalr/terraform-provider-scalr/actions) [![Coverage status](https://coveralls.io/repos/github/Scalr/terraform-provider-scalr/badge.svg?branch=develop)](https://coveralls.io/github/Scalr/terraform-provider-scalr?branch=develop)

The Scalr Terraform provider can be used to manage the components within the Scalr IaCP.
This will allow you to automate the creation of workspaces, variables, VCS providers and much more.

## Using the provider
### Requirements
- [Terraform](https://www.terraform.io/downloads.html) >= 0.12.x
Download the latest provider build for your OS and architecture
from the [releases page](https://github.com/Scalr/terraform-provider-scalr/releases)
that is compatible with your Scalr server version (under the "required" section).
Extract the archive to get the provider binary.

Follow the instructions on the [official documentation page](https://docs.scalr.com/en/latest/scalr-terraform-provider/index.html) to learn how to use it.
## Developing the provider
### Requirements
- [Terraform](https://www.terraform.io/downloads.html) >= 0.15.x
- [Go](https://golang.org/doc/install) >= 1.18
- [jq](https://stedolan.github.io/jq/) >= 1.0

### Setup
If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed
on your machine (version 1.18+ is *required*).

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

For development purposes you can make the provider use your local copy of go-scalr like this:
```sh
go mod edit -replace github.com/scalr/go-scalr=/Users/<username>/Projects/scalr/go-scalr # this should be your path
```
Remember to remove this link before committing:
```sh
go mod edit -dropreplace github.com/scalr/go-scalr
```

To update the go-scalr version:
```sh
go get github.com/scalr/go-scalr@develop
```

### Testing
#### Unit tests
```sh
$ make test
```
#### Acceptance tests
You will need to set up the environment variables for your Scalr installation. For example:
```sh
export SCALR_HOSTNAME=abcdef.scalr.com
export SCALR_TOKEN=.....
```

You can run the acceptance tests like this:
```sh
make testacc
```
If you want to run one or more specific tests you can pass the targets as an environment variable:
```sh
TESTARGS="-run TestAccScalrWorkspace_basic TestAccScalrWorkspace_update" make testacc
```

### CI

#### Acceptance tests

To run tests with the container from the current branch. You need to specify branch flags in the commit message.
Flags:
- `[API_BRANCH]` - whether to use current branch as API branch.

For example, commit message:
```
Commit title

Some description
[API_BRANCH]
```
