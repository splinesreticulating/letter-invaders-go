package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	// Screen dimensions
	screenWidth  = 80
	screenHeight = 23
	statusHeight = 2
	gameHeight   = screenHeight - statusHeight

	// Gameplay timing and difficulty
	defaultTickInterval = 1500 * time.Millisecond
	maxWordsOnScreen    = 8
	wordsPerLevel       = 15
	baseSpawnChance     = 0.08
	spawnChancePerLevel = 0.01

	// Word filtering
	minWordLength = 1
	maxWordLength = 12

	// Particle effects
	baseParticleCount   = 8
	particlesPerChar    = 2
	particleSpeedMin    = 0.5
	particleSpeedMax    = 2.0
	particleLifetimeMin = 3
	particleLifetimeMax = 6

	// Colors
	colorCyan      = "#00FFFF"
	colorDarkCyan  = "#00CED1"
	colorLightGrey = "#CCCCCC"
	colorDarkGrey  = "#888888"
	colorBlack     = "#000000"
)

var particleChars = []rune{'*', '+', '#', 'o', '.', '~', '^', 'x'}

var tickInterval = defaultTickInterval

// word represents a falling word in the game
type word struct {
	text    string // the actual word text
	x, y    int    // position on screen
	matched int    // number of characters matched by player
}

// particle represents a single particle in an explosion effect
type particle struct {
	x, y     float64 // position (floats for smooth animation)
	vx, vy   float64 // velocity vector
	char     rune    // display character
	lifetime int     // ticks remaining
}

// effect represents an explosion effect composed of multiple particles
type effect struct {
	particles []particle
}

// styles holds all lipgloss styles used for rendering
type styles struct {
	highlight lipgloss.Style
	word      lipgloss.Style
	separator lipgloss.Style
	status    lipgloss.Style
	pause     lipgloss.Style
	help      lipgloss.Style
	title     lipgloss.Style
	stats     lipgloss.Style
}

// model represents the game state
type model struct {
	words      []word
	effects    []effect
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
	styles     styles
}

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(tickInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// initStyles creates and returns all the lipgloss styles used in the game
func initStyles() styles {
	return styles{
		highlight: lipgloss.NewStyle().Background(lipgloss.Color(colorCyan)).Foreground(lipgloss.Color(colorBlack)).Bold(true),
		word:      lipgloss.NewStyle().Foreground(lipgloss.Color(colorDarkCyan)),
		separator: lipgloss.NewStyle().Foreground(lipgloss.Color(colorDarkCyan)),
		status:    lipgloss.NewStyle().Foreground(lipgloss.Color(colorLightGrey)),
		pause:     lipgloss.NewStyle().Foreground(lipgloss.Color(colorCyan)).Bold(true),
		help:      lipgloss.NewStyle().Foreground(lipgloss.Color(colorDarkGrey)),
		title:     lipgloss.NewStyle().Foreground(lipgloss.Color(colorCyan)).Bold(true),
		stats:     lipgloss.NewStyle().Foreground(lipgloss.Color(colorLightGrey)),
	}
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
		if len(word) >= minWordLength && len(word) <= maxWordLength {
			words = append(words, strings.ToLower(word))
		}
	}
	return words, scanner.Err()
}

func createExplosion(x, y int, wordLen int) effect {
	particles := []particle{}
	numParticles := baseParticleCount + wordLen*particlesPerChar

	for i := 0; i < numParticles; i++ {
		angle := float64(i) * 2.0 * math.Pi / float64(numParticles)
		speed := particleSpeedMin + rand.Float64()*(particleSpeedMax-particleSpeedMin)
		particles = append(particles, particle{
			x:        float64(x) + float64(i%wordLen),
			y:        float64(y),
			vx:       speed * math.Cos(angle),
			vy:       speed * math.Sin(angle),
			char:     particleChars[rand.Intn(len(particleChars))],
			lifetime: particleLifetimeMin + rand.Intn(particleLifetimeMax-particleLifetimeMin),
		})
	}

	return effect{particles: particles}
}

