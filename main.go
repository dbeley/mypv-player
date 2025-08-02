package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Config stores subscription URLs.
type Config struct {
	Subscriptions []string `json:"subscriptions"`
}

func configPath() string {
	usr, _ := user.Current()
	return filepath.Join(usr.HomeDir, ".config", "mypv-player", "config.json")
}

func loadConfig() Config {
	path := configPath()
	var cfg Config
	f, err := os.Open(path)
	if err == nil {
		defer f.Close()
		json.NewDecoder(f).Decode(&cfg)
	}
	return cfg
}

func saveConfig(cfg Config) {
	path := configPath()
	os.MkdirAll(filepath.Dir(path), 0o755)
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(cfg)
}

// resultItem represents a video search result.
type resultItem struct {
	title string
	id    string
}

func (r resultItem) Title() string       { return r.title }
func (r resultItem) Description() string { return r.id }
func (r resultItem) FilterValue() string { return r.title }

// subItem wraps a subscription URL for list display.
type subItem string

func (s subItem) Title() string       { return string(s) }
func (s subItem) Description() string { return "" }
func (s subItem) FilterValue() string { return string(s) }

// search queries yt-dlp for videos.
func search(query string) []resultItem {
	cmd := exec.Command("yt-dlp", "ytsearch10:"+query, "--get-title", "--get-id")
	out, err := cmd.Output()
	if err != nil {
		log.Fatalf("search failed: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var results []resultItem
	for i := 0; i+1 < len(lines); i += 2 {
		results = append(results, resultItem{title: lines[i], id: lines[i+1]})
	}
	return results
}

// fetchLatest returns the latest video for each subscription URL.
func fetchLatest(subs []string) []resultItem {
	var items []resultItem
	for _, s := range subs {
		cmd := exec.Command("yt-dlp", "--flat-playlist", "--get-title", "--get-id", s)
		out, err := cmd.Output()
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(strings.NewReader(string(out)))
		if scanner.Scan() {
			title := scanner.Text()
			if scanner.Scan() {
				id := scanner.Text()
				items = append(items, resultItem{title: title, id: id})
			}
		}
	}
	return items
}

func play(id string, audioOnly bool) {
	url := "https://www.youtube.com/watch?v=" + id
	args := []string{url}
	if audioOnly {
		args = append([]string{"--no-video"}, args...)
	}
	cmd := exec.Command("mpv", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

// model holds UI state.
type model struct {
	stage     string
	input     textinput.Model
	list      list.Model
	queue     []resultItem
	cfg       Config
	audioOnly bool
	width     int
	height    int
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Search YouTube..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 50
	return model{stage: "input", input: ti, cfg: loadConfig()}
}

func toItems(rs []resultItem) []list.Item {
	items := make([]list.Item, len(rs))
	for i := range rs {
		items[i] = rs[i]
	}
	return items
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		if m.stage != "input" && m.stage != "addsub" {
			m.list.SetSize(m.width, m.height-2)
		}
		return m, nil
	case tea.KeyMsg:
		switch m.stage {
		case "input":
			switch msg.String() {
			case "ctrl+c", "esc":
				return m, tea.Quit
			case "enter":
				query := strings.TrimSpace(m.input.Value())
				if query == "" {
					return m, nil
				}
				res := search(query)
				m.list = list.New(toItems(res), list.NewDefaultDelegate(), m.width, m.height-2)
				m.list.Title = "Results"
				m.stage = "results"
				m.input.SetValue("")
			case "ctrl+s":
				res := fetchLatest(m.cfg.Subscriptions)
				m.list = list.New(toItems(res), list.NewDefaultDelegate(), m.width, m.height-2)
				m.list.Title = "Latest"
				m.stage = "latest"
			case "ctrl+m":
				subs := make([]list.Item, len(m.cfg.Subscriptions))
				for i, s := range m.cfg.Subscriptions {
					subs[i] = subItem(s)
				}
				m.list = list.New(subs, list.NewDefaultDelegate(), m.width, m.height-2)
				m.list.Title = "Subscriptions"
				m.stage = "subs"
			case "ctrl+q":
				m.list = list.New(toItems(m.queue), list.NewDefaultDelegate(), m.width, m.height-2)
				m.list.Title = "Queue"
				m.stage = "queue"
			case "ctrl+t":
				m.audioOnly = !m.audioOnly
			}
		case "results", "latest":
			switch msg.String() {
			case "enter":
				if item, ok := m.list.SelectedItem().(resultItem); ok {
					play(item.id, m.audioOnly)
				}
			case "a":
				if item, ok := m.list.SelectedItem().(resultItem); ok {
					m.queue = append(m.queue, item)
				}
			case "ctrl+t":
				m.audioOnly = !m.audioOnly
			case "esc":
				m.stage = "input"
			}
		case "subs":
			switch msg.String() {
			case "a":
				m.stage = "addsub"
				m.input.Placeholder = "Channel URL"
				m.input.Focus()
			case "esc":
				m.stage = "input"
			}
		case "addsub":
			switch msg.String() {
			case "enter":
				url := strings.TrimSpace(m.input.Value())
				if url != "" {
					m.cfg.Subscriptions = append(m.cfg.Subscriptions, url)
					saveConfig(m.cfg)
				}
				m.input.SetValue("")
				m.input.Placeholder = "Search YouTube..."
				m.stage = "input"
			case "esc":
				m.input.SetValue("")
				m.input.Placeholder = "Search YouTube..."
				m.stage = "input"
			}
		case "queue":
			switch msg.String() {
			case "enter":
				if item, ok := m.list.SelectedItem().(resultItem); ok {
					play(item.id, m.audioOnly)
				}
			case "d":
				idx := m.list.Index()
				if idx >= 0 && idx < len(m.queue) {
					m.queue = append(m.queue[:idx], m.queue[idx+1:]...)
					m.list.RemoveItem(idx)
				}
			case "ctrl+t":
				m.audioOnly = !m.audioOnly
			case "esc":
				m.stage = "input"
			}
		}
	}

	switch m.stage {
	case "input", "addsub":
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	default:
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}
}

func (m model) View() string {
	switch m.stage {
	case "input":
		return fmt.Sprintf(
			"Audio Only: %v\n%s\n\nEnter to search | ctrl+s latest | ctrl+m manage subs | ctrl+q queue | ctrl+t toggle audio | esc quit",
			m.audioOnly, m.input.View())
	case "results":
		return m.list.View() + "\n\nenter play | a queue | ctrl+t toggle audio | esc back"
	case "latest":
		return m.list.View() + "\n\nenter play | a queue | ctrl+t toggle audio | esc back"
	case "subs":
		return m.list.View() + "\n\na add subscription | esc back"
	case "addsub":
		return "Add subscription:\n" + m.input.View() + "\n\nenter save | esc cancel"
	case "queue":
		return m.list.View() + "\n\nenter play | d delete | ctrl+t toggle audio | esc back"
	}
	return ""
}

func main() {
	if err := tea.NewProgram(initialModel()).Start(); err != nil {
		log.Fatal(err)
	}
}
