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
	"strings"

	"zettelstore.de/z/ast"
)

// postProcessBlocks is the entry point for post-processing a list of block nodes.
func postProcessBlocks(bs ast.BlockSlice) ast.BlockSlice {
	pp := postProcessor{marker: map[string]ast.Node{}}
	return pp.processBlockSlice(bs)
}

// postProcessInlines is the entry point for post-processing a list of inline nodes.
func postProcessInlines(is ast.InlineSlice) ast.InlineSlice {
	pp := postProcessor{marker: map[string]ast.Node{}}
	return pp.processInlineSlice(is)
}

// postProcessor is a visitor that cleans the abstract syntax tree.
type postProcessor struct {
	marker  map[string]ast.Node
	inVerse bool
}

// VisitPara post-processes a paragraph.
func (pp *postProcessor) VisitPara(pn *ast.ParaNode) {
	pn.Inlines = pp.processInlineSlice(pn.Inlines)
}

// VisitVerbatim post-processes a verbatim block.
func (pp *postProcessor) VisitVerbatim(vn *ast.VerbatimNode) {}

// VisitRegion post-processes a region.
func (pp *postProcessor) VisitRegion(rn *ast.RegionNode) {
	oldVerse := pp.inVerse
	if rn.Code == ast.RegionVerse {
		pp.inVerse = true
	}
	rn.Blocks = pp.processBlockSlice(rn.Blocks)
	pp.inVerse = oldVerse
	rn.Inlines = pp.processInlineSlice(rn.Inlines)
}

// VisitHeading post-processes a heading.
func (pp *postProcessor) VisitHeading(hn *ast.HeadingNode) {
	hn.Inlines = pp.processInlineSlice(hn.Inlines)
}

// VisitHRule post-processes a horizontal rule.
func (pp *postProcessor) VisitHRule(hn *ast.HRuleNode) {}

// VisitList post-processes a list.
func (pp *postProcessor) VisitList(ln *ast.ListNode) {
	for i, item := range ln.Items {
		ln.Items[i] = pp.processItemSlice(item)
	}
}

// VisitDefinition post-processes a definition list.
func (pp *postProcessor) VisitDefinition(dn *ast.DefinitionNode) {
	for i, def := range dn.Definitions {
		dn.Definitions[i].Term = pp.processInlineSlice(def.Term)
		for j, b := range def.Descriptions {
			dn.Definitions[i].Descriptions[j] = pp.processDescriptionSlice(b)
		}
	}
}

// VisitTable post-processes a table.
func (pp *postProcessor) VisitTable(tn *ast.TableNode) {
	width := tableWidth(tn)
	tn.Align = make([]ast.Alignment, 0, width)
	if len(tn.Rows) > 0 && isHeaderRow(tn.Rows[0]) {
		tn.Header = tn.Rows[0]
		tn.Rows = tn.Rows[1:]
		for _, cell := range tn.Header {
			if inlines := cell.Inlines; len(inlines) > 0 {
				if textNode, ok := inlines[0].(*ast.TextNode); ok {
					if strings.HasPrefix(textNode.Text, "=") {
						textNode.Text = textNode.Text[1:]
					}
				}
				if textNode, ok := inlines[len(inlines)-1].(*ast.TextNode); ok {
					if tnl := len(textNode.Text); tnl > 0 {
						align := getAlignment(textNode.Text[tnl-1])
						if align != ast.AlignDefault {
							textNode.Text = textNode.Text[0 : tnl-1]
						}
						tn.Align = append(tn.Align, align)
					}
				}
			}
		}
	}
	for len(tn.Align) < width {
		tn.Align = append(tn.Align, ast.AlignDefault)
	}
	if len(tn.Header) > 0 {
		tn.Header = appendCells(tn.Header, width, tn.Align)
		for i, cell := range tn.Header {
			pp.processCell(cell, tn.Align[i])
		}
	}
	for i, row := range tn.Rows {
		tn.Rows[i] = appendCells(row, width, tn.Align)
		row = tn.Rows[i]
		for i, cell := range row {
			pp.processCell(cell, tn.Align[i])
		}
	}
}

func tableWidth(tn *ast.TableNode) int {
	width := 0
	for _, row := range tn.Rows {
		if width < len(row) {
			width = len(row)
		}
	}
	return width
}

func appendCells(row ast.TableRow, width int, colAlign []ast.Alignment) ast.TableRow {
	for len(row) < width {
		row = append(row, &ast.TableCell{Align: colAlign[len(row)]})
	}
	return row
}

func isHeaderRow(row ast.TableRow) bool {
	for i := 0; i < len(row); i++ {
		if inlines := row[i].Inlines; len(inlines) > 0 {
			if textNode, ok := inlines[0].(*ast.TextNode); ok {
				if strings.HasPrefix(textNode.Text, "=") {
					return true
				}
			}
		}
	}
	return false
}

