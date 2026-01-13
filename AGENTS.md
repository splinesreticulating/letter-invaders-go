# Repository Guidelines

## Project Structure & Module Organization
- `main.go` contains the entire game implementation (Bubble Tea model, input handling, rendering).
- `go.mod` and `go.sum` define Go module dependencies.
- Dictionary assets live at the repo root: `letters.txt` (default list) and `short_words.txt` (kid-friendly 1–3 letter words).
- There is no separate `internal/` or `pkg/` tree yet; keep new packages minimal and colocated when possible.

## Build, Test, and Development Commands
- `go build` — build the `letter-invaders-go` binary in the repo root.
- `./letter-invaders-go` — run the game using the system dictionary (`/usr/share/dict/words`).
- `./letter-invaders-go -d short_words.txt` — run with the bundled short-word dictionary.
- `go run .` — quick local run without creating a binary (uses the same flags).

## Coding Style & Naming Conventions
- Use standard Go formatting (`gofmt`), with tabs for indentation.
- Prefer concise, lowerCamelCase names for locals and UpperCamelCase for exported identifiers.
- Keep constants grouped by purpose, as in `main.go` (screen, gameplay, colors).
- Favor small helper functions over inlined logic when readability improves.

## Testing Guidelines
- No automated tests are currently present. If you add tests, use Go’s `testing` package and name files `*_test.go`.
- Suggested command: `go test ./...`.

## Commit & Pull Request Guidelines
- Commit messages follow short, imperative summaries (e.g., “Add configurable speed control…” or “Fix: Allow single letter words…”). Keep them focused and descriptive.
- PRs should include a brief description, the gameplay impact, and any new flags or controls added.
- If UI changes are made, include a short terminal screenshot or GIF where possible.

## Configuration & Gameplay Tips
- Dictionary files must be one word per line; words are normalized to lowercase and filtered to 1–12 characters.
- If you add new flags (e.g., `-speed`), document them in `README.md` and update the Controls or Usage section accordingly.
