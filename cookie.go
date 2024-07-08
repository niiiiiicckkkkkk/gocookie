package main

import (
	"cookieclicker/entity"
	"cookieclicker/panel"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const cookie rune = 0x1F36A

type model struct {
	cookies     float64
	inventory   map[string]int
	termWidth   int
	termHeight  int
	uiSelection int
	itemsOffset int
	ticks       int
	history     [3]float64
	historyIdx  int
}

type save struct {
	Cookies   float64
	Inventory map[string]int
}

func loadGame() (model, error) {
	saveFile, err := os.Open("save.json")
	if err != nil {
		return model{}, err
	}
	d := json.NewDecoder(saveFile)
	var save save
	if err := d.Decode(&save); err != nil {
		return model{}, err
	}
	if save.Inventory == nil {
		save.Inventory = make(map[string]int)
	}
	return model{inventory: save.Inventory, cookies: save.Cookies}, nil
}

func (m model) closeGame() {
	saveFile, err := os.Create("save.json")
	if err != nil {
		return
	}

	e := json.NewEncoder(saveFile)
	e.Encode(save{Inventory: m.inventory, Cookies: m.cookies})
	saveFile.Close()
}

func (m model) cps() (rate float64) {
	rate = 0
	for _, name := range items {
		rate += lookup[name].Cps(m.inventory[name])
	}
	return
}

type gametick struct{}

var lookup, items = entity.Items()

func tick() tea.Msg {
	time.Sleep(time.Second)
	return gametick{}
}

func (m model) Init() tea.Cmd {
	return tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// update the gamestate
	case gametick:
		m.ticks += 1
		m.ticks = m.ticks % 5
		if m.ticks == 0 {
			// log for graph
			m.history[m.historyIdx] = m.cookies
			m.historyIdx += 1
			m.historyIdx = m.historyIdx % 3
		}
		m.cookies += m.cps()
		return m, tick
	case tea.WindowSizeMsg:
		m.termHeight = msg.Height
		m.termWidth = msg.Width
		return m, nil

	case tea.KeyMsg:
		// Ctrl+c exits. Even with short running programs it's good to have
		// a quit key, just in case your logic is off. Users will be very
		// annoyed if they can't exit.
		switch msg.String() {
		case "q":
			m.closeGame()
			return m, tea.Quit
		case "c":
			m.cookies += 1
			return m, nil
		case "j":
			if m.uiSelection == 2 && m.itemsOffset < len(items)-3 {
				m.itemsOffset += 1
			} else if m.uiSelection < 2 {
				m.uiSelection += 1
			}
		case "k":
			if m.uiSelection == 0 && m.itemsOffset > 0 {
				m.itemsOffset -= 1
			} else if m.uiSelection > 0 {
				m.uiSelection -= 1
			}
		case "b":
			i := m.itemsOffset + m.uiSelection
			item := lookup[items[i]]
			cost := item.Cost(1, m.inventory[items[i]])
			if cost <= m.cookies {
				m.cookies -= cost
				m.inventory[items[i]] += 1
			}
		}
	}

	// If we happen to get any other messages, don't do anything.
	return m, nil
}

func (m model) cookiePanel(rows, cols int) *panel.AsciiPanel {
	cookieDisplay := panel.NewPanel(rows, cols)
	var b strings.Builder

	b.WriteRune(cookie)
	b.WriteRune(0x0)
	b.WriteString("  : ")
	b.WriteString(fmt.Sprintf("%d (%.1f)", int(m.cookies), m.cps()))

	cookieDisplay.WriteString(b.String(), 0, 0)

	return cookieDisplay
}

func (m model) inventoryPanel(rows, cols int) *panel.AsciiPanel {
	inventoryDisplay := panel.NewPanel(rows, cols)

	for i := 0; i < 3; i++ {
		idx := i + m.itemsOffset
		var b strings.Builder
		// make ascii panel
		p := panel.NewPanel(rows/5, cols-2)
		item := lookup[items[idx]]
		name := items[idx]
		rate := item.Cps(m.inventory[name])
		_cost := item.Cost(1, m.inventory[name])
		owned := m.inventory[name]
		icon := []rune{item.Icon, 0x0}
		cost := fmt.Sprintf("%7d", int(_cost))
		if _cost > m.cookies {
			cost = "-------"
		}
		readout := fmt.Sprintf("%s %7d  (%8.1f)  %s", string(icon), owned, rate, cost)

		b.WriteString(readout)
		p.WriteString(b.String(), 0, 0)

		// if i == m.display... add a frame
		if i == m.uiSelection {
			p.Frame()
			inventoryDisplay.Insert(p, rows/3*i, 0, rows/5+2, cols)
			continue
		}

		// insert into inventory display
		inventoryDisplay.Insert(p, rows/3*i, 0, rows/5, cols-2)
	}

	return inventoryDisplay
}

