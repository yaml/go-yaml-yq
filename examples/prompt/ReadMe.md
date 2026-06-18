# prompt

Loads a YAML file into a `go.yaml.in/yaml/v4` `*yaml.Node`, then prompts on
standard input for yq expressions. Each expression is evaluated with `yq.Nodes`
against the original loaded document and printed as YAML. Type `quit` or `exit`
to leave the prompt.

Build:

```bash
make build
```

Run:

```bash
make run
./prompt sample.yaml
```

Example session:

```text
yq> .project
go-yaml-yq
yq> .settings.pure
true
yq> quit
```
