package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// ── TUI item ──────────────────────────────────────────────────────────────

type tuiItem struct {
	label    string // primary label
	sub      string // secondary hint (dim, right side)
	selected bool
	disabled bool // can't be toggled (greyed out)
}

// ── Picker ────────────────────────────────────────────────────────────────
// Picker renders an interactive list. multi=true gives checkboxes,
// multi=false gives a radio (single selection).

type picker struct {
	title  string
	hint   string // footer hint line
	items  []tuiItem
	cursor int
	multi  bool
	drawn  int // lines drawn in last render, for clearing
}

func newRadio(title string, items []tuiItem) *picker {
	p := &picker{title: title, multi: false, items: items}
	// pre-select first non-disabled item
	for i, it := range items {
		if !it.disabled {
			p.items[i].selected = true
			p.cursor = i
			break
		}
	}
	p.hint = "↑↓ navigate  ENTER confirm  Q quit"
	return p
}

func newCheckbox(title string, items []tuiItem) *picker {
	p := &picker{title: title, multi: true, items: items}
	p.hint = "↑↓ navigate  SPACE toggle  ENTER confirm  Q quit"
	return p
}

// Run renders the picker, handles input, and returns whether the user confirmed.
// Returns false if user quit / cancelled.
func (p *picker) run() (ok bool) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return false
	}
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return false
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\033[?25l") // hide cursor
	defer fmt.Print("\033[?25h") // restore cursor on exit

	p.render()
	for {
		b, err := reader.ReadByte()
		if err != nil {
			break
		}

		switch b {
		case 3, 'q', 'Q': // ctrl-c or q
			p.clear()
			return false
		case 13: // enter
			p.clear()
			return true
		case 32: // space
			if p.multi && !p.items[p.cursor].disabled {
				p.items[p.cursor].selected = !p.items[p.cursor].selected
			}
		case 27: // ESC sequence
			b2, _ := reader.ReadByte()
			if b2 == '[' {
				b3, _ := reader.ReadByte()
				switch b3 {
				case 'A': // up
					p.moveCursor(-1)
				case 'B': // down
					p.moveCursor(1)
				}
			}
		}

		p.clear()
		p.render()
	}
	return false
}

func (p *picker) moveCursor(delta int) {
	next := p.cursor + delta
	for next >= 0 && next < len(p.items) {
		if !p.items[next].disabled {
			p.cursor = next
			return
		}
		next += delta
	}
}

func (p *picker) render() {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\n  %s%s%s\n\n", bold, p.title, reset))
	for i, it := range p.items {
		cursor := "  "
		if i == p.cursor {
			cursor = fmt.Sprintf("%s▶%s ", cyan, reset)
		}

		var check string
		if p.multi {
			if it.disabled {
				check = fmt.Sprintf("%s[ ]%s", dim, reset)
			} else if it.selected {
				check = fmt.Sprintf("%s[✓]%s", green, reset)
			} else {
				check = "[ ]"
			}
		} else {
			if it.selected {
				check = fmt.Sprintf("%s●%s", green, reset)
			} else {
				check = "○"
			}
		}

		label := it.label
		if it.disabled {
			label = fmt.Sprintf("%s%s%s", dim, it.label, reset)
		}

		sub := ""
		if it.sub != "" {
			sub = fmt.Sprintf("  %s%s%s", dim, it.sub, reset)
		}

		sb.WriteString(fmt.Sprintf("  %s%s %s%s\n", cursor, check, label, sub))
	}
	sb.WriteString(fmt.Sprintf("\n  %s%s%s\n", dim, p.hint, reset))

	out := sb.String()
	p.drawn = strings.Count(out, "\n")
	fmt.Print(out)
}

func (p *picker) clear() {
	for i := 0; i < p.drawn; i++ {
		fmt.Print("\033[A\033[2K")
	}
	p.drawn = 0
}

// Selected returns indices of selected items.
func (p *picker) selected() []int {
	var out []int
	for i, it := range p.items {
		if it.selected {
			out = append(out, i)
		}
	}
	return out
}

// FirstSelected returns the index of the first selected item, or -1.
func (p *picker) firstSelected() int {
	for i, it := range p.items {
		if it.selected {
			return i
		}
	}
	return -1
}
