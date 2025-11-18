# Letter Invaders (Go)

A modern Go implementation of the classic typing game Letter Invaders.

## About

This is a rewrite of the classic C version using Go and the [Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI framework. The original game had ~2000+ lines of C code with manual terminal handling - this version is ~300 lines of clean, modern Go.

## Features

- Type falling words before they reach the bottom
- Progressive difficulty with level increases
- Score tracking and WPM calculation
- Clean terminal UI with highlighted words
- Pause/resume functionality

## Installation

```bash
go build
```

## Usage

```bash
# Run with default dictionary
./letter-invaders-go

# Use a custom dictionary
./letter-invaders-go -d /path/to/dictionary.txt
```

## Controls

- **Type letters** - Match and destroy falling words
- **Backspace** - Clear current input
- **p** - Pause/resume game
- **Ctrl+L** - Redraw screen
- **q or Ctrl+C** - Quit

## Dictionary Format

The dictionary file should contain one word per line. Words are automatically filtered to 1-3 letters.

## Credits

Based on the original Letter Invaders by Larry Moss (1991)
- Original C version: https://github.com/davidcsterratt/letter-invaders
- Go rewrite using Bubble Tea framework
