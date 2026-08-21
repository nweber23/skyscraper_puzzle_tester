package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// version is overwritten at build time via -ldflags "-X main.version=...".
var version = "dev"

type screen int

const (
	screenMenu screen = iota
	screenSizePicker
	screenRunsInput
	screenRunning
	screenSummary
)

type menuItem int

const (
	itemMode menuItem = iota
	itemSize
	itemRuns
	itemStart
	itemCount
)

var (
	styleTitle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	styleCursor = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	stylePass   = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	styleFail   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	styleDim    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

// config holds the run-wide settings supplied via CLI flags.
type config struct {
	binPath string
	timeout time.Duration
	order   ViewOrder
}

// model is the bubbletea model driving the whole TUI.
type model struct {
	cfg config

	screen screen
	cursor int

	mandatory bool // true = fixed 4x4, false = bonus (selectable size)
	size      int
	runs      int
	runsInput string

	queue    []TestCase
	results  []TestResult
	index    int
	progress progress.Model

	selectedFailure int
}

func initialModel(cfg config) model {
	return model{
		cfg:       cfg,
		screen:    screenMenu,
		mandatory: true,
		size:      4,
		runs:      50,
		runsInput: "50",
		progress:  progress.New(progress.WithDefaultGradient()),
	}
}

func (m model) Init() tea.Cmd {
	if m.screen == screenRunning {
		return runNextTestCmd(m)
	}
	return nil
}

type testDoneMsg struct {
	result TestResult
}

func runNextTestCmd(m model) tea.Cmd {
	tc := m.queue[m.index]
	cfg := m.cfg
	return func() tea.Msg {
		res := RunBinary(cfg.binPath, tc.Args, cfg.timeout)
		return testDoneMsg{result: EvaluateResult(tc, res)}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.screen {
		case screenMenu:
			return m.updateMenu(msg)
		case screenSizePicker:
			return m.updateSizePicker(msg)
		case screenRunsInput:
			return m.updateRunsInput(msg)
		case screenSummary:
			return m.updateSummary(msg)
		}
	case testDoneMsg:
		m.results = append(m.results, msg.result)
		m.index++
		if m.index >= len(m.queue) {
			m.screen = screenSummary
			return m, nil
		}
		cmd := m.progress.SetPercent(float64(m.index) / float64(len(m.queue)))
		return m, tea.Batch(cmd, runNextTestCmd(m))
	case progress.FrameMsg:
		newModel, cmd := m.progress.Update(msg)
		m.progress = newModel.(progress.Model)
		return m, cmd
	}
	return m, nil
}

func (m model) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < int(itemCount)-1 {
			m.cursor++
		}
	case "left", "h":
		if menuItem(m.cursor) == itemMode {
			m.mandatory = true
		}
	case "right", "l":
		if menuItem(m.cursor) == itemMode {
			m.mandatory = false
			if m.size < 4 {
				m.size = 4
			}
		}
	case "enter":
		switch menuItem(m.cursor) {
		case itemSize:
			if !m.mandatory {
				m.screen = screenSizePicker
			}
		case itemRuns:
			m.screen = screenRunsInput
		case itemStart:
			return m.startRun()
		}
	}
	return m, nil
}

func (m model) updateSizePicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "enter":
		m.screen = screenMenu
	case "up", "k":
		if m.size < 9 {
			m.size++
		}
	case "down", "j":
		if m.size > 4 {
			m.size--
		}
	}
	return m, nil
}

func (m model) updateRunsInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.screen = screenMenu
	case "enter":
		if v, err := strconv.Atoi(m.runsInput); err == nil && v > 0 {
			m.runs = v
		}
		m.screen = screenMenu
	case "backspace":
		if len(m.runsInput) > 0 {
			m.runsInput = m.runsInput[:len(m.runsInput)-1]
		}
	default:
		s := msg.String()
		if len(s) == 1 && s[0] >= '0' && s[0] <= '9' {
			m.runsInput += s
		}
	}
	return m, nil
}

func (m model) startRun() (tea.Model, tea.Cmd) {
	n := 4
	if !m.mandatory {
		n = m.size
	}
	m.queue = BuildTestQueue(n, m.runs, m.cfg.order)
	m.results = nil
	m.index = 0
	m.selectedFailure = 0
	m.screen = screenRunning
	m.progress = progress.New(progress.WithDefaultGradient())
	return m, runNextTestCmd(m)
}

