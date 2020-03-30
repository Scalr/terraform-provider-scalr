# Terraform provider for Scalr

- Website: https://www.scalr.com/

## Requirements

-	[Terraform](https://www.terraform.io/downloads.html) 0.10.x
-	[Go](https://golang.org/doc/install) 1.11 (to build the provider plugin)

## Building The Provider

Clone repository to: `$GOPATH/src/github.com/Scalr/terraform-provider-scalr`

```sh
$ mkdir -p $GOPATH/src/github.com/Scalr; cd $GOPATH/src/github.com/Scalr
$ git clone git@github.com:Scalr/terraform-provider-scalr
```

Enter the provider directory and build the provider

```sh
$ cd $GOPATH/src/github.com/Scalr/terraform-provider-scalr
$ make build
```

## Using the provider

If you're building the provider, follow the instructions to
[install it as a plugin.](https://www.terraform.io/docs/plugins/basics.html#installing-a-plugin)
After placing it into your plugins directory,  run `terraform init` to initialize it.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed
on your machine (version 1.11+ is *required*). You'll also need to correctly setup a
[GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`.

To compile the provider, run `make build`. This will build the provider and put the provider binary
in the `$GOPATH/bin` directory.

```sh
$ make build
...
$ $GOPATH/bin/terraform-provider-scalr
...
```

## Testing

Run unit tests
```sh
$ make test
```
