# Build & Test

All commands run inside Docker. Use `make` targets instead of running `go` directly.

| Command            | Purpose                          |
|--------------------|----------------------------------|
| `make all`         | staticcheck + format + test + build |
| `make build`       | compile binary                   |
| `make test`        | run tests in `./test/...`        |
| `make staticcheck` | static analysis                  |
| `make format`      | gofmt                            |
| `make tidy`        | go mod tidy                      |
| `make shell`       | interactive container shell      |
