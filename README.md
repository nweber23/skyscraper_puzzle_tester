# rush01-tester

[![Latest Release](https://img.shields.io/github/v/release/nweber23/skyscraper_puzzle_tester?label=release)](https://github.com/nweber23/skyscraper_puzzle_tester/releases/latest)

A TUI test harness for 42 School's **Rush01** (Skyscraper puzzle) assignment.
It fuzz-tests your compiled binary against randomly generated puzzles and a
battery of invalid-input cases, and checks the output with a self-contained
constraint validator — it does not contain or reveal any actual solver
logic, so it can't be used to shortcut the assignment itself.

## Why

Rush01 has no reference test suite, and hand-checking Skyscraper solutions
is tedious and error-prone once you're past the 4x4 mandatory grid. Because
a puzzle can have multiple valid solutions, this tool never compares your
output against one fixed "correct" grid — it re-derives whether *your*
output actually satisfies the given clues.

## Installation

**Option A — download a release binary**

```bash
# Linux amd64
curl -L https://github.com/nweber23/skyscraper_puzzle_tester/releases/latest/download/rush01-tester_linux_amd64 -o rush01-tester && chmod +x rush01-tester

# Linux arm64
curl -L https://github.com/nweber23/skyscraper_puzzle_tester/releases/latest/download/rush01-tester_linux_arm64 -o rush01-tester && chmod +x rush01-tester

# macOS amd64 (Intel)
curl -L https://github.com/nweber23/skyscraper_puzzle_tester/releases/latest/download/rush01-tester_darwin_amd64 -o rush01-tester && chmod +x rush01-tester

# macOS arm64 (Apple Silicon)
curl -L https://github.com/nweber23/skyscraper_puzzle_tester/releases/latest/download/rush01-tester_darwin_arm64 -o rush01-tester && chmod +x rush01-tester
```

**Option B — `go install`**

```bash
go install github.com/nweber23/skyscraper_puzzle_tester@latest
```

Note: `go install` names the binary after the module path
(`skyscraper_puzzle_tester`), not `rush01-tester`. Rename it after install
if you want the shorter name:

```bash
mv "$(go env GOPATH)/bin/skyscraper_puzzle_tester" "$(go env GOPATH)/bin/rush01-tester"
```

**Option C — build from source**

```bash
git clone https://github.com/nweber23/skyscraper_puzzle_tester.git
cd skyscraper_puzzle_tester
go build -o rush01-tester .
```

## Usage

```bash
./rush01-tester --bin=./ex00/rush01
```

```
rush01-tester

  > Mode: Mandatory (4x4)  (left/right to change)
    Size picker (disabled in mandatory mode)
    Fuzz runs: 50
    Start

up/down to move, enter to select, q to quit
```

While running:

```
Running tests

[████████████████████████████░░░░░░░░░░] 62%

31/50 - fuzz #31
```

Summary, with color-coded results (green PASS / red FAIL in a real
terminal):

```
Results: 47/50 passed

PASS fuzz #1
PASS fuzz #2
FAIL fuzz #3
...
PASS unsolvable view combination

Failure detail

case: fuzz #3
reason: grid failed constraint validation
  rowLeft[2]: expected 3, got 2

grid:
2 1 4 3
1 4 3 2
4 3 2 1
3 2 1 4

failure 1/3 - up/down to browse

m: menu, q: quit
```

## Flags

| Flag | Default | Description |
|---|---|---|
| `--bin` | `./a.out` | Path to the compiled Rush01 binary under test |
| `--size` | `4` | Grid size for bonus mode (4-9); ignored in mandatory mode |
| `--runs` | `50` | Number of fuzz test runs |
| `--timeout` | `5m` | Per-test-run timeout (Go duration syntax, e.g. `500ms`, `3s`, `5m`) — a run exceeding this is killed and auto-failed |
| `--mandatory` | `false` | Skip the menu and immediately run mandatory (4x4) mode |
| `--order` | `simple` | Clue order sent to the binary: `simple` (subject's worked example: colTop×4, colBottom×4, rowLeft×4, rowRight×4) or `clockwise` |
| `--version` | — | Print the tool version and exit |

## How validation works

For every generated puzzle, the tool independently computes what each of
the 4N clue values *should* be from the puzzle's known solution grid, feeds
those clues to your binary, then re-derives the same 4N visibility counts
from **your** output grid and compares them — along with checking every row
and column is a permutation of 1..N. It never compares your grid against
the original solution directly, since a valid Skyscraper puzzle can have
more than one solution. Invalid-input cases (malformed argument counts,
out-of-range values, non-numeric input, contradictory clues, etc.) are
checked separately: your binary is expected to reject them without
producing a valid-looking grid.

## License

MIT — see [`LICENSE`](LICENSE).

## Note

This is purely a testing tool. It does not solve Rush01 and must not be
submitted as (or mistaken for) part of your assignment — use it only to
validate your own solver's output, in keeping with the spirit of the 42
norm.
