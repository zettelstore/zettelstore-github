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

// Package ast provides the abstract syntax tree.
package ast

// Visitor is the interface all visitors must implement.
type Visitor interface {
	// Block nodes
	VisitVerbatim(vn *VerbatimNode)
	VisitRegion(rn *RegionNode)
	VisitHeading(hn *HeadingNode)
	VisitHRule(hn *HRuleNode)
	VisitNestedList(ln *NestedListNode)
	VisitDescriptionList(dn *DescriptionListNode)
	VisitPara(pn *ParaNode)
	VisitTable(tn *TableNode)
	VisitBLOB(bn *BLOBNode)

	// Inline nodes
	VisitText(tn *TextNode)
	VisitTag(tn *TagNode)
	VisitSpace(sn *SpaceNode)
	VisitBreak(bn *BreakNode)
	VisitLink(ln *LinkNode)
	VisitImage(in *ImageNode)
	VisitCite(cn *CiteNode)
	VisitFootnote(fn *FootnoteNode)
	VisitMark(mn *MarkNode)
	VisitFormat(fn *FormatNode)
	VisitLiteral(ln *LiteralNode)
}
