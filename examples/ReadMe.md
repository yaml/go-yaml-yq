# examples

These examples are standalone command programs that load YAML with
`go.yaml.in/yaml/v4` into `*yaml.Node` values, evaluate expressions with
`github.com/yaml/go-yaml-yq`, and print YAML results.

Each example directory supports:

```bash
make build
make run
make clean
```

Examples:

```bash
make -C query run
make -C update run
make -C prompt run
make -C merge run
make -C color run
```

The Makefiles use the repository's local Makes/Go installation under
`../.cache`, so no system Go installation is required.
