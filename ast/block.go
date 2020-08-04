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

// Definition of Block nodes.

// ParaNode contains just a sequence of inline elements.
// Another name is "paragraph".
type ParaNode struct {
	Inlines InlineSlice
}

func (pn *ParaNode) blockNode()       {}
func (pn *ParaNode) itemNode()        {}
func (pn *ParaNode) descriptionNode() {}

// Accept a visitor and visit the node.
func (pn *ParaNode) Accept(v Visitor) { v.VisitPara(pn) }

//--------------------------------------------------------------------------

// VerbatimNode contains lines of uninterpreted text
type VerbatimNode struct {
	Code  VerbatimCode
	Attrs *Attributes
	Lines []string
}

// VerbatimCode specifies the format that is applied to code inline nodes.
type VerbatimCode int

// Constants for VerbatimCode
const (
	_               VerbatimCode = iota
	VerbatimProg                 // Program code.
	VerbatimComment              // Block comment
	VerbatimHTML                 // Block HTML, e.g. for Markdown
)

func (vn *VerbatimNode) blockNode() {}
func (vn *VerbatimNode) itemNode()  {}

// Accept a visitor an visit the node.
func (vn *VerbatimNode) Accept(v Visitor) { v.VisitVerbatim(vn) }

//--------------------------------------------------------------------------

// RegionNode encapsulates a region of block nodes.
type RegionNode struct {
	Code    RegionCode
	Attrs   *Attributes
	Blocks  BlockSlice
	Inlines InlineSlice // Additional text at the end of the region
}

// RegionCode specifies the actual region type.
type RegionCode int

// Values for RegionCode
const (
	_           RegionCode = iota
	RegionSpan             // Just a span of blocks
	RegionQuote            // A longer quotation
	RegionVerse            // Line breaks matter
)

func (rn *RegionNode) blockNode() {}
func (rn *RegionNode) itemNode()  {}

// Accept a visitor and visit the node.
func (rn *RegionNode) Accept(v Visitor) { v.VisitRegion(rn) }

//--------------------------------------------------------------------------

// HeadingNode stores the heading text and level.
type HeadingNode struct {
	Level   int
	Inlines InlineSlice // Heading text, possibly formatted
	Attrs   *Attributes
}

func (hn *HeadingNode) blockNode() {}
func (hn *HeadingNode) itemNode()  {}

// Accept a visitor and visit the node.
func (hn *HeadingNode) Accept(v Visitor) { v.VisitHeading(hn) }

//--------------------------------------------------------------------------

// HRuleNode specifies a horizontal rule.
type HRuleNode struct {
	Attrs *Attributes
}

func (hn *HRuleNode) blockNode() {}
func (hn *HRuleNode) itemNode()  {}

// Accept a visitor and visit the node.
func (hn *HRuleNode) Accept(v Visitor) { v.VisitHRule(hn) }

//--------------------------------------------------------------------------

// ListNode specifies a list, either ordered or unordered.
type ListNode struct {
	Code  ListCode
	Items []ItemSlice
	Attrs *Attributes
}

// ListCode specifies the actual list type.
type ListCode int

// Values for ListCode
const (
	_             ListCode = iota
	ListOrdered            // Ordered list.
	ListUnordered          // Unordered list.
	ListQuote              // Quote list.
)

func (ln *ListNode) blockNode() {}
func (ln *ListNode) itemNode()  {}

// Accept a visitor and visit the node.
func (ln *ListNode) Accept(v Visitor) { v.VisitList(ln) }

//--------------------------------------------------------------------------

// DefinitionNode specifies a definition list.
type DefinitionNode struct {
	Definitions []Definition
}

// Definition is one element of a definition list.
type Definition struct {
	Term         InlineSlice
	Descriptions []DescriptionSlice
}

func (dn *DefinitionNode) blockNode() {}

// Accept a visitor and visit the node.
func (dn *DefinitionNode) Accept(v Visitor) { v.VisitDefinition(dn) }

//--------------------------------------------------------------------------

// TableNode specifies a full table
type TableNode struct {
	Header TableRow    // The header row
	Align  []Alignment // Default column alignment
	Rows   []TableRow  // The slice of cell rows
}

// TableCell contains the data for one table cell
type TableCell struct {
	Align   Alignment   // Cell alignment
	Inlines InlineSlice // Cell content
}

// TableRow is a slice of cells.
type TableRow []*TableCell

// Alignment specifies text alignment.
// Currently only for tables.
type Alignment int

// Constants for Alignment.
const (
	_            Alignment = iota
	AlignDefault           // Default alignment, inherited
	AlignLeft              // Left alignment
	AlignCenter            // Center the content
	AlignRight             // Right alignment
)

func (tn *TableNode) blockNode() {}

// Accept a visitor and visit the node.
func (tn *TableNode) Accept(v Visitor) { v.VisitTable(tn) }

//--------------------------------------------------------------------------

// BLOBNode contains just binary data that must be interpreted accordung to
// a syntax.
type BLOBNode struct {
	Title  string
	Syntax string
	Blob   []byte
}

func (bn *BLOBNode) blockNode() {}

// Accept a visitor and visit the node.
func (bn *BLOBNode) Accept(v Visitor) { v.VisitBLOB(bn) }