func getAlignment(ch byte) ast.Alignment {
	switch ch {
	case ':':
		return ast.AlignCenter
	case '<':
		return ast.AlignLeft
	case '>':
		return ast.AlignRight
	default:
		return ast.AlignDefault
	}
}

// processCell tries to recognize cell formatting.
func (pp *postProcessor) processCell(cell *ast.TableCell, colAlign ast.Alignment) {
	if len(cell.Inlines) == 0 {
		return
	}
	if textNode, ok := cell.Inlines[0].(*ast.TextNode); ok && len(textNode.Text) > 0 {
		align := getAlignment(textNode.Text[0])
		if align == ast.AlignDefault {
			cell.Align = colAlign
		} else {
			textNode.Text = textNode.Text[1:]
			cell.Align = align
		}
	} else {
		cell.Align = colAlign
	}
	cell.Inlines = pp.processInlineSlice(cell.Inlines)
}

// VisitBLOB does nothing.
func (pp *postProcessor) VisitBLOB(bn *ast.BLOBNode) {}

// VisitText does nothing.
func (pp *postProcessor) VisitText(tn *ast.TextNode) {}

// VisitTag does nothing.
func (pp *postProcessor) VisitTag(tn *ast.TagNode) {}

// VisitSpace does nothing.
func (pp *postProcessor) VisitSpace(sn *ast.SpaceNode) {}

// VisitBreak does nothing.
func (pp *postProcessor) VisitBreak(bn *ast.BreakNode) {}

// VisitLink post-processes a link.
func (pp *postProcessor) VisitLink(ln *ast.LinkNode) {
	ln.Inlines = pp.processInlineSlice(ln.Inlines)
}

// VisitImage post-processes an image.
func (pp *postProcessor) VisitImage(in *ast.ImageNode) {
	if len(in.Inlines) > 0 {
		in.Inlines = pp.processInlineSlice(in.Inlines)
	}
}

// VisitCite post-processes a citation.
func (pp *postProcessor) VisitCite(cn *ast.CiteNode) {
	cn.Inlines = pp.processInlineSlice(cn.Inlines)
}

// VisitFootnote post-processes a footnote.
func (pp *postProcessor) VisitFootnote(fn *ast.FootnoteNode) {
	fn.Inlines = pp.processInlineSlice(fn.Inlines)
}

// VisitMark post-processes a mark.
// It checks that there is no duplicate mark text. The first of such mark will
// survive, all following marks will have a delete mark text.
func (pp *postProcessor) VisitMark(mn *ast.MarkNode) {
	if len(mn.Text) > 0 {
		s := string(mn.Text)
		if n, ok := pp.marker[s]; ok && n != mn {
			mn.Text = ""
		} else {
			pp.marker[s] = mn
		}
	}
}

// VisitFormat post-processes formatted inline nodes.
func (pp *postProcessor) VisitFormat(fn *ast.FormatNode) {
	if fn.Attrs != nil {
		if fn.Attrs.HasDefault() {
			switch fn.Code {
			case ast.FormatItalic:
				fn.Code = ast.FormatEmph
				fn.Attrs.RemoveDefault()
			case ast.FormatBold:
				fn.Code = ast.FormatStrong
				fn.Attrs.RemoveDefault()
			}
		}
	}
	fn.Inlines = pp.processInlineSlice(fn.Inlines)
}

// VisitEdit post-processes an edit inline.
func (pp *postProcessor) VisitEdit(en *ast.EditNode) {
	en.Deletes = pp.processInlineSlice(en.Deletes)
	en.Inserts = pp.processInlineSlice(en.Inserts)
}

// VisitLiteral post-processes an inline literal.
func (pp *postProcessor) VisitLiteral(cn *ast.LiteralNode) {}

// processBlockSlice post-processes a slice of blocks.
// It is one of the working horses for post-processing.
func (pp *postProcessor) processBlockSlice(bns ast.BlockSlice) ast.BlockSlice {
	for _, bn := range bns {
		bn.Accept(pp)
	}
	fromPos, toPos := 0, 0
	for fromPos < len(bns) {
		bns[toPos] = bns[fromPos]
		fromPos++
		switch bn := bns[toPos].(type) {
		case *ast.ParaNode:
			if len(bn.Inlines) > 0 {
				toPos++
			}
		case *nullItemNode:
		case *nullDescriptionNode:
		default:
			toPos++
		}
	}
	for pos := toPos; pos < len(bns); pos++ {
		bns[pos] = nil // Allow excess nodes to be garbage collected.
	}
	return bns[:toPos:toPos]
}

// processItemSlice post-processes a slice of items.
// It is one of the working horses for post-processing.
func (pp *postProcessor) processItemSlice(ins ast.ItemSlice) ast.ItemSlice {
	for _, in := range ins {
		in.Accept(pp)
	}
	fromPos, toPos := 0, 0
	for fromPos < len(ins) {
		ins[toPos] = ins[fromPos]
		fromPos++
		switch in := ins[toPos].(type) {
		case *ast.ParaNode:
			if len(in.Inlines) > 0 {
				toPos++
			}
		case *nullItemNode:
		case *nullDescriptionNode:
		default:
			toPos++
		}
	}
	for pos := toPos; pos < len(ins); pos++ {
		ins[pos] = nil // Allow excess nodes to be garbage collected.
	}
	return ins[:toPos:toPos]
}

