<p align="center">
  <img src="assets/screenshot-hero.png" alt="skit — Interactive script runner for package.json" width="820">
</p>

<p align="center">
  <a href="https://github.com/subut0n/skit/actions/workflows/ci.yml"><img src="https://github.com/subut0n/skit/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/subut0n/skit/releases/latest"><img src="https://img.shields.io/github/v/release/subut0n/skit?color=00b4ff" alt="Release"></a>
  <a href="LICENSE"><img src="https://img.shields.io/github/license/subut0n/skit?color=b44dff" alt="License"></a>
  <img src="https://img.shields.io/badge/deps-zero-00e88f" alt="Zero deps">
</p>

---

Stop memorizing script names. Stop opening `package.json`. Just type `skit` and pick.

```
skit              # interactive menu
skit test         # run directly
skit -w           # pick a workspace
```

It detects your runner automatically — bun, pnpm, yarn, or npm — from the lockfile in your project.

---

## Get it

```bash
curl -fsSL https://raw.githubusercontent.com/subut0n/skit/main/install.sh | sh
```

<details>
<summary>Other methods</summary>

```bash
# go install
go install github.com/subut0n/skit@latest

# from source (Go 1.25+)
git clone https://github.com/subut0n/skit.git && cd skit
make build && sudo mv skit /usr/local/bin/

# pre-built binary
# → https://github.com/subut0n/skit/releases
```

</details>

---

## How it works

### Interactive menu

<p align="center">
  <img src="assets/screenshot-menu.png" alt="Interactive menu" width="660">
</p>

Arrow keys to navigate, Enter to run, `/` to filter, `q` to quit.

### Filter as you type

<p align="center">
  <img src="assets/screenshot-filter.png" alt="Filter mode" width="660">
</p>

### Direct execution

<p align="center">
  <img src="assets/screenshot-direct.png" alt="Direct execution" width="660">
</p>

### Monorepo workspaces

<p align="center">
  <img src="assets/screenshot-workspace.png" alt="Workspace picker" width="660">
</p>

Works with npm, yarn, bun, and pnpm workspace configs.

### History

<p align="center">
  <img src="assets/screenshot-history.png" alt="Execution history" width="660">
</p>

Tracks what you ran, when, and where — across all your projects.

### All commands

<p align="center">
  <img src="assets/screenshot-help.png" alt="Help output" width="660">
</p>

---

## Runner detection

Lockfile in your project determines the runner:

| Lockfile | Runner |
|----------|--------|
| `bun.lockb` / `bun.lock` | `bun run` |
| `pnpm-lock.yaml` | `pnpm run` |
| `yarn.lock` | `yarn run` |
| `package-lock.json` | `npm run` |

No lockfile? Falls back to npm.

---

## Script descriptions

By default, skit shows the raw command. Add an `"x-skit"` field to your `package.json` for human-readable descriptions:

```json
{
  "scripts": {
    "dev": "next dev",
    "build": "next build",
    "test": "vitest run",
    "test:watch": "vitest"
  },
  "x-skit": {
    "dev": "Start dev server",
    "build": "Build for production",
    "test": "Run test suite",
    "test:watch": "Run tests in watch mode"
  }
}
```

Scripts with a `:` in their name are automatically grouped by prefix.

---

## Configuration

First launch triggers a setup wizard. Reconfigure anytime:

```bash
skit --config     # full setup
skit --lang       # language (en, fr, es, de)
skit --colors     # color scheme
skit --keys       # key bindings
```

**Color schemes** — Rainbow (default), Deuteranopia, Tritanopia, High Contrast

**Key schemes** — Arrows (default), WASD, or any two custom keys

Config lives in `~/.config/skit/config.json`.

---

## Build

Requires Go 1.25+.

```bash
make build       # build with version injection
make test        # tests + race detector
make dist        # cross-compile (linux, macOS, windows)
make coverage    # tests with coverage
```

## Project layout

```
internal/
  ansi/        ANSI escape codes
  config/      ~/.config/skit/ persistence
  detector/    lockfile → runner mapping
  history/     execution history (last 50)
  i18n/        translations (en, fr, es, de)
  parser/      package.json + workspace parsing
  ui/          raw-mode TUI + fallback menu
```

## License

[MIT](LICENSE)