func (m model) graph(rows, cols int) *panel.AsciiPanel {
	cookie := "⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣀⡴⠚⣉⡙⠲⠦⠤⠤⣤⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀\n" +
		"⠀⠀⠀⠀⠀⠀⢀⣴⠛⠉⠉⠀⣾⣷⣿⡆⠀⠀⠀⠐⠛⠿⢟⡲⢦⡀⠀⠀⠀⠀\n" +
		"⠀⠀⠀⠀⣠⢞⣭⠎⠀⠀⠀⠀⠘⠛⠛⠀⠀⢀⡀⠀⠀⠀⠀⠈⠓⠿⣄⠀⠀⠀\n" +
		"⠀⠀⠀⡜⣱⠋⠀⠀⣠⣤⢄⠀⠀⠀⠀⠀⠀⣿⡟⣆⠀⠀⠀⠀⠀⠀⠻⢷⡄⠀\n" +
		"⠀⢀⣜⠜⠁⠀⠀⠀⢿⣿⣷⣵⠀⠀⠀⠀⠀⠿⠿⠿⠀⠀⣴⣶⣦⡀⠀⠰⣹⡆\n" +
		"⢀⡞⠆⠀⣀⡀⠀⠀⠘⠛⠉⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢿⣿⣶⠇⠀⢠⢻⡇\n" +
		"⢸⠃⠘⣾⣏⡇⠀⠀⠀⠀⠀⠀⠀⡀⠀⠀⠀⠀⠀⠀⣠⣤⣤⡉⠁⠀⠀⠈⠫⣧\n" +
		"⡸⡄⠀⠘⠟⠀⠀⠀⠀⠀⠀⣰⣿⣟⢧⠀⠀⠀⠀⠰⡿⣿⣿⢿⠀⠀⣰⣷⢡⢸\n" +
		"⣿⡇⠀⠀⠀⣰⣿⡻⡆⠀⠀⠻⣿⣿⣟⠀⠀⠀⠀⠀⠉⠉⠉⠀⠀⠘⢿⡿⣸⡞\n" +
		"⠹⣽⣤⣤⣤⣹⣿⡿⠇⠀⠀⠀⠀⠉⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⡔⣽⠀\n" +
		"⠀⠙⢻⡙⠟⣹⠟⢷⣶⣄⢀⣴⣶⣄⠀⠀⠀⠀⠀⢀⣤⡦⣄⠀⠀⢠⣾⢸⠏⠀\n" +
		"⠀⠀⠘⠀⠀⠀⠀⠀⠈⢷⢼⣿⡿⡽⠀⠀⠀⠀⠀⠸⣿⣿⣾⠀⣼⡿⣣⠟⠀⠀\n" +
		"⠀⠀⠀⠀⠀⠀⠀⠀⢠⡾⣆⠑⠋⠀⢀⣀⠀⠀⠀⠀⠈⠈⢁⣴⢫⡿⠁⠀⠀⠀\n" +
		"⠀⠀⠀⠀⠀⠀⠀⠀⠈⠙⣧⣄⡄⠴⣿⣶⣿⢀⣤⠶⣞⣋⣩⣵⠏⠀⠀⠀⠀⠀\n" +
		"⠀⠀⠀⠀⠀⠀⠀⠀⠀⢺⣿⢯⣭⣭⣯⣯⣥⡵⠿⠟⠛⠉⠉⠀⠀⠀⠀⠀⠀⠀"

	p := panel.NewPanel(rows, cols)
	p.WriteString(cookie, 0, 0)
	return p
}

func (m model) View() string {
	if m.termHeight == 0 || m.termWidth == 0 {
		return ""
	}

	background := panel.NewPanel(m.termHeight, m.termWidth)
	rows, cols := m.termHeight-2, m.termWidth-2

	frame := panel.NewPanel(rows, cols)

	cookieDisplay := m.cookiePanel((rows-2)/4-2, (cols/2)-4)
	frame.Insert(cookieDisplay, 1, (cols/2)+2, (rows-2)/4-2, (cols/2)-4)
	ui := m.inventoryPanel(rows-2, (cols / 2))
	frame.Insert(ui, 1, 1, rows-2, cols/2)

	graph := m.graph((rows-2)/4*3, (cols/2)-4)
	frame.Insert(graph, (rows-2)/4, (cols/2)+2, (rows-2)/4*3, (cols/2)-4)

	background.Insert(frame, 1, 1, m.termHeight-2, m.termWidth-2)
	if background.Error != nil {
		return background.Error.Error()
	} else {
		var b strings.Builder
		b.WriteString(background.Render())
		return b.String()
	}
}

func main() {
	m, err := loadGame()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Printf("Uh oh, there was an error: %v\n", err)
		os.Exit(1)
	}
}
