package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	screenWidth  = 80
	screenHeight = 23
	statusHeight = 2
	gameHeight   = screenHeight - statusHeight
)

type word struct {
	text    string
	x, y    int
	matched int
}

type model struct {
	words      []word
	score      int
	level      int
	lives      int
	wordsTyped int
	dict       []string
	current    *word
	input      string
	gameOver   bool
	paused     bool
	startTime  time.Time
	width      int
	height     int
}

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func loadDictionary(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var words []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if len(word) > 0 {
			words = append(words, strings.ToLower(word))
		}
	}
	return words, scanner.Err()
}

func initialModel(dict []string) model {
	return model{
		words:     []word{},
		score:     0,
		level:     1,
		lives:     3,
		dict:      dict,
		startTime: time.Now(),
		width:     screenWidth,
		height:    screenHeight,
	}
}

func (m model) Init() tea.Cmd {
	return tickCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.gameOver {
			if msg.String() == "q" || msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "ctrl+l":
			return m, tea.ClearScreen
		case "ctrl+p":
			// Use Ctrl+P for pause to avoid conflicts with typing 'p'
			m.paused = !m.paused
			return m, nil
		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
			m.current = nil
			return m, nil
		default:
			// Handle letter input (including 'p')
			if len(msg.String()) == 1 && msg.String() >= "a" && msg.String() <= "z" {
				m.input += msg.String()
				m = m.matchWord()
				return m, nil
			}
		}

	case tickMsg:
		if !m.paused && !m.gameOver {
			m = m.moveWords()
			m = m.maybeAddWord()
		}
		return m, tickCmd()

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

	return m, nil
}

func (m model) matchWord() model {
	if len(m.input) == 0 {
		m.current = nil
		return m
	}

	// Try to find a word that matches the input
	for i := range m.words {
		w := &m.words[i]
		if strings.HasPrefix(w.text, m.input) {
			m.current = w
			w.matched = len(m.input)

			// Check if word is complete
			if m.input == w.text {
				m.score += len(w.text) * (m.level + 1)
				m.wordsTyped++
				m.words = append(m.words[:i], m.words[i+1:]...)
				m.input = ""
				m.current = nil

				// Level up every 15 words
				if m.wordsTyped%15 == 0 {
					m.level++
				}
			}
			return m
		}
	}

	// No match found - reset
	m.input = ""
	m.current = nil
	return m
}

func (m model) moveWords() model {
	for i := len(m.words) - 1; i >= 0; i-- {
		m.words[i].y++
		if m.words[i].y >= gameHeight {
			// Word reached bottom - lose a life
			m.words = append(m.words[:i], m.words[i+1:]...)
			m.lives--
			if m.lives <= 0 {
				m.gameOver = true
			}
		}
	}
	return m
}

func (m model) maybeAddWord() model {
	if len(m.words) >= 10 {
		return m
	}

	// Probability increases with level
	if rand.Float64() < 0.15+float64(m.level)*0.02 {
		newWord := m.dict[rand.Intn(len(m.dict))]
		maxX := screenWidth - len(newWord) - 1
		if maxX < 0 {
			maxX = 0
		}
		m.words = append(m.words, word{
			text: newWord,
			x:    rand.Intn(maxX + 1),
			y:    0,
		})
	}
	return m
}

func (m model) View() string {
	if m.gameOver {
		return m.renderGameOver()
	}

	// Create empty screen
	screen := make([][]rune, gameHeight)
	for i := range screen {
		screen[i] = make([]rune, screenWidth)
		for j := range screen[i] {
			screen[i][j] = ' '
		}
	}

	// Draw words
	highlightStyle := lipgloss.NewStyle().Background(lipgloss.Color("5")).Foreground(lipgloss.Color("15"))
	normalStyle := lipgloss.NewStyle()

	for _, w := range m.words {
		if w.y >= 0 && w.y < gameHeight {
			for i, ch := range w.text {
				if w.x+i < screenWidth {
					screen[w.y][w.x+i] = ch
				}
			}
		}
	}

	// Render screen to string
	var b strings.Builder
	b.WriteString("\n")
	for y := 0; y < gameHeight; y++ {
		line := string(screen[y])
		// Highlight current word if it's on this line
		if m.current != nil && m.current.y == y {
			before := line[:m.current.x]
			matched := m.current.text[:m.current.matched]
			unmatched := m.current.text[m.current.matched:]
			after := ""
			if m.current.x+len(m.current.text) < len(line) {
				after = line[m.current.x+len(m.current.text):]
			}
			line = before + highlightStyle.Render(matched) + normalStyle.Render(unmatched) + after
		}
		b.WriteString(line)
		b.WriteString("\n")
	}

	// Status line
	b.WriteString(strings.Repeat("─", screenWidth))
	b.WriteString("\n")
	elapsed := time.Since(m.startTime).Seconds()
	wpm := 0
	if elapsed > 0 {
		wpm = int(float64(m.wordsTyped) * 60.0 / elapsed)
	}
	status := fmt.Sprintf("Score: %d  Level: %d  Lives: %d  Words: %d  WPM: %d  Input: %s",
		m.score, m.level, m.lives, m.wordsTyped, wpm, m.input)
	b.WriteString(status)

	if m.paused {
		b.WriteString("\n\n[PAUSED - Press Ctrl+P to resume]")
	}

	b.WriteString("\n\n[ctrl+c/q: quit | ctrl+p: pause | ctrl+l: redraw]")

	return b.String()
}

func (m model) renderGameOver() string {
	var b strings.Builder
	b.WriteString("\n\n")
	b.WriteString(lipgloss.NewStyle().Bold(true).Render("GAME OVER"))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("Final Score: %d\n", m.score))
	b.WriteString(fmt.Sprintf("Level Reached: %d\n", m.level))
	b.WriteString(fmt.Sprintf("Words Typed: %d\n", m.wordsTyped))
	b.WriteString("\n\nPress 'q' to quit")
	return b.String()
}

func main() {
	dictPath := flag.String("d", "/usr/share/dict/words", "Path to dictionary file")
	flag.Parse()

	dict, err := loadDictionary(*dictPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading dictionary: %v\n", err)
		os.Exit(1)
	}

	if len(dict) == 0 {
		fmt.Fprintln(os.Stderr, "Dictionary is empty")
		os.Exit(1)
	}

	rand.Seed(time.Now().UnixNano())

	p := tea.NewProgram(initialModel(dict), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
