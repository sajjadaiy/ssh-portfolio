package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	wb "github.com/charmbracelet/wish/bubbletea"
	"github.com/muesli/termenv"
)

// --- PALETTE ---

var (
	colAccent = lipgloss.Color("#00D1FF") // neon cyan — hero color
	colFG     = lipgloss.Color("#E8EDF3") // cool off-white
	colFGDim  = lipgloss.Color("#8B9BB4") // steel gray
	colMuted  = lipgloss.Color("#4A5568") // dark steel
	colRule   = lipgloss.Color("#2D3748") // dividers
	colTech   = lipgloss.Color("#B794F4") // violet — tech stack
	colLink   = lipgloss.Color("#68D391") // mint — links / CTAs
	colTagBg  = lipgloss.Color("#1E293B") // tag pill background
)

// --- STYLES ---

var (
	logoStyle    = lipgloss.NewStyle().Foreground(colAccent).Bold(true)
	taglineStyle = lipgloss.NewStyle().Foreground(colMuted).Italic(true)
	sectionStyle = lipgloss.NewStyle().Foreground(colFGDim).Bold(true)
	ruleStyle    = lipgloss.NewStyle().Foreground(colRule)

	itemTitleNormal   = lipgloss.NewStyle().Foreground(colFGDim)
	itemTitleSelected = lipgloss.NewStyle().Foreground(colAccent).Bold(true)
	itemTechNormal    = lipgloss.NewStyle() // not shown for unselected
	itemTechSelected  = lipgloss.NewStyle().Foreground(colMuted)

	tagStyle = lipgloss.NewStyle().
			Foreground(colFGDim).
			Background(colTagBg).
			Padding(0, 1)

	detailTitleStyle = lipgloss.NewStyle().Foreground(colAccent).Bold(true)
	detailTechStyle  = lipgloss.NewStyle().Foreground(colTech).Italic(true)
	detailMetaStyle  = lipgloss.NewStyle().Foreground(colMuted)

	hintStyle   = lipgloss.NewStyle().Foreground(colMuted).Italic(true)
	linkStyle   = lipgloss.NewStyle().Foreground(colLink)
	splashStyle = lipgloss.NewStyle().Foreground(colAccent).Bold(true)
	cursorStyle = lipgloss.NewStyle().Foreground(colFGDim).Bold(true)

	socialIconStyle = lipgloss.NewStyle().Foreground(colAccent)
	socialTextStyle = lipgloss.NewStyle().Foreground(colFGDim)
)

// --- MODEL ---

const (
	ViewSplash = iota
	ViewList
	ViewDetail
	ViewHelp
)

type tickMsg time.Time
type blinkMsg struct{}

