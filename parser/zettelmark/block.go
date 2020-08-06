//-----------------------------------------------------------------------------
// Copyright (c) 2020 Detlef Stern
//
// This file is part of zettelstore.
//
// Zettelstore is free software: you can redistribute it and/or modify it under
// the terms of the GNU Affero General Public License as published by the Free
// Software Foundation, either version 3 of the License, or (at your option)
// any later version.
//
// Zettelstore is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE. See the GNU Affero General Public License
// for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with Zettelstore. If not, see <http://www.gnu.org/licenses/>.
//-----------------------------------------------------------------------------

// Package zettelmark provides a parser for zettelmarkup.
package zettelmark

import (
	"fmt"

	"zettelstore.de/z/ast"
	"zettelstore.de/z/input"
)

// parseBlockSlice parses a sequence of blocks.
func (cp *zmkP) parseBlockSlice() ast.BlockSlice {
	inp := cp.inp
	var lastPara *ast.ParaNode
	result := make(ast.BlockSlice, 0, 2)
	for inp.Ch != input.EOS {
		bn, cont := cp.parseBlock(lastPara)
		if bn != nil {
			result = append(result, bn)
		}
		if !cont {
			lastPara, _ = bn.(*ast.ParaNode)
		}
	}
	if cp.nestingLevel != 0 {
		panic("Nesting level was not decremented")
	}
	return result
}

// parseBlock parses one block.
func (cp *zmkP) parseBlock(lastPara *ast.ParaNode) (res ast.BlockNode, cont bool) {
	inp := cp.inp
	pos := inp.Pos
	if cp.nestingLevel <= maxNestingLevel {
		cp.nestingLevel++
		defer func() { cp.nestingLevel-- }()

		var bn ast.BlockNode
		success := false

		switch inp.Ch {
		case input.EOS:
			return nil, false
		case '\n', '\r':
			inp.EatEOL()
			for _, l := range cp.lists {
				if lits := len(l.Items); lits > 0 {
					l.Items[lits-1] = append(l.Items[lits-1], &nullItemNode{})
				}
			}
			if cp.defl != nil {
				defPos := len(cp.defl.Definitions) - 1
				if ldds := len(cp.defl.Definitions[defPos].Descriptions); ldds > 0 {
					cp.defl.Definitions[defPos].Descriptions[ldds-1] = append(
						cp.defl.Definitions[defPos].Descriptions[ldds-1], &nullDescriptionNode{})
				}
			}
			return nil, false
		case ':':
			bn, success = cp.parseColon()
		case '`':
			cp.clearStacked()
			bn, success = cp.parseVerbatim()
		case '"', '<':
			cp.clearStacked()
			bn, success = cp.parseRegion()
		case '=':
			cp.clearStacked()
			bn, success = cp.parseHeading()
		case '-':
			cp.clearStacked()
			bn, success = cp.parseHRule()
		case '*', '#', '>':
			cp.table = nil
			cp.defl = nil
			bn, success = cp.parseList()
		case ';':
			cp.lists = nil
			cp.table = nil
			bn, success = cp.parseDefTerm()
		case ' ':
			cp.table = nil
			bn, success = cp.parseIndent()
		case '|':
			cp.lists = nil
			cp.defl = nil
			bn, success = cp.parseRow()
		}

		if success {
			return bn, false
		}
	}
	inp.SetPos(pos)
	cp.clearStacked()
	pn := cp.parsePara()
	if lastPara != nil {
		lastPara.Inlines = append(lastPara.Inlines, pn.Inlines...)
		return nil, true
	}
	return pn, false
}

// parseColon determines which element should be parsed.
func (cp *zmkP) parseColon() (ast.BlockNode, bool) {
	inp := cp.inp
	if inp.PeekN(1) == ':' {
		cp.clearStacked()
		return cp.parseRegion()
	}
	return cp.parseDefDescr()
}

