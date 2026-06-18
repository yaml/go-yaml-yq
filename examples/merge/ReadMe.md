# merge

Loads two YAML files into `go.yaml.in/yaml/v4` `*yaml.Node` values, merges them
with the yq multiply/merge expression `$1 * $2`, and prints the merged YAML copy.

Build:

```bash
make build
```

Run:

```bash
make run
make run FILE1=base.yaml FILE2=overlay.yaml
```

Direct usage:

```bash
./merge base.yaml overlay.yaml
```