func (m model) updateSummary(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up", "k":
		if m.selectedFailure > 0 {
			m.selectedFailure--
		}
	case "down", "j":
		if m.selectedFailure < len(m.failedResults())-1 {
			m.selectedFailure++
		}
	case "m":
		m.screen = screenMenu
		m.cursor = 0
	}
	return m, nil
}

// countLabel renders "passed/total" for one category, styled green when
// everything passed and red when at least one case in the category failed.
func countLabel(passed, total int) string {
	label := fmt.Sprintf("%d/%d", passed, total)
	if total == 0 {
		return styleDim.Render(label)
	}
	if passed == total {
		return stylePass.Render(label)
	}
	return styleFail.Render(label)
}

func (m model) failedResults() []TestResult {
	var out []TestResult
	for _, r := range m.results {
		if !r.Passed {
			out = append(out, r)
		}
	}
	return out
}

func (m model) View() string {
	switch m.screen {
	case screenMenu:
		return m.viewMenu()
	case screenSizePicker:
		return m.viewSizePicker()
	case screenRunsInput:
		return m.viewRunsInput()
	case screenRunning:
		return m.viewRunning()
	case screenSummary:
		return m.viewSummary()
	}
	return ""
}

func (m model) viewMenu() string {
	var b strings.Builder
	b.WriteString(styleTitle.Render("rush01-tester") + "\n\n")

	modeLabel := "Mandatory (4x4)"
	if !m.mandatory {
		modeLabel = fmt.Sprintf("Bonus (size %d)", m.size)
	}

	sizeHint := ""
	if m.mandatory {
		sizeHint = styleDim.Render(" (disabled in mandatory mode)")
	}

	items := []string{
		"Mode: " + modeLabel + "  (left/right to change)",
		"Size picker" + sizeHint,
		fmt.Sprintf("Fuzz runs: %d", m.runs),
		"Start",
	}

	for i, item := range items {
		cursor := "  "
		if i == m.cursor {
			cursor = styleCursor.Render("> ")
		}
		b.WriteString(cursor + item + "\n")
	}

	b.WriteString("\n" + styleDim.Render("up/down to move, enter to select, q to quit"))
	return b.String()
}

func (m model) viewSizePicker() string {
	var b strings.Builder
	b.WriteString(styleTitle.Render("Bonus grid size") + "\n\n")
	for n := 4; n <= 9; n++ {
		cursor := "  "
		if n == m.size {
			cursor = styleCursor.Render("> ")
		}
		b.WriteString(fmt.Sprintf("%s%dx%d\n", cursor, n, n))
	}
	b.WriteString("\n" + styleDim.Render("up/down to change, enter/esc to confirm"))
	return b.String()
}

func (m model) viewRunsInput() string {
	return styleTitle.Render("Fuzz run count") + "\n\n" +
		"> " + m.runsInput + "\n\n" +
		styleDim.Render("digits only, enter to confirm, esc to cancel")
}

func (m model) viewRunning() string {
	name := ""
	if m.index < len(m.queue) {
		name = m.queue[m.index].Name
	}
	return styleTitle.Render("Running tests") + "\n\n" +
		m.progress.View() + "\n\n" +
		fmt.Sprintf("%d/%d - %s", m.index, len(m.queue), name)
}

func (m model) viewSummary() string {
	var b strings.Builder

	var fuzzPassed, fuzzTotal, errPassed, errTotal int
	for _, r := range m.results {
		if r.Case.IsFuzz {
			fuzzTotal++
			if r.Passed {
				fuzzPassed++
			}
		} else {
			errTotal++
			if r.Passed {
				errPassed++
			}
		}
	}
	passed := fuzzPassed + errPassed

	b.WriteString(styleTitle.Render(fmt.Sprintf("Results: %d/%d passed", passed, len(m.results))) + "\n")
	b.WriteString(fmt.Sprintf("  %s puzzle solving\n", countLabel(fuzzPassed, fuzzTotal)))
	b.WriteString(fmt.Sprintf("  %s error handling\n", countLabel(errPassed, errTotal)))
	b.WriteString("\n")

	for _, r := range m.results {
		mark := stylePass.Render("PASS")
		if !r.Passed {
			mark = styleFail.Render("FAIL")
		}
		b.WriteString(fmt.Sprintf("%s %s\n", mark, r.Case.Name))
	}

	failures := m.failedResults()
	if len(failures) > 0 {
		b.WriteString("\n" + styleTitle.Render("Failure detail") + "\n\n")
		fr := failures[m.selectedFailure]
		b.WriteString(renderFailure(fr))
		b.WriteString(fmt.Sprintf("\n%s", styleDim.Render(fmt.Sprintf("failure %d/%d - up/down to browse", m.selectedFailure+1, len(failures)))))
	}

	b.WriteString("\n\n" + styleDim.Render("m: menu, q: quit"))
	return b.String()
}