// processDescriptionSlice post-processes a slice of descriptions.
// It is one of the working horses for post-processing.
func (pp *postProcessor) processDescriptionSlice(dns ast.DescriptionSlice) ast.DescriptionSlice {
	for _, dn := range dns {
		dn.Accept(pp)
	}
	fromPos, toPos := 0, 0
	for fromPos < len(dns) {
		dns[toPos] = dns[fromPos]
		fromPos++
		switch dn := dns[toPos].(type) {
		case *ast.ParaNode:
			if len(dn.Inlines) > 0 {
				toPos++
			}
		case *nullDescriptionNode:
		default:
			toPos++
		}
	}
	for pos := toPos; pos < len(dns); pos++ {
		dns[pos] = nil // Allow excess nodes to be garbage collected.
	}
	return dns[:toPos:toPos]
}

// processInlineSlice post-processes a slice of inline nodes.
// It is one of the working horses for post-processing.
func (pp *postProcessor) processInlineSlice(ins ast.InlineSlice) ast.InlineSlice {
	if len(ins) == 0 {
		return nil
	}
	for _, in := range ins {
		in.Accept(pp)
	}

	if !pp.inVerse {
		ins = processInlineSliceHead(ins)
	}
	toPos := pp.processInlineSliceCopy(ins)
	toPos = pp.processInlineSliceTail(ins, toPos)
	ins = ins[:toPos:toPos]
	pp.processInlineSliceInplace(ins)
	return ins
}

// processInlineSliceHead removes leading spaces and empty text.
func processInlineSliceHead(ins ast.InlineSlice) ast.InlineSlice {
	for i := 0; i < len(ins); i++ {
		switch in := ins[i].(type) {
		case *ast.SpaceNode:
		case *ast.TextNode:
			if len(in.Text) > 0 {
				return ins[i:]
			}
		default:
			return ins[i:]
		}
	}
	return ins[0:0]
}

// processInlineSliceCopy goes forward through the slice and tries to eliminate
// elements that follow the current element.
//
// Two text nodes are merged into one.
//
// Two spaces following a break are merged into a hard break.
func (pp *postProcessor) processInlineSliceCopy(ins ast.InlineSlice) int {
	maxPos := len(ins)
	for {
		again := false
		fromPos, toPos := 0, 0
		for fromPos < maxPos {
			ins[toPos] = ins[fromPos]
			fromPos++
			switch in := ins[toPos].(type) {
			case *ast.TextNode:
				for fromPos < maxPos {
					if tn, ok := ins[fromPos].(*ast.TextNode); ok {
						in.Text = in.Text + tn.Text
						fromPos++
					} else {
						break
					}
				}
			case *ast.SpaceNode:
				if fromPos < maxPos {
					switch nn := ins[fromPos].(type) {
					case *ast.BreakNode:
						if len(in.Lexeme) > 1 {
							nn.Hard = true
							ins[toPos] = nn
							fromPos++
						}
					case *ast.TextNode:
						if pp.inVerse {
							ins[toPos] = &ast.TextNode{Text: strings.Repeat(string(160), len(in.Lexeme)) + nn.Text}
							fromPos++
							again = true
						}
					}
				}
			case *ast.BreakNode:
				if pp.inVerse {
					in.Hard = true
				}
			}
			toPos++
		}
		for pos := toPos; pos < maxPos; pos++ {
			ins[pos] = nil // Allow excess nodes to be garbage collected.
		}
		if !again {
			return toPos
		}
		maxPos = toPos
	}
}

// processInlineSliceTail removes empty text nodes, breaks and spaces at the end.
func (pp *postProcessor) processInlineSliceTail(ins ast.InlineSlice, toPos int) int {
	for toPos > 0 {
		switch n := ins[toPos-1].(type) {
		case *ast.TextNode:
			if len(n.Text) > 0 {
				return toPos
			}
		case *ast.BreakNode:
		case *ast.SpaceNode:
		default:
			return toPos
		}
		toPos--
		ins[toPos] = nil // Kill node to enable garbage collection
	}
	return toPos
}

func (pp *postProcessor) processInlineSliceInplace(ins ast.InlineSlice) {
	for _, in := range ins {
		switch n := in.(type) {
		case *ast.TextNode:
			if n.Text == "..." {
				n.Text = "\u2026"
			} else if len(n.Text) == 4 && strings.IndexByte(",;:!?", n.Text[3]) >= 0 && n.Text[:3] == "..." {
				n.Text = "\u2026" + n.Text[3:]
			}
		}
	}
}
