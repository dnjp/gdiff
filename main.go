package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"image"
	"io"
	"os"
	"strconv"

	"9fans.net/go/draw"
)

type Line struct {
	t LineType
	s []byte
	f []byte
}

type Col struct {
	bg *draw.Image
	fg *draw.Image
}

type LineType int

const (
	Lfile LineType = iota
	Lsep
	Ladd
	Ldel
	Lnone
	Ncols
)

func linetype(text []byte) LineType {
	t := Lnone
	if bytes.Contains(text, []byte("+++")) {
		t = Lfile
	} else if bytes.Contains(text, []byte("---")) {
		if len(text) > 4 {
			t = Lfile
		}
	} else if bytes.Contains(text, []byte("@@")) {
		t = Lsep
	} else if bytes.Contains(text, []byte("+")) {
		t = Ladd
	} else if bytes.Contains(text, []byte{'-'}) {
		t = Ldel
	}
	return t
}

type Sep int

const (
	Scrollwidth Sep = 12
	Scrollgap       = 2
	Margin          = 8
	Hpadding        = 4
	Vpadding        = 2
)

var (
	display                            *draw.Display
	mousectl *draw.Mousectl
	keyboardctl *draw.Keyboardctl
	black                              = flag.Bool("b", false, "draw black background")
	sr, scrollr, scrposr, listr, textr draw.Rectangle
	cols                               [Ncols]Col
	scrlcol                            Col
	scrollsize, lineh, nlines, offset  int
	lsize, maxlength, Δpan             int
	ellipsis                           string = "..."
)

func redraw() {
	// var lr draw.Rectangle
	// var i, h, y int
}

func eresized(lines []*Line, new bool) error {
	if new {
		err := display.Attach(draw.RefNone)
		if err != nil {
			return err
		}
	}
	sr = display.ScreenImage.R
	scrollr = sr
	scrollr.Max.X = scrollr.Min.X + int(Scrollwidth) + int(Scrollgap)
	listr = sr
	listr.Min.X = scrollr.Max.X
	textr = listr.Inset(Margin)
	lineh = Vpadding+display.Font.Height
	nlines = textr.Dy() / lineh
	scrollsize = draw.MouseScrollSize(len(lines))
	if offset > 0 && offset+nlines > len(lines) {
		offset = len(lines) - nlines+1
	}
	redraw()
	return nil
}

func initcol(c *Col, fg, bg draw.Color) error {
	var err error
	c.fg, err = display.AllocImage(image.Rect(0, 0, 1, 1), display.ScreenImage.Pix, true, fg)
	if err != nil {
		return err
	}
	c.bg, err = display.AllocImage(image.Rect(0, 0, 1, 1), display.ScreenImage.Pix, true, bg)
	if err != nil {
		return err
	}
	return nil
}

func initcols(black bool) error {
	var err error
	if black {
		if err = initcol(&scrlcol, 0x22272EFF, 0xADBAC7FF); err != nil {
			return err
		}
		if err = initcol(&cols[Lfile], 0xADBAC7FF, 0x2D333BFF); err != nil {
			return err
		}
		if err = initcol(&cols[Lsep], 0xADBAC7FF, 0x263549FF); err != nil {
			return err
		}
		if err = initcol(&cols[Ladd], 0xADBAC7FF, 0x273732FF); err != nil {
			return err
		}
		if err = initcol(&cols[Ldel], 0xADBAC7FF, 0x3F2D32FF); err != nil {
			return err
		}
		if err = initcol(&cols[Lnone], 0xADBAC7FF, 0x22272EFF); err != nil {
			return err
		}
	} else {
		if err = initcol(&scrlcol, draw.White, 0x999999FF); err != nil {
			return err
		}
		if err = initcol(&cols[Lfile], draw.Black, 0xEFEFEFFF); err != nil {
			return err
		}
		if err = initcol(&cols[Lsep], draw.Black, 0xEAFFFFFF); err != nil {
			return err
		}
		if err = initcol(&cols[Ladd], draw.Black, 0xE6FFEDFF); err != nil {
			return err
		}
		if err = initcol(&cols[Ldel], draw.Black, 0xFFEEF0FF); err != nil {
			return err
		}
		if err = initcol(&cols[Lnone], draw.Black, draw.White); err != nil {
			return err
		}
	}
	return nil
}

func parseline(f, s []byte) *Line {
	l := &Line{
		t: linetype(s),
		s: s,
	}
	if l.t != Lfile && l.t != Lsep {
		l.f = f
	} else {
		l.f = nil
	}
	lens := len(s)
	if lens > maxlength {
		maxlength = lens
	}
	return l
}

func tokenize(s []byte, n int) [][]byte {
	return bytes.SplitN(s, []byte{' '}, n)
}

func lineno(s []byte) int {
	var p []byte
	var t [][]byte
	var n, l int

	p = make([]byte, len(s))
	copy(p, s)
	t = tokenize(p, 5)
	n = len(t)
	if n <= 0 {
		return -1
	}
	l, _ = strconv.Atoi(string(t[2]))
	return l
}

func parse() ([]*Line, error) {
	var l *Line
	var s, f []byte
	var ab bool
	var n int
	var err error
	lines := make([]*Line, 0)

	reader := bufio.NewReader(os.Stdin)
	for {
		s, _, err = reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return lines, err
		}
		l = parseline(f, s)
		// TODO: check len(s)
		if l.t == Lfile && l.s[0] == '-' && bytes.Contains(l.s[4:], []byte("a/")) {
			ab = true
		}
		if l.t == Lfile && l.s[0] == '+' {
			f = l.s[4:]
			if ab && bytes.Contains(f, []byte("b/")) {
				f = f[1:]
				_, err = os.Lstat(string(f))
				if err != nil {
					f = f[1:]
				}
			}
		} else if l.t == Lsep {
			n = lineno(l.s)
		} else if l.t == Ladd || l.t == Lnone {
			n++
		}
		lines = append(lines, l)
	}
	return lines, nil
}

func main() {
	flag.Parse()

	lines, err := parse()
	if err != nil {
		panic(err)
	}
	if len(lines) == 0 {
		fmt.Fprintf(os.Stderr, "no diff\n")
		os.Exit(1)
	}

	errs := make(chan error)
	display, err = draw.Init(errs, "", "label", "1000x500")
	if err != nil {
		panic(err)
	}
	err = initcols(*black)
	if err != nil {
		panic(err)
	}

	mousectl = display.InitMouse()
	keyboardctl = display.InitKeyboard()

	defer display.Close()
	panic(<-errs)
}
