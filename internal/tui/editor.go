package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

type editor struct {
	subject []rune
	scol    int

	body     [][]rune
	row, col int
	want     int // desired visual column for up/down, -1 = unset

	field  int // 0 = subject, 1 = body
	scroll int
	w, h   int // body text area size, set by the model
}

func newEditor(subject, body string) editor {
	body = strings.ReplaceAll(body, "\r\n", "\n")
	var lines [][]rune
	for _, l := range strings.Split(body, "\n") {
		lines = append(lines, []rune(l))
	}
	return editor{
		subject: []rune(subject),
		scol:    len([]rune(subject)),
		body:    lines,
		want:    -1,
		field:   1,
		w:       40,
		h:       10,
	}
}

func (e *editor) text() (string, string) {
	lines := make([]string, len(e.body))
	for i, l := range e.body {
		lines[i] = string(l)
	}
	return string(e.subject), strings.Join(lines, "\n")
}

func insertAt(s []rune, i int, r rune) []rune {
	s = append(s, 0)
	copy(s[i+1:], s[i:])
	s[i] = r
	return s
}

func removeAt(s []rune, i int) []rune {
	return append(s[:i], s[i+1:]...)
}

func (e *editor) insert(s string) {
	for _, r := range s {
		switch r {
		case '\r':
		case '\n':
			if e.field == 1 {
				e.newline()
			}
		case '\t':
			e.insertRune(' ')
		default:
			e.insertRune(r)
		}
	}
}

func (e *editor) insertRune(r rune) {
	e.want = -1
	if e.field == 0 {
		e.subject = insertAt(e.subject, e.scol, r)
		e.scol++
		return
	}
	e.body[e.row] = insertAt(e.body[e.row], e.col, r)
	e.col++
}

func (e *editor) newline() {
	if e.field == 0 {
		return
	}
	e.want = -1
	line := e.body[e.row]
	head := append([]rune(nil), line[:e.col]...)
	tail := append([]rune(nil), line[e.col:]...)
	e.body[e.row] = head
	e.body = append(e.body, nil)
	copy(e.body[e.row+2:], e.body[e.row+1:])
	e.body[e.row+1] = tail
	e.row++
	e.col = 0
}

func (e *editor) backspace() {
	e.want = -1
	if e.field == 0 {
		if e.scol > 0 {
			e.subject = removeAt(e.subject, e.scol-1)
			e.scol--
		}
		return
	}
	if e.col > 0 {
		e.body[e.row] = removeAt(e.body[e.row], e.col-1)
		e.col--
		return
	}
	if e.row > 0 {
		prev := e.body[e.row-1]
		e.col = len(prev)
		e.body[e.row-1] = append(append([]rune(nil), prev...), e.body[e.row]...)
		e.body = append(e.body[:e.row], e.body[e.row+1:]...)
		e.row--
	}
}

func (e *editor) del() {
	e.want = -1
	if e.field == 0 {
		if e.scol < len(e.subject) {
			e.subject = removeAt(e.subject, e.scol)
		}
		return
	}
	line := e.body[e.row]
	if e.col < len(line) {
		e.body[e.row] = removeAt(line, e.col)
		return
	}
	if e.row < len(e.body)-1 {
		e.body[e.row] = append(append([]rune(nil), line...), e.body[e.row+1]...)
		e.body = append(e.body[:e.row+1], e.body[e.row+2:]...)
	}
}

func (e *editor) left() {
	e.want = -1
	if e.field == 0 {
		if e.scol > 0 {
			e.scol--
		}
		return
	}
	if e.col > 0 {
		e.col--
	} else if e.row > 0 {
		e.row--
		e.col = len(e.body[e.row])
	}
}

func (e *editor) right() {
	e.want = -1
	if e.field == 0 {
		if e.scol < len(e.subject) {
			e.scol++
		}
		return
	}
	if e.col < len(e.body[e.row]) {
		e.col++
	} else if e.row < len(e.body)-1 {
		e.row++
		e.col = 0
	}
}

func (e *editor) home() {
	e.want = -1
	if e.field == 0 {
		e.scol = 0
	} else {
		e.col = 0
	}
}

func (e *editor) end() {
	e.want = -1
	if e.field == 0 {
		e.scol = len(e.subject)
	} else {
		e.col = len(e.body[e.row])
	}
}

// vertical moves the cursor by one *visual* (wrapped) row.
func (e *editor) vertical(dir int) {
	if e.field == 0 {
		if dir > 0 {
			e.field = 1
			e.row = 0
			e.col = min(e.scol, len(e.body[0]))
			e.want = -1
		}
		return
	}
	w := max(e.w, 1)
	k := min(e.col/w, len(e.body[e.row])/w)
	if e.want < 0 {
		e.want = e.col - k*w
	}
	nk, row := k+dir, e.row
	switch {
	case nk < 0:
		if row == 0 {
			e.field = 0
			e.scol = min(e.col, len(e.subject))
			e.want = -1
			return
		}
		row--
		nk = len(e.body[row]) / w
	case nk > len(e.body[row])/w:
		if row == len(e.body)-1 {
			return
		}
		row++
		nk = 0
	}
	start := nk * w
	maxCol := start + w - 1
	if nk == len(e.body[row])/w {
		maxCol = len(e.body[row])
	}
	e.row = row
	e.col = min(start+e.want, maxCol)
}

func (e *editor) key(msg tea.KeyPressMsg) {
	switch msg.String() {
	case "left":
		e.left()
	case "right":
		e.right()
	case "up":
		e.vertical(-1)
	case "down":
		e.vertical(1)
	case "home", "ctrl+a":
		e.home()
	case "end", "ctrl+e":
		e.end()
	case "enter":
		e.newline()
	case "backspace":
		e.backspace()
	case "delete":
		e.del()
	default:
		if msg.Text != "" {
			e.insert(msg.Text)
		}
	}
}

func (e *editor) cursorVisualRow() int {
	w := max(e.w, 1)
	n := 0
	for r := 0; r < e.row; r++ {
		n += len(e.body[r])/w + 1
	}
	return n + min(e.col/w, len(e.body[e.row])/w)
}

func (e *editor) fixScroll() {
	cr := e.cursorVisualRow()
	if cr < e.scroll {
		e.scroll = cr
	}
	if cr >= e.scroll+e.h {
		e.scroll = cr - e.h + 1
	}
}

func cursorCell(rs []rune, i int) string {
	ch, after := " ", ""
	if i < len(rs) {
		ch, after = string(rs[i]), string(rs[i+1:])
	}
	return string(rs[:i]) + "\x1b[7m" + ch + reset + after
}

func (e *editor) renderSubject(w int, active bool) string {
	if !active {
		return string(e.subject)
	}
	start := 0
	if e.scol >= w {
		start = e.scol - w + 1
	}
	end := min(start+w, len(e.subject))
	return cursorCell(e.subject[start:end], e.scol-start)
}

func (e *editor) renderBody(w, h int, active bool) []string {
	w = max(w, 1)
	var rows []string
	for r, line := range e.body {
		n := len(line)
		for k := 0; k <= n/w; k++ {
			s, en := k*w, min(k*w+w, n)
			seg := string(line[s:en])
			if active && r == e.row && k == min(e.col/w, n/w) {
				seg = cursorCell(line[s:en], e.col-s)
			}
			rows = append(rows, seg)
		}
	}
	sc := min(e.scroll, max(0, len(rows)-h))
	return rows[sc:min(sc+h, len(rows))]
}