// parsePara parses paragraphed inline material.
func (cp *zmkP) parsePara() *ast.ParaNode {
	pn := &ast.ParaNode{}
	for {
		in := cp.parseInline()
		if in == nil {
			return pn
		}
		pn.Inlines = append(pn.Inlines, in)
		if _, ok := in.(*ast.BreakNode); ok {
			ch := cp.inp.Ch
			switch ch {
			// Must contain all cases from above switch in parseBlock.
			case input.EOS, '\n', '\r', '`', '"', '<', '=', '-', '*', '#', '>', ';', ':', ' ', '|':
				return pn
			}
		}
	}
}

// countDelim read from input until a non-delimiter is found and returns number of delimiter chars.
func (cp *zmkP) countDelim(delim rune) int {
	cnt := 0
	for cp.inp.Ch == delim {
		cnt++
		cp.inp.Next()
	}
	return cnt
}

// parseVerbatim parses a verbatim block.
func (cp *zmkP) parseVerbatim() (rn *ast.VerbatimNode, success bool) {
	inp := cp.inp
	fch := inp.Ch
	cnt := cp.countDelim(fch)
	if cnt < 3 {
		return nil, false
	}
	attrs := cp.parseAttributes(true)
	inp.SkipToEOL()
	if inp.Ch == input.EOS {
		return nil, false
	}
	rn = &ast.VerbatimNode{Code: ast.VerbatimProg, Attrs: attrs}
	for {
		inp.EatEOL()
		posL := inp.Pos
		switch inp.Ch {
		case fch:
			if cp.countDelim(fch) >= cnt {
				inp.SkipToEOL()
				return rn, true
			}
			inp.SetPos(posL)
		case input.EOS:
			return nil, false
		}
		inp.SkipToEOL()
		rn.Lines = append(rn.Lines, inp.Src[posL:inp.Pos])
	}
}

var runeRegion = map[rune]ast.RegionCode{
	':': ast.RegionSpan,
	'<': ast.RegionQuote,
	'"': ast.RegionVerse,
}

// parseRegion parses a block region.
func (cp *zmkP) parseRegion() (rn *ast.RegionNode, success bool) {
	inp := cp.inp
	fch := inp.Ch
	code, ok := runeRegion[fch]
	if !ok {
		panic(fmt.Sprintf("%q is not a region char", fch))
	}
	cnt := cp.countDelim(fch)
	if cnt < 3 {
		return nil, false
	}
	attrs := cp.parseAttributes(true)
	inp.SkipToEOL()
	if inp.Ch == input.EOS {
		return nil, false
	}
	rn = &ast.RegionNode{Code: code, Attrs: attrs}
	var lastPara *ast.ParaNode
	inp.EatEOL()
	for {
		posL := inp.Pos
		switch inp.Ch {
		case fch:
			if cp.countDelim(fch) >= cnt {
				cp.clearStacked() // remove any lists defined in the region
				for inp.Ch == ' ' {
					inp.Next()
				}
				for {
					switch inp.Ch {
					case input.EOS, '\n', '\r':
						return rn, true
					}
					in := cp.parseInline()
					if in == nil {
						return rn, true
					}
					rn.Inlines = append(rn.Inlines, in)
				}
			}
			inp.SetPos(posL)
		case input.EOS:
			return nil, false
		}
		bn, cont := cp.parseBlock(lastPara)
		if bn != nil {
			rn.Blocks = append(rn.Blocks, bn)
		}
		if !cont {
			lastPara, _ = bn.(*ast.ParaNode)
		}
	}
}

// parseHeading parses a head line.
func (cp *zmkP) parseHeading() (hn *ast.HeadingNode, success bool) {
	inp := cp.inp
	lvl := cp.countDelim(inp.Ch)
	if lvl < 3 {
		return nil, false
	}
	if lvl > 7 {
		lvl = 7
	}
	if inp.Ch != ' ' {
		return nil, false
	}
	inp.Next()
	for inp.Ch == ' ' {
		inp.Next()
	}
	hn = &ast.HeadingNode{Level: lvl - 1}
	for {
		switch inp.Ch {
		case input.EOS, '\n', '\r':
			return hn, true
		}
		in := cp.parseInline()
		if in == nil {
			return hn, true
		}
		if inp.Ch == '{' {
			attrs := cp.parseAttributes(true)
			hn.Attrs = attrs
			inp.SkipToEOL()
			return hn, true
		}
		hn.Inlines = append(hn.Inlines, in)
	}
}

