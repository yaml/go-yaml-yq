# query

Loads a YAML file into a `go.yaml.in/yaml/v4` `*yaml.Node`, evaluates a yq
query expression with `yq.Nodes`, and prints every result as YAML.

Build:

```bash
make build
```

Run:

```bash
make run
make run EXPR='.data | keys | .[]'
./query sample.yaml '.metadata.labels.app'
```
