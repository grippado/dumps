# Contributing

Thanks for helping. The project is a bash CLI (`cli/dumps`), a Go binary (`server/`, module in `server/go.mod`) and two browser scripts (`web/`). It also ships the `dump-this` agent skill and command (`agents/`) and the project landing page (`site/`, published with GitHub Pages at https://grippado.github.io/dumps/).

## Before you open a PR

Run everything CI runs: gofmt, go vet, shellcheck and the Go tests with the race detector.

```bash
make check
```

You can run the halves separately.

```bash
make lint
```

```bash
make test
```

CI also cross-builds `dumpsd` for linux/arm64, linux/amd64, darwin/arm64 and darwin/amd64 with `CGO_ENABLED=0`. Keep it free of cgo and of dependencies outside the standard library.

## Goldens are the spec

The files in `server/testdata` (rendered `home.html`, `home-ga.html`, `listing.html` and the injected HTML under `inject/`) define the exact output of `dumpsd`. A diff in a golden is a behavior change.

When the change is intended, regenerate them deliberately and review the diff before committing.

```bash
cd server && DUMPSD_UPDATE_GOLDEN=1 go test ./...
```

Never update goldens just to make a failing test pass.

## Commits

Use Conventional Commits (`feat:`, `fix:`, `docs:`, `test:`, `refactor:`, `chore:`), with a scope when it helps, for example `fix(server): reject empty title`. Keep commits focused and explain the why in the body.

## Pull requests

- Describe what changed and why, and how you tested it.
- Add or update tests with the code. Auth, path and injection changes need tests.
- Update the docs in `docs/` when you change flags, config keys or behavior. Keep the English files and their Portuguese mirrors (`README.pt-BR.md`, `docs/pt-BR/`) in sync, section by section.
- Report security issues privately, see [SECURITY.md](SECURITY.md).