// parseHRule parses a horizontal rule.
func (cp *zmkP) parseHRule() (hn *ast.HRuleNode, success bool) {
	inp := cp.inp
	if cp.countDelim(inp.Ch) < 3 {
		return nil, false
	}
	attrs := cp.parseAttributes(true)
	inp.SkipToEOL()
	return &ast.HRuleNode{Attrs: attrs}, true
}

var mapRuneList = map[rune]ast.ListCode{
	'*': ast.ListUnordered,
	'#': ast.ListOrdered,
	'>': ast.ListQuote,
}

// parseList parses a list.
func (cp *zmkP) parseList() (res ast.BlockNode, success bool) {
	inp := cp.inp
	codes := []ast.ListCode{}
loopInit:
	for {
		code, ok := mapRuneList[inp.Ch]
		if !ok {
			panic(fmt.Sprintf("%q is not a region char", inp.Ch))
		}
		codes = append(codes, code)
		inp.Next()
		switch inp.Ch {
		case '*', '#', '>':
		case ' ':
			break loopInit
		default:
			return nil, false
		}
	}
	for inp.Ch == ' ' {
		inp.Next()
	}
	switch inp.Ch {
	case input.EOS, '\n', '\r':
		return nil, false
	}

	if len(codes) < len(cp.lists) {
		cp.lists = cp.lists[:len(codes)]
	}
	var ln *ast.ListNode
	newLnCount := 0
	for i, code := range codes {
		if i < len(cp.lists) {
			if cp.lists[i].Code != code {
				ln = &ast.ListNode{Code: code}
				newLnCount++
				cp.lists[i] = ln
				cp.lists = cp.lists[:i+1]
			} else {
				ln = cp.lists[i]
			}
		} else {
			ln = &ast.ListNode{Code: code}
			newLnCount++
			cp.lists = append(cp.lists, ln)
		}
	}
	ln.Items = append(ln.Items, ast.ItemSlice{cp.parseLinePara()})
	listDepth := len(cp.lists)
	for i := 0; i < newLnCount; i++ {
		childPos := listDepth - i - 1
		parentPos := childPos - 1
		if parentPos < 0 {
			return cp.lists[0], true
		}
		if prevItems := cp.lists[parentPos].Items; len(prevItems) > 0 {
			lastItem := len(prevItems) - 1
			prevItems[lastItem] = append(prevItems[lastItem], cp.lists[childPos])
		} else {
			cp.lists[parentPos].Items = []ast.ItemSlice{
				ast.ItemSlice{cp.lists[childPos]},
			}
		}
	}
	return nil, true
}

// parseDefTerm parses a term of a definition list.
func (cp *zmkP) parseDefTerm() (res ast.BlockNode, success bool) {
	inp := cp.inp
	inp.Next()
	if inp.Ch != ' ' {
		return nil, false
	}
	inp.Next()
	for inp.Ch == ' ' {
		inp.Next()
	}
	defl := cp.defl
	if defl == nil {
		defl = &ast.DefinitionNode{}
		cp.defl = defl
	}
	defl.Definitions = append(defl.Definitions, ast.Definition{})
	defPos := len(defl.Definitions) - 1
	if defPos == 0 {
		res = defl
	}
	for {
		in := cp.parseInline()
		if in == nil {
			if defl.Definitions[defPos].Term == nil {
				return nil, false
			}
			return res, true
		}
		defl.Definitions[defPos].Term = append(defl.Definitions[defPos].Term, in)
		if _, ok := in.(*ast.BreakNode); ok {
			return res, true
		}
	}
}

// parseDefDescr parses a description of a definition list.
func (cp *zmkP) parseDefDescr() (res ast.BlockNode, success bool) {
	inp := cp.inp
	inp.Next()
	if inp.Ch != ' ' {
		return nil, false
	}
	inp.Next()
	for inp.Ch == ' ' {
		inp.Next()
	}
	defl := cp.defl
	if defl == nil || len(defl.Definitions) == 0 {
		return nil, false
	}
	defPos := len(defl.Definitions) - 1
	if defl.Definitions[defPos].Term == nil {
		return nil, false
	}
	pn := cp.parseLinePara()
	if pn == nil {
		return nil, false
	}
	cp.lists = nil
	cp.table = nil
	defl.Definitions[defPos].Descriptions = append(defl.Definitions[defPos].Descriptions, ast.DescriptionSlice{pn})
	return nil, true
}

