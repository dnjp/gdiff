package main

import (
	"9fans.net/go/draw"
	"flag"
	"image"
)

type Line struct {
	t int
	s string
	f string
	l int
}

type Col struct {
	bg *draw.Image
	fg *draw.Image
}

type ColItem int

const (
	Lfile ColItem = iota
	Lsep
	Ladd
	Ldel
	Lnone
	Ncols
)

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
	black = flag.Bool("b", false, "draw black background")
	sr, scrollr, scrposr, listr, textr draw.Rectangle
	cols                               [Ncols]Col
	scrlcol                            Col
	scrollsize, lineh, nlines, offset  int
	lines                              []*Line
	lsize, lcount, maxlength, Δpan     int
	ellipsis                           string = "..."
)

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

func main() {
	flag.Parse()
	errs := make(chan error)
	var err error
	display, err = draw.Init(errs, "", "label", "1000x500")
	if err != nil {
		panic(err)
	}
	err = initcols(*black)
	if err != nil {
		panic(err)
	}
	defer display.Close()
	panic(<-errs)
}
