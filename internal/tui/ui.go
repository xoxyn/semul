package tui

import "strings"

const reset = "\x1b[0m"

func c(code, s string) string { return "\x1b[" + code + "m" + s + reset }

func rlen(s string) int { return len([]rune(s)) }

// visWidth is the printed width of s, ignoring ANSI escapes.
func visWidth(s string) int {
	n, esc := 0, false
	for _, r := range s {
		switch {
		case esc:
			if r == 'm' {
				esc = false
			}
		case r == 0x1b:
			esc = true
		default:
			n++
		}
	}
	return n
}

// fit truncates or pads s to exactly w columns. ANSI escapes are preserved.
func fit(s string, w int) string {
	var b strings.Builder
	vis := 0
	rs := []rune(s)
	for i := 0; i < len(rs); i++ {
		r := rs[i]
		if r == 0x1b {
			j := i
			for j < len(rs) && rs[j] != 'm' {
				j++
			}
			if j >= len(rs) {
				break
			}
			b.WriteString(string(rs[i : j+1]))
			i = j
			continue
		}
		if vis >= w {
			continue
		}
		b.WriteRune(r)
		vis++
	}
	b.WriteString(reset)
	if vis < w {
		b.WriteString(strings.Repeat(" ", w-vis))
	}
	return b.String()
}

// plainPad truncates/pads plain (unstyled) text to exactly w columns.
func plainPad(s string, w int) string {
	rs := []rune(s)
	if len(rs) > w {
		rs = rs[:w]
	}
	return string(rs) + strings.Repeat(" ", w-len(rs))
}

// box draws a lazygit-style panel with the title in the top border and an
// optional footer (e.g. "2 of 6") in the bottom border.
func box(title, footer string, lines []string, w, h int, active bool) string {
	if w < 4 || h < 3 {
		return ""
	}
	bc, tc := "90", "1"
	if active {
		bc, tc = "32", "1;32"
	}
	b := func(s string) string { return c(bc, s) }
	iw, ih := w-2, h-2

	t := " " + title + " "
	top := b("╭─") + c(tc, t) + b(strings.Repeat("─", max(0, iw-1-rlen(t)))+"╮")

	var bottom string
	if footer != "" {
		f := " " + footer + " "
		bottom = b("╰"+strings.Repeat("─", max(0, iw-rlen(f)-1))) + c("90", f) + b("─╯")
	} else {
		bottom = b("╰" + strings.Repeat("─", iw) + "╯")
	}

	rows := []string{top}
	for i := 0; i < ih; i++ {
		l := ""
		if i < len(lines) {
			l = lines[i]
		}
		rows = append(rows, b("│")+fit(l, iw)+b("│"))
	}
	rows = append(rows, bottom)
	return strings.Join(rows, "\n")
}

func joinH(a, b string) string {
	al, bl := strings.Split(a, "\n"), strings.Split(b, "\n")
	out := make([]string, max(len(al), len(bl)))
	for i := range out {
		x, y := "", ""
		if i < len(al) {
			x = al[i]
		}
		if i < len(bl) {
			y = bl[i]
		}
		out[i] = x + y
	}
	return strings.Join(out, "\n")
}

func center(content string, w, h int) string {
	lines := strings.Split(content, "\n")
	cw := 0
	for _, l := range lines {
		cw = max(cw, visWidth(l))
	}
	pad := strings.Repeat(" ", max(0, (w-cw)/2))
	for i := range lines {
		lines[i] = pad + lines[i]
	}
	top := strings.Repeat("\n", max(0, (h-len(lines))/2))
	return top + strings.Join(lines, "\n")
}

func wrap(s string, w int) []string {
	if w < 1 {
		w = 1
	}
	var out []string
	for _, para := range strings.Split(s, "\n") {
		if strings.TrimSpace(para) == "" {
			out = append(out, "")
			continue
		}
		line := ""
		for _, word := range strings.Fields(para) {
			for rlen(word) > w {
				if line != "" {
					out = append(out, line)
					line = ""
				}
				r := []rune(word)
				out = append(out, string(r[:w]))
				word = string(r[w:])
			}
			switch {
			case line == "":
				line = word
			case rlen(line)+1+rlen(word) <= w:
				line += " " + word
			default:
				out = append(out, line)
				line = word
			}
		}
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}