// Helper functions

// remove deletes an element from a slice at the given index
func remove[T any](slice []T, index int) []T {
	return append(slice[:index], slice[index+1:]...)
}

// isLetter checks if a string is a single letter a-z
func isLetter(s string) bool {
	return len(s) == 1 && s[0] >= 'a' && s[0] <= 'z'
}

// isInBounds checks if a position is within the game screen bounds
func isInBounds(x, y int) bool {
	return x >= 0 && x < screenWidth && y >= 0 && y < gameHeight
}

func initialModel(dict []string) model {
	return model{
		words:     []word{},
		effects:   []effect{},
		score:     0,
		level:     1,
		lives:     3,
		dict:      dict,
		startTime: time.Now(),
		width:     screenWidth,
		height:    screenHeight,
		styles:    initStyles(),
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
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+l":
			return m, tea.ClearScreen
		case " ":
			// Space bar pauses - can't conflict with typing words
			m.paused = !m.paused
			return m, nil
		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
			m.current = nil
			return m, nil
		default:
			// Handle letter input
			key := msg.String()
			if isLetter(key) {
				m.input += key
				m = m.matchWord()
				return m, nil
			}
		}

	case tickMsg:
		if !m.paused && !m.gameOver {
			m = m.moveWords()
			m = m.updateEffects()
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

// Game logic helper methods

// calculateScore returns the score for completing a word of given length
func (m model) calculateScore(wordLength int) int {
	return wordLength * (m.level + 1)
}

// calculateWPM returns the current words per minute
func (m model) calculateWPM() int {
	elapsed := time.Since(m.startTime).Seconds()
	if elapsed > 0 {
		return int(float64(m.wordsTyped) * 60.0 / elapsed)
	}
	return 0
}

// checkLevelUp increases the level if enough words have been typed
func (m model) checkLevelUp() model {
	if m.wordsTyped%wordsPerLevel == 0 {
		m.level++
	}
	return m
}

// shouldSpawnWord determines if a new word should be added this tick
func (m model) shouldSpawnWord() bool {
	if len(m.words) >= maxWordsOnScreen {
		return false
	}
	minWords := 1 + m.level/3
	if len(m.words) < minWords {
		return true
	}
	spawnChance := baseSpawnChance + float64(m.level)*spawnChancePerLevel
	return rand.Float64() < spawnChance
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
				m.score += m.calculateScore(len(w.text))
				m.wordsTyped++
				m = m.checkLevelUp()

				// Create explosion effect at word position
				m.effects = append(m.effects, createExplosion(w.x, w.y, len(w.text)))

				m.words = remove(m.words, i)
				m.input = ""
				m.current = nil
			}
			return m
		}
	}

	// No match found - reset
	m.input = ""
	m.current = nil
	return m
}

func (m model) updateEffects() model {
	// Update all particles in all effects
	for i := len(m.effects) - 1; i >= 0; i-- {
		effect := &m.effects[i]

		// Update each particle
		for j := len(effect.particles) - 1; j >= 0; j-- {
			p := &effect.particles[j]
			p.x += p.vx
			p.y += p.vy
			p.lifetime--

			// Remove dead particles
			if p.lifetime <= 0 {
				effect.particles = remove(effect.particles, j)
			}
		}

		// Remove effects with no particles left
		if len(effect.particles) == 0 {
			m.effects = remove(m.effects, i)
		}
	}
	return m
}

func (m model) moveWords() model {
	for i := len(m.words) - 1; i >= 0; i-- {
		m.words[i].y++
		if m.words[i].y >= gameHeight {
			// Word reached bottom - lose a life
			m.words = remove(m.words, i)
			m.lives--
			if m.lives <= 0 {
				m.gameOver = true
			}
		}
	}
	return m
}

