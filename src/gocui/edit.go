// Copyright 2014 The gocui Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gocui

import "strings"

const maxInt = int(^uint(0) >> 1)

// Editor interface must be satisfied by gocui editors.
type Editor interface {
	Edit(v *View, key Key, ch rune, mod Modifier)
}

// The EditorFunc type is an adapter to allow the use of ordinary functions as
// Editors. If f is a function with the appropriate signature, EditorFunc(f)
// is an Editor object that calls f.
type EditorFunc func(v *View, key Key, ch rune, mod Modifier)

// Edit calls f(v, key, ch, mod)
func (f EditorFunc) Edit(v *View, key Key, ch rune, mod Modifier) {
	f(v, key, ch, mod)
}

// DefaultEditor is the default editor.
var DefaultEditor Editor = EditorFunc(simpleEditor)

func findPos(v *View, pos int) (int, int) {
	line := 0
	linesExist := v.lines != nil
	for linesExist && line < len(v.lines) {
		ll := 0
		if v.lines[line] == nil {
			ll = 1
		} else {
			ll = len(v.lines[line]) + 1
		}
		if pos >= ll {
			pos -= ll
			line++
		} else {
			break
		}
	}
	return pos, line
}

func findAbsPos(v *View, x, y int) int {
	pos := 0
	for y > 0 {
		pos += len(v.lines[y-1]) + 1
		y--
	}
	return pos + x
}

func (v *View) WriteRuneAtPos(pos int, ch rune, log func(...interface{})) {
	v.EditWritePos(pos, ch)
}

func (v *View) DeleteRuneAtPos(pos int, log func(...interface{})) {
	v.EditDeletePos(pos, true)
}

func (v *View) resetToBuf() {
	v.lines = nil
	lines := strings.Split(v.buffer, string('\n'))
	v.lines = make([][]cell, len(lines))
	for i, line := range lines {
		runes := []rune(line)
		v.lines[i] = make([]cell, len(line))
		for j, ch := range runes {
			v.lines[i][j] = cell{ch, v.BgColor, v.FgColor}
		}
	}
}

// simpleEditor is used as the default gocui editor.
func simpleEditor(v *View, key Key, ch rune, mod Modifier) {
	switch {
	case ch != 0 && mod == 0:
		v.EditWrite(ch)
	case key == KeySpace:
		v.EditWrite(' ')
	case key == KeyBackspace || key == KeyBackspace2:
		v.EditDelete(true)
	case key == KeyDelete:
		v.EditDelete(false)
	case key == KeyEnter:
		v.EditWrite('\n')
	case key == KeyArrowDown:
		v.MoveCursor(0, 1)
	case key == KeyArrowUp:
		v.MoveCursor(0, -1)
	case key == KeyArrowLeft:
		v.MoveCursor(-1, 0)
	case key == KeyArrowRight:
		v.MoveCursor(1, 0)
	}
}

// EditDelete deletes a rune at the cursor position. back determines the
// direction.
func (v *View) EditDelete(back bool) {
	x, y := v.cx, v.cy
	pos := findAbsPos(v, x, y)
	v.EditDeletePos(pos, back)
	if back && pos > 0 {
		pos--
	}
	v.cx, v.cy = findPos(v, pos)
}

func (v *View) EditDeletePos(pos int, back bool) {
	v.tainted = true
	if back && pos == 0{
		// don't delete at index
		return
	}
	if back && pos > 0 {
		pos--
	}
	if pos == 0 {
		v.buffer = v.buffer[1:]
	} else if pos == len(v.buffer)-1 {
		v.buffer = v.buffer[:len(v.buffer)-1]
	} else {
		v.buffer = v.buffer[:pos] + v.buffer[pos+1:]
	}

	v.resetToBuf()
}

func (v *View) EditWrite(ch rune) {
	x, y := v.cx, v.cy
	pos := findAbsPos(v, x, y)
	v.EditWritePos(pos, ch)
	v.cx, v.cy = findPos(v, pos+1)
}
func (v *View) EditWritePos(pos int, ch rune) {
	v.tainted = true
	if pos == 0 {
		v.buffer = string(ch) + v.buffer
	} else if pos >= len(v.buffer) {
		v.buffer = v.buffer + string(ch)
	} else {
		v.buffer = v.buffer[:pos] + string(ch) + v.buffer[pos:]
	}

	v.resetToBuf()
}

// MoveCursor moves the cursor taking into account the width of the line/view,
// displacing the origin if necessary.
func (v *View) MoveCursor(dx, dy int) {
	_, maxY := v.Size()
	cx, cy := v.cx+dx, v.cy+dy
	if v.lines == nil {
		return
	}
	if cy < 0 || cy >= maxY || cy >= len(v.lines) {
		return
	}
	if cx < 0 {
		cy--
		if cy >= 0 {
			if v.lines[cy] == nil {
				cx = 0
			} else {
				cx = len(v.lines[cy])
			}
		}
	} else if (v.lines[cy] != nil && cx > len(v.lines[cy])) || (v.lines[cy] == nil && cx > 0) {
		cy++
		cx = 0
	}
	if cy < 0 || cy >= maxY || cy >= len(v.lines) {
		return
	}
	v.cx = cx
	v.cy = cy
}
