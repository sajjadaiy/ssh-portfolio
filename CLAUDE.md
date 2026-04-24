# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

An SSH-served, terminal-based interactive portfolio built with Go. Visitors connect via `ssh sajjad.tech` and get a Bubble Tea TUI rendered over Wish (Charmbracelet's SSH framework). The entire application lives in a single file: `main.go`.

## Commands

```bash
# Run locally (requires .ssh/term_info_ed25519 host key to exist or be generated)
go run main.go

# Build binary
go build -o portfolio main.go

# Build and run Docker container
docker build -t portfolio .
docker run -p 23234:23234 portfolio

# Connect to test (once server is running)
ssh -p 23234 localhost
```

The server binds to `0.0.0.0:23234`. On first run, Wish auto-generates the host key at `.ssh/term_info_ed25519`.

## Architecture

Everything is in `main.go`, organized in six clearly labeled sections:

1. **DATA & CONTENT** — `Item` and `Social` structs, inline content slices (`items`, `socials`), easter egg quotes/hints
2. **STYLES** — All Lip Gloss styles as package-level vars; `init()` forces TrueColor via `lipgloss.SetColorProfile(termenv.TrueColor)`
3. **MODEL** — `model` struct holds all UI state; five views (`ViewSplash`, `ViewList`, `ViewDetail`, `ViewMatrix`, `ViewHelp`) as iota constants
4. **UPDATE** — Bubble Tea `Update()` handles window resize, tick animations, key input, Konami code detection, typed-buffer easter eggs
5. **VIEW** — `View()` dispatches to `renderSplash()`, `renderMatrix()`, or the inline list/detail/help renderer; responsive layout adapts at `contentWidth < 60` and `width < 50` breakpoints
6. **SERVER** — `main()` starts a Wish SSH server; the Bubble Tea middleware wires each SSH session to a fresh `initialModel()`

**Content updates**: Add/edit entries directly in the `items` slice at the top of `main.go`. No database or config files.

**Deployment**: Hosted on a local server tunneled through AWS EC2 (ISP routing restriction). Docker is the deployment artifact.

## Key Constraints

- TrueColor is force-set in `init()` because SSH clients often don't advertise `COLORTERM=truecolor` — do not remove this.
- OSC 8 hyperlinks are rendered for social links but may not display in all SSH clients; the visible URL is always shown as fallback.
- `CGO_ENABLED=0` in the Dockerfile is required for a static binary on Alpine.