func (m model) maybeAddWord() model {
	if m.shouldSpawnWord() {
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

	screen := m.initScreen()
	m.drawWords(screen)
	m.drawParticles(screen)
	return m.renderScreen(screen)
}

// initScreen creates an empty screen buffer
func (m model) initScreen() [][]rune {
	screen := make([][]rune, gameHeight)
	for i := range screen {
		screen[i] = make([]rune, screenWidth)
		for j := range screen[i] {
			screen[i][j] = ' '
		}
	}
	return screen
}

// drawWords renders all words to the screen buffer
func (m model) drawWords(screen [][]rune) {
	for _, w := range m.words {
		if w.y >= 0 && w.y < gameHeight {
			for i, ch := range w.text {
				if w.x+i < screenWidth {
					screen[w.y][w.x+i] = ch
				}
			}
		}
	}
}

// drawParticles renders all explosion particles to the screen buffer
func (m model) drawParticles(screen [][]rune) {
	for _, effect := range m.effects {
		for _, p := range effect.particles {
			px, py := int(p.x), int(p.y)
			if isInBounds(px, py) {
				screen[py][px] = p.char
			}
		}
	}
}

// renderScreen converts the screen buffer to a styled string
func (m model) renderScreen(screen [][]rune) string {
	var b strings.Builder
	b.WriteString("\n")

	m.renderGameArea(&b, screen)
	m.renderStatusLine(&b)
	m.renderHelp(&b)

	return b.String()
}

// renderGameArea renders the game play area with highlighting
func (m model) renderGameArea(b *strings.Builder, screen [][]rune) {
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
			line = before + m.styles.highlight.Render(matched) + m.styles.word.Render(unmatched) + after
		} else {
			line = m.styles.word.Render(line)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}
}

// renderStatusLine renders the status bar with score, level, etc.
func (m model) renderStatusLine(b *strings.Builder) {
	b.WriteString(m.styles.separator.Render(strings.Repeat("─", screenWidth)))
	b.WriteString("\n")

	status := fmt.Sprintf("Score: %d  Level: %d  Lives: %d  Words: %d  WPM: %d  Input: %s",
		m.score, m.level, m.lives, m.wordsTyped, m.calculateWPM(), m.input)
	b.WriteString(m.styles.status.Render(status))

	if m.paused {
		b.WriteString("\n\n" + m.styles.pause.Render("[PAUSED - Press SPACE to resume]"))
	}
}

// renderHelp renders the help text at the bottom
func (m model) renderHelp(b *strings.Builder) {
	b.WriteString("\n\n" + m.styles.help.Render("[ctrl+c: quit | SPACE: pause | ctrl+l: redraw]"))
}

func (m model) renderGameOver() string {
	var b strings.Builder
	b.WriteString("\n\n")
	b.WriteString(m.styles.title.Render("GAME OVER"))
	b.WriteString("\n\n")
	b.WriteString(m.styles.stats.Render(fmt.Sprintf("Final Score: %d\n", m.score)))
	b.WriteString(m.styles.stats.Render(fmt.Sprintf("Level Reached: %d\n", m.level)))
	b.WriteString(m.styles.stats.Render(fmt.Sprintf("Words Typed: %d\n", m.wordsTyped)))
	b.WriteString("\n\n" + m.styles.help.Render("Press 'q' to quit"))
	return b.String()
}

func main() {
	dictPath := flag.String("d", "/usr/share/dict/words", "Path to dictionary file")
	speed := flag.Int("speed", 1500, "Speed in milliseconds (higher = slower, default 1500)")
	flag.Parse()

	tickInterval = time.Duration(*speed) * time.Millisecond

	dict, err := loadDictionary(*dictPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading dictionary: %v\n", err)
		os.Exit(1)
	}

	if len(dict) == 0 {
		fmt.Fprintln(os.Stderr, "Dictionary is empty")
		os.Exit(1)
	}

	p := tea.NewProgram(initialModel(dict), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
