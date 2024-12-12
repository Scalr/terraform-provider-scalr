# skaff

A tool that helps to add new resources and data sources to the provider.
Generates the necessary boilerplate code from templates.

## Usage

The name for new resource or data source should be given in snake_case, lowercase,
as it will appear in Terraform configuration (without 'scalr_' provider prefix).

For example: `agent_pool`.

### From project's makefile

Makefile contains convenient targets to skaff:

```bash
make resource name=agent_pool
```

```bash
make datasource name=agent_pool
```

### Run the tool directly

```bash
cd skaff
go run cmd/main.go -type=resource -name=agent_pool
```

```bash
cd skaff
go run cmd/main.go -type=data_source -name=agent_pool
```

## What's next

Review the generated code carefully and make necessary adjustments.
The tool gives a minimal, simple but full-featured code to start with.
Templates may be updated and improved in the future.

Please note that the generated code is based on the terraform-provider-framework library only,
no SDKv2 imports should be used.

Don't forget to add new resource or data source to the provider's `Resources` or `DataSources` functions respectively.