// parseIndent parses initial spaces to continue a list.
func (cp *zmkP) parseIndent() (res ast.BlockNode, success bool) {
	inp := cp.inp
	cnt := 0
	for {
		inp.Next()
		if inp.Ch != ' ' {
			break
		}
		cnt++
	}
	if cp.lists != nil {
		// Identation for a list?
		if len(cp.lists) < cnt {
			cnt = len(cp.lists)
		}
		cp.lists = cp.lists[:cnt]
		if cnt == 0 {
			return nil, false
		}
		ln := cp.lists[cnt-1]
		pn := cp.parseLinePara()
		lbn := ln.Items[len(ln.Items)-1]
		if lpn, ok := lbn[len(lbn)-1].(*ast.ParaNode); ok {
			lpn.Inlines = append(lpn.Inlines, pn.Inlines...)
		} else {
			ln.Items[len(ln.Items)-1] = append(ln.Items[len(ln.Items)-1], pn)
		}
		return nil, true
	}
	if cp.defl != nil {
		// Indentation for definition list
		defPos := len(cp.defl.Definitions) - 1
		if cnt < 1 || defPos < 0 {
			return nil, false
		}
		if len(cp.defl.Definitions[defPos].Descriptions) == 0 {
			// Continuation of a definition term
			for {
				in := cp.parseInline()
				if in == nil {
					return nil, true
				}
				cp.defl.Definitions[defPos].Term = append(cp.defl.Definitions[defPos].Term, in)
				if _, ok := in.(*ast.BreakNode); ok {
					return nil, true
				}
			}
		} else {
			// Continuation of a definition description
			pn := cp.parseLinePara()
			if pn == nil {
				return nil, false
			}
			descrPos := len(cp.defl.Definitions[defPos].Descriptions) - 1
			lbn := cp.defl.Definitions[defPos].Descriptions[descrPos]
			if lpn, ok := lbn[len(lbn)-1].(*ast.ParaNode); ok {
				lpn.Inlines = append(lpn.Inlines, pn.Inlines...)
			} else {
				descrPos := len(cp.defl.Definitions[defPos].Descriptions) - 1
				cp.defl.Definitions[defPos].Descriptions[descrPos] = append(cp.defl.Definitions[defPos].Descriptions[descrPos], pn)
			}
			return nil, true
		}
	}
	return nil, false
}

// parseLinePara parses one line of inline material.
func (cp *zmkP) parseLinePara() *ast.ParaNode {
	pn := &ast.ParaNode{}
	for {
		in := cp.parseInline()
		if in == nil {
			if pn.Inlines == nil {
				return nil
			}
			return pn
		}
		pn.Inlines = append(pn.Inlines, in)
		if _, ok := in.(*ast.BreakNode); ok {
			return pn
		}
	}
}

// parseRow parse one table row.
func (cp *zmkP) parseRow() (res ast.BlockNode, success bool) {
	inp := cp.inp
	row := ast.TableRow{}
	for {
		inp.Next()
		cell := cp.parseCell()
		if cell != nil {
			row = append(row, cell)
		}
		switch inp.Ch {
		case '\n', '\r':
			inp.EatEOL()
			fallthrough
		case input.EOS:
			// add to table
			if cp.table == nil {
				cp.table = &ast.TableNode{Rows: []ast.TableRow{row}}
				return cp.table, true
			}
			cp.table.Rows = append(cp.table.Rows, row)
			return nil, true
		}
		// inp.Ch must be '|'
	}
}

// parseCell parses one single cell of a table row.
func (cp *zmkP) parseCell() *ast.TableCell {
	inp := cp.inp
	var slice ast.InlineSlice
	for {
		switch inp.Ch {
		case input.EOS, '\n', '\r':
			if len(slice) == 0 {
				return nil
			}
			fallthrough
		case '|':
			return &ast.TableCell{Inlines: slice}
		}
		slice = append(slice, cp.parseInline())
	}
}
