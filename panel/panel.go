package panel

import (
	"fmt"
	"slices"
	"strings"
)

type AsciiPanel struct {
	rows  int
	cols  int
	data  []rune
	Error error
}

type AsciiPanelErr struct {
	row  int
	col  int
	kind string
}

const horizontal rune = 0x2500
const vertical rune = 0x2502
const lowerLeft rune = 0x2570
const upperRight rune = 0x256E
const upperLeft rune = 0x256D
const lowerRight rune = 0x256F
const arrowDown rune = 0x2193

func repeat[A any](a A, n int) (out []A) {
	out = make([]A, n)

	for i := 0; i < n; i++ {
		out[i] = a
	}
	return
}

func (e AsciiPanelErr) Error() string {
	return fmt.Sprintf("%s error in row %d on col %d", e.kind, e.row, e.col)
}

func (p *AsciiPanel) err(e error) {
	if p.Error == nil {
		p.Error = e
	}
}

func NewPanel(rows, cols int) *AsciiPanel {
	return &AsciiPanel{rows: rows, cols: cols, data: repeat(' ', rows*cols)}
}

func (p *AsciiPanel) modify(f func(r, c int) (rune, bool, bool)) {
	for i := 0; i < p.rows; i++ {
		for j := 0; j < p.cols; j++ {
			r, mod, done := f(i, j)
			if done {
				return
			} else if mod {
				p.data[i*p.cols+j] = r
			}
		}
	}
}

func (p *AsciiPanel) WriteLine(s string, row, col int) int {
	if row >= p.rows {
		p.err(AsciiPanelErr{row, col, "panel overflow"})
		return 0
	}
	stringRunes := []rune(s)
	pos := 0
	lastMod := 0

	f := func(i, j int) (rune, bool, bool) {
		var r rune
		inRow := i == row
		isStart := j >= col
		hasRunes := pos < len(stringRunes)

		if inRow && isStart && hasRunes {
			newRune := stringRunes[pos]
			pos += 1
			lastMod = j
			return newRune, true, false
		} else if hasRunes {
			return r, false, false
		} else {
			return r, false, true
		}
	}

	p.modify(f)
	if pos < len(stringRunes) {
		p.err(AsciiPanelErr{row, p.cols, "row overflow"})
	}
	return lastMod
}

func (p *AsciiPanel) WriteString(s string, row, col int) int {
	lines := strings.Split(s, "\n")

	for _, l := range lines {
		p.WriteLine(l, row, col)
		row += 1
	}
	return row + 1
}

func (p *AsciiPanel) Frame() {
	top := slices.Concat([]rune{upperLeft}, repeat(horizontal, p.cols), []rune{upperRight})
	bot := slices.Concat([]rune{lowerLeft}, repeat(horizontal, p.cols), []rune{lowerRight})
	framedData := make([]rune, 0)

	for i := 0; i < p.rows; i++ {
		framedData = append(framedData, vertical)
		for j := 0; j < p.cols; j++ {
			framedData = append(framedData, p.data[i*p.cols+j])
		}
		framedData = append(framedData, vertical)
	}
	p.rows += 2
	p.cols += 2
	p.data = slices.Concat(top, framedData, bot)
}

func (p *AsciiPanel) Insert(img *AsciiPanel, y, x, rows, cols int) {

	for i := 0; i < img.rows; i++ {
		p.WriteLine(string(img.data[i*img.cols:(i+1)*img.cols]), y+i, x)
	}

	if img.Error != nil {
		p.err(img.Error)
	}
}

func (p *AsciiPanel) Render() string {
	var b strings.Builder

	for i := 0; i < p.rows; i++ {
		for j := 0; j < p.cols; j++ {
			b.WriteRune(p.data[i*p.cols+j])
		}
		b.WriteByte('\n')
	}

	return b.String()
}