func renderFailure(r TestResult) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("case: %s\n", r.Case.Name))
	b.WriteString(fmt.Sprintf("reason: %s\n", r.Reason))

	if len(r.PermErrors) > 0 {
		b.WriteString("permutation errors:\n")
		for _, e := range r.PermErrors {
			b.WriteString("  - " + e + "\n")
		}
	}

	for _, mm := range r.Mismatches {
		b.WriteString(fmt.Sprintf("  %s[%d]: expected %d, got %d\n", mm.Direction, mm.Index, mm.Expected, mm.Actual))
	}

	if r.ParsedGrid != nil {
		if r.Case.IsFuzz && r.Case.ColTop != nil {
			b.WriteString("\ngrid (with expected top/bottom/left/right clues):\n" +
				renderAnnotatedGrid(r.ParsedGrid, r.Case.ColTop, r.Case.ColBottom, r.Case.RowLeft, r.Case.RowRight))
		} else {
			b.WriteString("\ngrid:\n" + renderGrid(r.ParsedGrid))
		}
	} else if r.Stdout != "" {
		b.WriteString("\nraw output:\n" + r.Stdout)
	}

	return b.String()
}

func renderGrid(grid [][]int) string {
	var b strings.Builder
	for _, row := range grid {
		cells := make([]string, len(row))
		for i, v := range row {
			cells[i] = strconv.Itoa(v)
		}
		b.WriteString(strings.Join(cells, " ") + "\n")
	}
	return b.String()
}

// renderAnnotatedGrid renders grid surrounded by its expected clue values
// (colTop above, colBottom below, rowLeft/rowRight beside each row) so a
// mismatch can be counter-checked by hand against the subject's appendix
// layout.
func renderAnnotatedGrid(grid [][]int, colTop, colBottom, rowLeft, rowRight []int) string {
	n := len(grid)
	pad := strings.Repeat(" ", 2)

	cols := func(vals []int) string {
		cells := make([]string, n)
		for i, v := range vals {
			cells[i] = strconv.Itoa(v)
		}
		return strings.Join(cells, " ")
	}

	border := "+" + strings.Repeat("-", 2*n+1) + "+"

	var b strings.Builder
	b.WriteString(pad + cols(colTop) + "\n")
	b.WriteString(pad + border + "\n")
	for i, row := range grid {
		cells := make([]string, n)
		for j, v := range row {
			cells[j] = strconv.Itoa(v)
		}
		b.WriteString(fmt.Sprintf("%d | %s | %d\n", rowLeft[i], strings.Join(cells, " "), rowRight[i]))
	}
	b.WriteString(pad + border + "\n")
	b.WriteString(pad + cols(colBottom) + "\n")
	return b.String()
}

func main() {
	binPath := flag.String("bin", "./a.out", "path to the compiled Rush01 binary to test")
	size := flag.Int("size", 4, "grid size for bonus mode (4-9); ignored in mandatory mode")
	runs := flag.Int("runs", 50, "number of fuzz test runs")
	timeout := flag.Duration("timeout", 5*time.Minute, "per-test-run timeout")
	mandatoryFlag := flag.Bool("mandatory", false, "skip the menu and run mandatory (4x4) mode immediately")
	orderFlag := flag.String("order", string(OrderSimple), "clue order sent to the binary: clockwise or simple")
	showVersion := flag.Bool("version", false, "print the tool version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println("rush01-tester", version)
		return
	}

	if *size < 4 || *size > 9 {
		fmt.Fprintln(os.Stderr, "--size must be between 4 and 9")
		os.Exit(1)
	}
	if *runs <= 0 {
		fmt.Fprintln(os.Stderr, "--runs must be positive")
		os.Exit(1)
	}

	order := ViewOrder(*orderFlag)
	if order != OrderClockwise && order != OrderSimple {
		fmt.Fprintln(os.Stderr, "--order must be 'clockwise' or 'simple'")
		os.Exit(1)
	}

	cfg := config{binPath: *binPath, timeout: *timeout, order: order}
	m := initialModel(cfg)
	m.size = *size
	m.runs = *runs
	m.runsInput = strconv.Itoa(*runs)

	if *mandatoryFlag {
		m.mandatory = true
		started, _ := m.startRun()
		m = started.(model)
	}

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error running TUI:", err)
		os.Exit(1)
	}
}
