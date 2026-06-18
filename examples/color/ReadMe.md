# color

Loads a YAML file into a `go.yaml.in/yaml/v4` `*yaml.Node`, then renders it
with yq's YAML or JSON encoder with ANSI color enabled.

Build:

```bash
make build
```

Run:

```bash
make run
make run FORMAT=json
./color sample.yaml yaml
./color sample.yaml json
```