func tickCmd() tea.Cmd {
	return tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func blinkCmd() tea.Cmd {
	return tea.Tick(400*time.Millisecond, func(t time.Time) tea.Msg { return blinkMsg{} })
}

type model struct {
	cursor      int
	view        int
	width       int
	height      int
	splashText  string
	splashIndex int
	blinkCount  int
	showCursor  bool
	splashDone  bool
}

var splashFullText = "> establishing connection\n> loading systems\n> ready."

func initialModel() model {
	return model{showCursor: true}
}

func (m model) Init() tea.Cmd { return tickCmd() }

// --- RESPONSIVE ---

// tier returns 0=xs(<50) 1=sm(50-79) 2=md(80-119) 3=lg(120+)
func (m model) tier() int {
	switch {
	case m.width < 50:
		return 0
	case m.width < 80:
		return 1
	case m.width < 120:
		return 2
	default:
		return 3
	}
}

func (m model) contentWidth() int {
	w := m.width
	if w == 0 {
		w = 80
	}
	var cw int
	switch m.tier() {
	case 0:
		cw = w - 4
	case 1:
		cw = w - 8
	default:
		cw = w - 12
		if cw > 82 {
			cw = 82
		}
	}
	if cw < 24 {
		cw = 24
	}
	return cw
}

// --- UPDATE ---

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tickMsg:
		if m.view == ViewSplash && !m.splashDone {
			if m.splashIndex < len(splashFullText) {
				m.splashText += string(splashFullText[m.splashIndex])
				m.splashIndex++
				return m, tickCmd()
			}
			m.splashDone = true
			return m, blinkCmd()
		}

	case blinkMsg:
		if m.view == ViewSplash && m.splashDone {
			m.showCursor = !m.showCursor
			m.blinkCount++
			if m.blinkCount >= 6 {
				m.view = ViewList
				return m, nil
			}
			return m, blinkCmd()
		}

	case tea.KeyMsg:
		key := msg.String()

		if m.view == ViewSplash {
			m.view = ViewList
			return m, nil
		}

		switch key {
		case "ctrl+c":
			return m, tea.Quit

		case "q":
			if m.view == ViewList {
				return m, tea.Quit
			}
			m.view = ViewList

		case "up", "k":
			if m.view == ViewList && m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.view == ViewList && m.cursor < len(items)-1 {
				m.cursor++
			}

		case "tab":
			if m.view == ViewList {
				m.cursor = (m.cursor + 1) % len(items)
			}

		case "enter", " ":
			if m.view == ViewList {
				m.view = ViewDetail
			}

		case "esc", "backspace":
			if m.view != ViewList {
				m.view = ViewList
			}

		case "?":
			if m.view == ViewHelp {
				m.view = ViewList
			} else if m.view == ViewList {
				m.view = ViewHelp
			}
		}
	}
	return m, nil
}

// --- VIEW HELPERS ---

func centerText(s string, width int) string {
	lines := strings.Split(s, "\n")
	var out []string
	for _, line := range lines {
		lw := lipgloss.Width(line)
		if lw >= width {
			out = append(out, line)
			continue
		}
		pad := (width - lw) / 2
		out = append(out, strings.Repeat(" ", pad)+line)
	}
	return strings.Join(out, "\n")
}

func hyperlink(url, text string) string {
	return fmt.Sprintf("\x1b]8;;%s\x1b\\%s\x1b]8;;\x1b\\", url, text)
}

func sectionHeader(title string, cw int) string {
	titleStr := sectionStyle.Render(title)
	ruleW := cw - lipgloss.Width(titleStr) - 1
	if ruleW < 2 {
		ruleW = 2
	}
	return titleStr + ruleStyle.Render(" "+strings.Repeat("─", ruleW))
}

func (m model) logoText() string {
	switch m.tier() {
	case 0:
		return "SAJJAD AIYOOB"
	case 1:
		return ` ╔═╗┌─┐ ┬ ┬┌─┐┌┬┐
 ╚═╗├─┤ │ │├─┤ ││
 ╚═╝┴ ┴└┘└┘┴ ┴─┴┘
   ╔═╗┬┬ ┬┌─┐┌─┐┌┐
   ╠═╣│└┬┘│ ││ │├┴┐
   ╩ ╩┴ ┴ └─┘└─┘└─┘`
	default:
		return ` ███████╗ █████╗      ██╗     ██╗ █████╗ ██████╗
 ██╔════╝██╔══██╗     ██║     ██║██╔══██╗██╔══██╗
 ███████╗███████║     ██║     ██║███████║██║  ██║
 ╚════██║██╔══██║██   ██║██   ██║██╔══██║██║  ██║
 ███████║██║  ██║╚█████╔╝╚█████╔╝██║  ██║██████╔╝
 ╚══════╝╚═╝  ╚═╝ ╚════╝  ╚════╝ ╚═╝  ╚═╝╚═════╝
      █████╗ ██╗██╗   ██╗ ██████╗  ██████╗ ██████╗
     ██╔══██╗██║╚██╗ ██╔╝██╔═══██╗██╔═══██╗██╔══██╗
     ███████║██║ ╚████╔╝ ██║   ██║██║   ██║██████╔╝
     ██╔══██║██║  ╚██╔╝  ██║   ██║██║   ██║██╔══██╗
     ██║  ██║██║   ██║   ╚██████╔╝╚██████╔╝██████╔╝
     ╚═╝  ╚═╝╚═╝   ╚═╝    ╚═════╝  ╚═════╝ ╚═════╝`
	}
}

