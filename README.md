![godepgraph](./docs/logo.png)

**godepgraph** is a packages dependency graph visualization tool for your local go module project.

## install

```bash
go install github.com/alovn/godepgraph@latest
```

## How to use

### path

You need run **godepgraph** in your go module project path, or use the parameter **--path**:

```bash
godepgraph --path=/workspace/bytego

bytego/internal/bytebufferpool
bytego/internal/fasttemplate
└── bytego/internal/bytebufferpool
bytego/middleware/cors
└── bytego
bytego/middleware/logger
├── bytego
└── bytego/internal/fasttemplate
bytego/middleware/pprof
└── bytego
bytego/middleware/recovery
└── bytego
bytego
```

### pkg

If you want to display a specified package dependency, You need to know the parameter **--pkg** isn't the full pkg name, for examples the full pkg name is `github.com/gostack-labs/bytego/middleware/cors`, the **--pkg** parameter use a short pkg name: `bytego/middleware/cors`.

```bash
godepgraph --pkg=bytego/middleware/cors

bytego/middleware/cors
└── bytego
```

### reverse

List dependent packages, need parameter ***--pkg***.

```bash
godepgraph --pkg=bytego --reverse

bytego
├── bytego/middleware/cors
├── bytego/middleware/logger
├── bytego/middleware/pprof
└── bytego/middleware/recovery
```

### std and thrid

The standard pkg and third pkg of dependence is not displayed by default, if want display it:

```bash
godepgraph --std --third
```

### web

You can also start a local web server, and view the graph in a web browser, default listening ***localhost:7788***.

```bash
godepgraph --web
godepgraph --web --listen=:8080
```

### output

If you have the graphviz tools installed, parameter ***--dot*** can get a picture, default: godepgraph.png. you can specify an output file with  **--output**, supoort format:**jpg,png,svg,gif,dot**.

```bash
godepgraph --path=/workspace/bytego --dot
godepgraph --path=/workspace/bytego --dot --output=xx.svg
```

the picture of godepgraph.png like this:

![godepgraph](./docs/godepgraph.png)
