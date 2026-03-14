# Installation

## npm (Node.js)

```sh
npm install @drift-agent/api-drift-engine
```

Downloads the correct pre-built binary for your platform automatically. See [npm SDK](/npm) for the full programmatic API.

## Homebrew (macOS / Linux)

```sh
brew tap DriftAgent/tap
brew install drift-guard
```

## Go install

```sh
go install github.com/DriftAgent/api-drift-engine/cmd/drift-guard@latest
```

## Build from source

```sh
go build -o drift-guard ./cmd/drift-guard
```

Or via Make:

```sh
make build
```