func (m model) hints() string {
	if m.view == ViewDetail || m.view == ViewHelp {
		return "esc  back  ·  q  quit"
	}
	return "↑↓  navigate  ·  ↵  open  ·  ?  help  ·  q  quit"
}

// --- RENDER SPLASH ---

func (m model) renderSplash() string {
	w, h := m.width, m.height
	if w == 0 {
		w = 80
	}
	if h == 0 {
		h = 24
	}

	var b strings.Builder
	b.WriteString(centerText(logoStyle.Render(m.logoText()), w))
	b.WriteString("\n\n")

	cursor := " "
	if m.showCursor {
		cursor = cursorStyle.Render("█")
	}
	b.WriteString(centerText(splashStyle.Render(m.splashText)+cursor, w))

	content := b.String()
	topPad := (h - strings.Count(content, "\n") - 1) / 3
	if topPad < 0 {
		topPad = 0
	}
	return strings.Repeat("\n", topPad) + content
}

// --- RENDER LIST ---

func (m model) renderItemRow(i int, item Item) string {
	selected := m.cursor == i

	var titleSty lipgloss.Style
	var cursor string
	if selected {
		titleSty = itemTitleSelected
		cursor = "›"
	} else {
		titleSty = itemTitleNormal
		cursor = " "
	}

	title := titleSty.Render(fmt.Sprintf("%s %s %s", cursor, item.Icon, item.Title))
	tag := tagStyle.Render(item.Tag)
	row := title + "  " + tag

	if selected && m.tier() >= 1 && item.TechStack != "" {
		tech := "     " + itemTechSelected.Render(item.TechStack)
		row += "\n" + tech
	}
	return row
}

func (m model) renderList(cw int) string {
	var b strings.Builder

	b.WriteString(sectionHeader("PROJECTS", cw) + "\n\n")
	for i, item := range items {
		if item.Category == "projects" {
			b.WriteString(m.renderItemRow(i, item) + "\n\n")
		}
	}

	b.WriteString(sectionHeader("ABOUT", cw) + "\n\n")
	for i, item := range items {
		if item.Category == "about" {
			b.WriteString(m.renderItemRow(i, item) + "\n\n")
		}
	}

	b.WriteString(sectionHeader("CONNECT", cw) + "\n\n")
	for _, s := range socials {
		// Compute padding manually — OSC8 sequences confound lipgloss.Width
		visibleW := lipgloss.Width(socialIconStyle.Render(s.Icon)) +
			2 + lipgloss.Width(socialTextStyle.Render(s.URL))
		lpad := (cw - visibleW) / 2
		if lpad < 0 {
			lpad = 0
		}
		icon := socialIconStyle.Render(s.Icon)
		text := hyperlink(s.Link, socialTextStyle.Render(s.URL))
		b.WriteString(strings.Repeat(" ", lpad) + icon + "  " + text + "\n")
	}

	return b.String()
}

// --- RENDER DETAIL ---

func (m model) renderDetail(cw int) string {
	item := items[m.cursor]

	var b strings.Builder

	b.WriteString(detailTitleStyle.Render(item.Icon+"  "+item.Title) + "\n")

	var meta []string
	if item.Year != "" {
		meta = append(meta, item.Year)
	}
	if item.Role != "" {
		meta = append(meta, item.Role)
	}
	if item.Tag != "" {
		meta = append(meta, item.Tag)
	}
	if len(meta) > 0 {
		b.WriteString("   " + detailMetaStyle.Render(strings.Join(meta, " · ")) + "\n")
	}

	b.WriteString("\n" + ruleStyle.Render(strings.Repeat("─", cw)) + "\n\n")

	if item.TechStack != "" {
		b.WriteString(detailTechStyle.Render(item.TechStack) + "\n")
	}

	if item.Repo != "" {
		if rd, ok := cache.get(item.Repo); ok {
			var parts []string
			if rd.Stars > 0 {
				parts = append(parts, fmt.Sprintf("★ %d", rd.Stars))
			}
			if rd.Language != "" {
				parts = append(parts, rd.Language)
			}
			if !rd.PushedAt.IsZero() {
				parts = append(parts, "updated "+relativeTime(rd.PushedAt))
			}
			if len(parts) > 0 {
				b.WriteString(detailMetaStyle.Render(strings.Join(parts, " · ")) + "\n")
			}
		}
	}

	b.WriteString("\n")

	descW := cw - 2
	if descW < 20 {
		descW = 20
	}
	desc := lipgloss.NewStyle().Foreground(colFG).Width(descW).Render(item.Description)
	b.WriteString(desc + "\n")

	if item.Link != "" {
		b.WriteString("\n")
		lnk := linkStyle.Render("→  " + item.Link)
		b.WriteString(hyperlink(item.Link, lnk) + "\n")
	}

	return b.String()
}

