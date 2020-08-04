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

// Package collect provides functions to collect items from a syntax tree.
package collect

import (
	"zettelstore.de/z/ast"
)

// References returns all references mentioned in the given zettel. This also
// includes references to images.
func References(zettel *ast.Zettel) (links, images []*ast.Reference) {
	lv := &linkVisitor{nil, nil}
	ast.NewTopDownTraverser(lv).VisitBlockSlice(zettel.Ast)
	return lv.links, lv.images
}

type linkVisitor struct {
	links, images []*ast.Reference
}

// VisitZettel does nothing.
func (lv *linkVisitor) VisitZettel(z *ast.Zettel) {}

// VisitVerbatim does nothing.
func (lv *linkVisitor) VisitVerbatim(vn *ast.VerbatimNode) {}

// VisitRegion does nothing.
func (lv *linkVisitor) VisitRegion(rn *ast.RegionNode) {}

// VisitHeading does nothing.
func (lv *linkVisitor) VisitHeading(hn *ast.HeadingNode) {}

// VisitHRule does nothing.
func (lv *linkVisitor) VisitHRule(hn *ast.HRuleNode) {}

// VisitList does nothing.
func (lv *linkVisitor) VisitList(ln *ast.ListNode) {}

// VisitDefinition does nothing.
func (lv *linkVisitor) VisitDefinition(dn *ast.DefinitionNode) {}

// VisitPara does nothing.
func (lv *linkVisitor) VisitPara(pn *ast.ParaNode) {}

// VisitTable does nothing.
func (lv *linkVisitor) VisitTable(tn *ast.TableNode) {}

// VisitBLOB does nothing.
func (lv *linkVisitor) VisitBLOB(bn *ast.BLOBNode) {}

// VisitText does nothing.
func (lv *linkVisitor) VisitText(tn *ast.TextNode) {}

// VisitTag does nothing.
func (lv *linkVisitor) VisitTag(tn *ast.TagNode) {}

// VisitSpace does nothing.
func (lv *linkVisitor) VisitSpace(sn *ast.SpaceNode) {}

// VisitBreak does nothing.
func (lv *linkVisitor) VisitBreak(bn *ast.BreakNode) {}

// VisitLink collects the given link as a reference.
func (lv *linkVisitor) VisitLink(ln *ast.LinkNode) {
	lv.links = append(lv.links, ln.Ref)
}

// VisitImage collects the image links as a reference.
func (lv *linkVisitor) VisitImage(in *ast.ImageNode) {
	if in.Ref != nil {
		lv.images = append(lv.images, in.Ref)
	}
}

// VisitCite does nothing.
func (lv *linkVisitor) VisitCite(cn *ast.CiteNode) {}

// VisitFootnote does nothing.
func (lv *linkVisitor) VisitFootnote(fn *ast.FootnoteNode) {}

// VisitMark does nothing.
func (lv *linkVisitor) VisitMark(mn *ast.MarkNode) {}

// VisitFormat does nothing.
func (lv *linkVisitor) VisitFormat(fn *ast.FormatNode) {}

// VisitEdit does nothing.
func (lv *linkVisitor) VisitEdit(en *ast.EditNode) {}

// VisitLiteral does nothing.
func (lv *linkVisitor) VisitLiteral(ln *ast.LiteralNode) {}
