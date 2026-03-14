# Installation

## npm (Node.js)

```sh
npm install @driftabot/engine
```

Downloads the correct pre-built binary for your platform automatically. See [npm SDK](/npm) for the full programmatic API.

## Homebrew (macOS / Linux)

```sh
brew tap DriftaBot/tap
brew install drift-agent
```

## Go install

```sh
go install github.com/DriftaBot/driftabot-engine/cmd/drift-agent@latest
```

## Build from source

```sh
go build -o drift-agent ./cmd/drift-agent
```

Or via Make:

```sh
make build
```