// --- RENDER HELP ---

func (m model) renderHelp(cw int) string {
	var b strings.Builder
	b.WriteString(sectionHeader("KEYBINDINGS", cw) + "\n\n")

	keys := []struct{ key, desc string }{
		{"↑ / k", "navigate up"},
		{"↓ / j", "navigate down"},
		{"↵ / space", "open detail"},
		{"esc / bksp", "go back"},
		{"tab", "cycle items"},
		{"?", "toggle this view"},
		{"q", "quit"},
	}
	for _, k := range keys {
		b.WriteString("  " + tagStyle.Render(k.key) + "   " + hintStyle.Render(k.desc) + "\n\n")
	}

	return b.String()
}

// --- VIEW ---

func (m model) View() string {
	if m.view == ViewSplash {
		return m.renderSplash()
	}

	cw := m.contentWidth()

	var b strings.Builder

	// Header
	b.WriteString(centerText(logoStyle.Render(m.logoText()), cw) + "\n")
	b.WriteString(centerText(taglineStyle.Render("Backend Developer  ·  Cloud  ·  Sri Lanka"), cw) + "\n\n")
	b.WriteString(ruleStyle.Render(strings.Repeat("─", cw)) + "\n\n")

	// Content
	switch m.view {
	case ViewList:
		b.WriteString(m.renderList(cw))
	case ViewDetail:
		b.WriteString(m.renderDetail(cw))
	case ViewHelp:
		b.WriteString(m.renderHelp(cw))
	}

	// Footer
	b.WriteString("\n" + ruleStyle.Render(strings.Repeat("─", cw)) + "\n")
	b.WriteString(centerText(hintStyle.Render(m.hints()), cw))

	content := b.String()

	w, h := m.width, m.height
	if w == 0 {
		w = 80
	}
	if h == 0 {
		h = 24
	}

	contentH := strings.Count(content, "\n") + 1
	topPad := 0
	if h > contentH+2 {
		topPad = (h - contentH) / 2
	}
	leftPad := 0
	if w > cw {
		leftPad = (w - cw) / 2
	}

	// Manual margin to avoid lipgloss width-detection issues with raw ANSI strings
	padStr := strings.Repeat(" ", leftPad)
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lines[i] = padStr + line
	}
	return strings.Repeat("\n", topPad) + strings.Join(lines, "\n")
}

func init() {
	lipgloss.SetColorProfile(termenv.TrueColor)
}

// --- SERVER ---

func main() {
	startGitHubRefresh()

	s, err := wish.NewServer(
		wish.WithAddress("0.0.0.0:23234"),
		wish.WithHostKeyPath(".ssh/term_info_ed25519"),
		wish.WithMiddleware(
			wb.Middleware(func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
				return initialModel(), []tea.ProgramOption{tea.WithAltScreen()}
			}),
		),
	)
	if err != nil {
		log.Fatalln(err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("Starting SSH server on port 23234...")

	go func() {
		if err = s.ListenAndServe(); err != nil {
			log.Fatalln(err)
		}
	}()

	<-done
	log.Println("Stopping server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		log.Fatalln(err)
	}
}
