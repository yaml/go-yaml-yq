# update

Loads a YAML file into a `go.yaml.in/yaml/v4` `*yaml.Node`, evaluates a yq
update expression with `yq.Node`, and prints the updated YAML copy.

Build:

```bash
make build
```

Run:

```bash
make run
make run EXPR='.env.LOG_LEVEL = "debug"'
./update sample.yaml '.replicas = 3'
```
