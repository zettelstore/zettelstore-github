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

// Package jsonenc encodes the abstract syntax tree into JSON.
package jsonenc

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strconv"

	"zettelstore.de/z/ast"
	"zettelstore.de/z/domain"
	"zettelstore.de/z/encoder"
)

func init() {
	encoder.Register("json", createEncoder)
}

func createEncoder() encoder.Encoder {
	return &jsonEncoder{}
}

type jsonEncoder struct {
	adaptLink  func(*ast.LinkNode) ast.InlineNode
	adaptImage func(*ast.ImageNode) ast.InlineNode
	meta       *domain.Meta
}

// SetOption sets an option for the encoder
func (je *jsonEncoder) SetOption(option encoder.Option) {
	switch opt := option.(type) {
	case *encoder.MetaOption:
		je.meta = opt.Meta
	case *encoder.AdaptLinkOption:
		je.adaptLink = opt.Adapter
	case *encoder.AdaptImageOption:
		je.adaptImage = opt.Adapter
	}
}

// WriteZettel writes the encoded zettel to the writer.
func (je *jsonEncoder) WriteZettel(w io.Writer, zettel *ast.Zettel) (int, error) {
	v := newVisitor(w, je)
	v.b.WriteByte('{')
	v.b.WriteString("\"title\":")
	v.acceptInlineSlice(zettel.Title)
	if je.meta != nil {
		v.acceptMeta(je.meta, false)
	} else {
		v.acceptMeta(zettel.Meta, false)
	}
	v.b.WriteString(",\"content\":")
	v.acceptBlockSlice(zettel.Ast)
	v.b.WriteByte('}')
	length, err := v.b.Flush()
	return length, err
}

// WriteMeta encodes meta data as HTML5.
func (je *jsonEncoder) WriteMeta(w io.Writer, meta *domain.Meta) (int, error) {
	v := newVisitor(w, je)
	v.b.WriteByte('{')
	v.acceptMeta(meta, true)
	v.b.WriteByte('}')
	length, err := v.b.Flush()
	return length, err
}

// WriteBlocks writes a block slice to the writer
func (je *jsonEncoder) WriteBlocks(w io.Writer, bs ast.BlockSlice) (int, error) {
	v := newVisitor(w, je)
	v.acceptBlockSlice(bs)
	length, err := v.b.Flush()
	return length, err
}

// WriteInlines writes an inline slice to the writer
func (je *jsonEncoder) WriteInlines(w io.Writer, is ast.InlineSlice) (int, error) {
	v := newVisitor(w, je)
	v.acceptInlineSlice(is)
	length, err := v.b.Flush()
	return length, err
}

// visitor writes the abstract syntax tree to an io.Writer.
type visitor struct {
	b   encoder.BufWriter
	enc *jsonEncoder
}

func newVisitor(w io.Writer, je *jsonEncoder) *visitor {
	return &visitor{b: encoder.NewBufWriter(w), enc: je}
}

func (v *visitor) acceptMeta(meta *domain.Meta, withTitle bool) {
	if withTitle {
		v.b.WriteString("\"title\":\"")
		v.b.Write(Escape(meta.GetDefault(domain.MetaKeyTitle, "")))
		v.b.WriteByte('"')
	}
	v.writeMetaList(meta, domain.MetaKeyTags, "tags")
	v.writeMetaString(meta, domain.MetaKeySyntax, "syntax")
	v.writeMetaString(meta, domain.MetaKeyRole, "role")
	if pairs := meta.PairsRest(); len(pairs) > 0 {
		v.b.WriteString(",\"header\":{\"")
		first := true
		for _, p := range pairs {
			if !first {
				v.b.WriteString("\",\"")
			}
			v.b.Write(Escape(p.Key))
			v.b.WriteString("\":\"")
			v.b.Write(Escape(p.Value))
			first = false
		}
		v.b.WriteString("\"}")
	}
}

func (v *visitor) writeMetaString(meta *domain.Meta, key string, native string) {
	if val, ok := meta.Get(key); ok && len(val) > 0 {
		v.b.WriteString(",\"")
		v.b.Write(Escape(native))
		v.b.WriteString("\":\"")
		v.b.Write(Escape(val))
		v.b.WriteByte('"')
	}
}

func (v *visitor) writeMetaList(meta *domain.Meta, key string, native string) {
	if vals, ok := meta.GetList(key); ok && len(vals) > 0 {
		v.b.WriteString(",\"")
		v.b.Write(Escape(native))
		v.b.WriteString("\":[\"")
		for i, val := range vals {
			if i > 0 {
				v.b.WriteString("\",\"")
			}
			v.b.Write(Escape(val))
		}
		v.b.WriteString("\"]")
	}
}

// VisitPara emits JSON code for a paragraph.
func (v *visitor) VisitPara(pn *ast.ParaNode) {
	v.writeNodeStart("Para")
	v.writeContentStart('i')
	v.acceptInlineSlice(pn.Inlines)
	v.b.WriteByte('}')
}

var verbatimCode = map[ast.VerbatimCode]string{
	ast.VerbatimProg:    "CodeBlock",
	ast.VerbatimComment: "CommentBlock",
	ast.VerbatimHTML:    "HTMLBlock",
}

// VisitVerbatim emits JSON code for verbatim lines.
func (v *visitor) VisitVerbatim(vn *ast.VerbatimNode) {
	code, ok := verbatimCode[vn.Code]
	if !ok {
		panic(fmt.Sprintf("Unknown verbatim code %v", vn.Code))
	}
	v.writeNodeStart(code)
	v.visitAttributes(vn.Attrs)
	v.writeContentStart('l')
	for i, line := range vn.Lines {
		if i > 0 {
			v.b.WriteByte(',')
		}
		v.writeEscaped(line)
	}
	v.b.WriteString("]}")
}

var regionCode = map[ast.RegionCode]string{
	ast.RegionSpan:  "SpanBlock",
	ast.RegionQuote: "QuoteBlock",
	ast.RegionVerse: "VerseBlock",
}

// VisitRegion writes JSON code for block regions.
func (v *visitor) VisitRegion(rn *ast.RegionNode) {
	code, ok := regionCode[rn.Code]
	if !ok {
		panic(fmt.Sprintf("Unknown region code %v", rn.Code))
	}
	v.writeNodeStart(code)
	v.visitAttributes(rn.Attrs)
	v.writeContentStart('b')
	v.acceptBlockSlice(rn.Blocks)
	if len(rn.Inlines) > 0 {
		v.writeContentStart('i')
		v.acceptInlineSlice(rn.Inlines)
	}
	v.b.WriteByte('}')
}

// VisitHeading writes the JSON code for a heading.
func (v *visitor) VisitHeading(hn *ast.HeadingNode) {
	v.writeNodeStart("Heading")
	v.visitAttributes(hn.Attrs)
	v.writeContentStart('n')
	v.b.WriteString(strconv.Itoa(hn.Level))
	v.writeContentStart('i')
	v.acceptInlineSlice(hn.Inlines)
	v.b.WriteByte('}')
}

// VisitHRule writes JSON code for a horizontal rule: <hr>.
func (v *visitor) VisitHRule(hn *ast.HRuleNode) {
	v.writeNodeStart("Hrule")
	v.visitAttributes(hn.Attrs)
	v.b.WriteByte('}')
}

var listCode = map[ast.NestedListCode]string{
	ast.NestedListOrdered:   "OrderedList",
	ast.NestedListUnordered: "BulletList",
	ast.NestedListQuote:     "QuoteList",
}

// VisitNestedList writes JSON code for lists and blockquotes.
func (v *visitor) VisitNestedList(ln *ast.NestedListNode) {
	v.writeNodeStart(listCode[ln.Code])
	v.writeContentStart('c')
	for i, item := range ln.Items {
		if i > 0 {
			v.b.WriteByte(',')
		}
		v.acceptItemSlice(item)
	}
	v.b.WriteString("]}")
}

// VisitDescriptionList emits a JSON description list.
func (v *visitor) VisitDescriptionList(dn *ast.DescriptionListNode) {
	v.writeNodeStart("DescriptionList")
	v.writeContentStart('g')
	for i, def := range dn.Descriptions {
		if i > 0 {
			v.b.WriteByte(',')
		}
		v.b.WriteByte('[')
		v.acceptInlineSlice(def.Term)

		if len(def.Descriptions) > 0 {
			for _, b := range def.Descriptions {
				v.b.WriteByte(',')
				v.acceptDescriptionSlice(b)
			}
		}
		v.b.WriteByte(']')
	}
	v.b.WriteString("]}")
}

// VisitTable emits a JSON table.
func (v *visitor) VisitTable(tn *ast.TableNode) {
	v.writeNodeStart("Table")
	v.writeContentStart('p')

	// Table header
	v.b.WriteByte('[')
	for i, cell := range tn.Header {
		if i > 0 {
			v.b.WriteByte(',')
		}
		v.writeCell(cell)
	}
	v.b.WriteString("],")

	// Table rows
	v.b.WriteByte('[')
	for i, row := range tn.Rows {
		if i > 0 {
			v.b.WriteByte(',')
		}
		v.b.WriteByte('[')
		for j, cell := range row {
			if j > 0 {
				v.b.WriteByte(',')
			}
			v.writeCell(cell)
		}
		v.b.WriteByte(']')
	}
	v.b.WriteString("]]}")
}

var alignmentCode = map[ast.Alignment]string{
	ast.AlignDefault: "[\"\",",
	ast.AlignLeft:    "[\"<\",",
	ast.AlignCenter:  "[\":\",",
	ast.AlignRight:   "[\">\",",
}

func (v *visitor) writeCell(cell *ast.TableCell) {
	v.b.WriteString(alignmentCode[cell.Align])
	v.acceptInlineSlice(cell.Inlines)
	v.b.WriteByte(']')
}

// VisitBLOB writes the binary object as a value.
func (v *visitor) VisitBLOB(bn *ast.BLOBNode) {
	v.writeNodeStart("Blob")
	v.writeContentStart('q')
	v.writeEscaped(bn.Title)
	v.writeContentStart('s')
	v.writeEscaped(bn.Syntax)
	v.writeContentStart('o')
	v.b.WriteBase64(bn.Blob)
	v.b.WriteString("\"}")
}

// VisitText writes text content.
func (v *visitor) VisitText(tn *ast.TextNode) {
	v.writeNodeStart("Text")
	v.writeContentStart('s')
	v.writeEscaped(tn.Text)
	v.b.WriteByte('}')
}

// VisitTag writes tag content.
func (v *visitor) VisitTag(tn *ast.TagNode) {
	v.writeNodeStart("Tag")
	v.writeContentStart('s')
	v.writeEscaped(tn.Tag)
	v.b.WriteByte('}')
}

// VisitSpace emits a white space.
func (v *visitor) VisitSpace(sn *ast.SpaceNode) {
	v.writeNodeStart("Space")
	if l := len(sn.Lexeme); l > 1 {
		v.writeContentStart('n')
		v.b.WriteString(strconv.Itoa(l))
	}
	v.b.WriteByte('}')
}

// VisitBreak writes JSON code for line breaks.
func (v *visitor) VisitBreak(bn *ast.BreakNode) {
	if bn.Hard {
		v.writeNodeStart("Hard")
	} else {
		v.writeNodeStart("Soft")
	}
	v.b.WriteByte('}')
}

var mapRefState = map[ast.RefState]string{
	ast.RefStateInvalid:      "invalid",
	ast.RefStateZettel:       "zettel",
	ast.RefStateZettelFound:  "zettel",
	ast.RefStateZettelBroken: "broken",
	ast.RefStateMaterial:     "material",
}

// VisitLink writes JSON code for links.
func (v *visitor) VisitLink(ln *ast.LinkNode) {
	if adapt := v.enc.adaptLink; adapt != nil {
		n := adapt(ln)
		var ok bool
		if ln, ok = n.(*ast.LinkNode); !ok {
			n.Accept(v)
			return
		}
	}
	v.writeNodeStart("Link")
	v.visitAttributes(ln.Attrs)
	v.writeContentStart('q')
	v.writeEscaped(mapRefState[ln.Ref.State])
	v.writeContentStart('s')
	v.writeEscaped(ln.Ref.String())
	v.writeContentStart('i')
	v.acceptInlineSlice(ln.Inlines)
	v.b.WriteByte('}')
}

// VisitImage writes JSON code for images.
func (v *visitor) VisitImage(in *ast.ImageNode) {
	if adapt := v.enc.adaptImage; adapt != nil {
		n := adapt(in)
		var ok bool
		if in, ok = n.(*ast.ImageNode); !ok {
			n.Accept(v)
			return
		}
	}
	v.writeNodeStart("Image")
	v.visitAttributes(in.Attrs)
	if in.Ref == nil {
		v.writeContentStart('j')
		v.b.WriteString("\"s\":")
		v.writeEscaped(in.Syntax)
		switch in.Syntax {
		case "svg":
			v.writeContentStart('q')
			v.writeEscaped(string(in.Blob))
		default:
			v.writeContentStart('o')
			v.b.WriteBase64(in.Blob)
			v.b.WriteByte('"')
		}
		v.b.WriteByte('}')
	} else {
		v.writeContentStart('s')
		v.writeEscaped(in.Ref.String())
	}
	if len(in.Inlines) > 0 {
		v.writeContentStart('i')
		v.acceptInlineSlice(in.Inlines)
	}
	v.b.WriteByte('}')
}

// VisitCite writes code for citations.
func (v *visitor) VisitCite(cn *ast.CiteNode) {
	v.writeNodeStart("Cite")
	v.visitAttributes(cn.Attrs)
	v.writeContentStart('s')
	v.writeEscaped(cn.Key)
	if len(cn.Inlines) > 0 {
		v.writeContentStart('i')
		v.acceptInlineSlice(cn.Inlines)
	}
	v.b.WriteByte('}')
}

// VisitFootnote write JSON code for a footnote.
func (v *visitor) VisitFootnote(fn *ast.FootnoteNode) {
	v.writeNodeStart("Footnote")
	v.visitAttributes(fn.Attrs)
	v.writeContentStart('i')
	v.acceptInlineSlice(fn.Inlines)
	v.b.WriteByte('}')
}

// VisitMark writes JSON code to mark a position.
func (v *visitor) VisitMark(mn *ast.MarkNode) {
	v.writeNodeStart("Mark")
	if len(mn.Text) > 0 {
		v.writeContentStart('s')
		v.writeEscaped(mn.Text)
	}
	v.b.WriteByte('}')
}

var formatCode = map[ast.FormatCode]string{
	ast.FormatItalic:    "Italic",
	ast.FormatEmph:      "Emph",
	ast.FormatBold:      "Bold",
	ast.FormatStrong:    "Strong",
	ast.FormatMonospace: "Mono",
	ast.FormatStrike:    "Strikethrough",
	ast.FormatDelete:    "Delete",
	ast.FormatUnder:     "Underline",
	ast.FormatInsert:    "Insert",
	ast.FormatSuper:     "Super",
	ast.FormatSub:       "Sub",
	ast.FormatQuote:     "Quote",
	ast.FormatQuotation: "Quotation",
	ast.FormatSmall:     "Small",
	ast.FormatSpan:      "Span",
}

// VisitFormat write JSON code for formatting text.
func (v *visitor) VisitFormat(fn *ast.FormatNode) {
	v.writeNodeStart(formatCode[fn.Code])
	v.visitAttributes(fn.Attrs)
	v.writeContentStart('i')
	v.acceptInlineSlice(fn.Inlines)
	v.b.WriteByte('}')
}

var literalCode = map[ast.LiteralCode]string{
	ast.LiteralProg:    "Code",
	ast.LiteralKeyb:    "Input",
	ast.LiteralOutput:  "Output",
	ast.LiteralComment: "Comment",
	ast.LiteralHTML:    "HTML",
}

// VisitLiteral write JSON code for literal inline text.
func (v *visitor) VisitLiteral(ln *ast.LiteralNode) {
	code, ok := literalCode[ln.Code]
	if !ok {
		panic(fmt.Sprintf("Unknown literal code %v", ln.Code))
	}
	v.writeNodeStart(code)
	v.visitAttributes(ln.Attrs)
	v.writeContentStart('s')
	v.writeEscaped(ln.Text)
	v.b.WriteByte('}')
}

func (v *visitor) acceptBlockSlice(bns ast.BlockSlice) {
	v.b.WriteByte('[')
	for i, bn := range bns {
		if i > 0 {
			v.b.WriteByte(',')
		}
		bn.Accept(v)
	}
	v.b.WriteByte(']')
}

func (v *visitor) acceptItemSlice(ins ast.ItemSlice) {
	v.b.WriteByte('[')
	for i, in := range ins {
		if i > 0 {
			v.b.WriteByte(',')
		}
		in.Accept(v)
	}
	v.b.WriteByte(']')
}

func (v *visitor) acceptDescriptionSlice(dns ast.DescriptionSlice) {
	v.b.WriteByte('[')
	for i, dn := range dns {
		if i > 0 {
			v.b.WriteByte(',')
		}
		dn.Accept(v)
	}
	v.b.WriteByte(']')
}

func (v *visitor) acceptInlineSlice(ins ast.InlineSlice) {
	v.b.WriteByte('[')
	for i, in := range ins {
		if i > 0 {
			v.b.WriteByte(',')
		}
		in.Accept(v)
	}
	v.b.WriteByte(']')
}

// visitAttributes write JSON attributes
func (v *visitor) visitAttributes(a *ast.Attributes) {
	if a == nil || len(a.Attrs) == 0 {
		return
	}
	keys := make([]string, 0, len(a.Attrs))
	for k := range a.Attrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	v.b.WriteString(",\"a\":{\"")
	for i, k := range keys {
		if i > 0 {
			v.b.WriteString("\",\"")
		}
		v.b.Write(Escape(k))
		v.b.WriteString("\":\"")
		v.b.Write(Escape(a.Attrs[k]))
	}
	v.b.WriteString("\"}")
}

func (v *visitor) writeNodeStart(t string) {
	v.b.WriteStrings("{\"t\":\"", t, "\"")
}

var contentCode = map[rune][]byte{
	'b': []byte(",\"b\":"),   // List of blocks
	'c': []byte(",\"c\":["),  // List of list of blocks
	'g': []byte(",\"g\":["),  // General list
	'i': []byte(",\"i\":"),   // List of inlines
	'j': []byte(",\"j\":{"),  // Embedded JSON object
	'l': []byte(",\"l\":["),  // List of lines
	'n': []byte(",\"n\":"),   // Number
	'o': []byte(",\"o\":\""), // Byte object
	'p': []byte(",\"p\":["),  // Generic tuple
	'q': []byte(",\"q\":"),   // String, if 's' is also needed
	's': []byte(",\"s\":"),   // String
	't': []byte("Content code 't' is not allowed"),
	'y': []byte("Content code 'y' is not allowed"), // field after 'j'
}

func (v *visitor) writeContentStart(code rune) {
	if b, ok := contentCode[code]; ok {
		v.b.Write(b)
		return
	}
	panic("Unknown content code " + strconv.Itoa(int(code)))
}

var (
	jsBackslash   = []byte{'\\', '\\'}
	jsDoubleQuote = []byte{'\\', '"'}
	jsNewline     = []byte{'\\', 'n'}
	jsTab         = []byte{'\\', 't'}
	jsCr          = []byte{'\\', 'r'}
	jsUnicode     = []byte{'\\', 'u', '0', '0', '0', '0'}
	jsHex         = []byte("0123456789ABCDEF")
)

func (v *visitor) writeEscaped(s string) {
	v.b.WriteByte('"')
	v.b.Write(Escape(s))
	v.b.WriteByte('"')
}

// Escape returns the given string as a byte slice, where every non-printable
// rune is made printable.
func Escape(s string) []byte {
	var buf bytes.Buffer

	last := 0
	for i, ch := range s {
		var b []byte
		switch ch {
		case '\t':
			b = jsTab
		case '\r':
			b = jsCr
		case '\n':
			b = jsNewline
		case '"':
			b = jsDoubleQuote
		case '\\':
			b = jsBackslash
		default:
			if ch < ' ' {
				b = jsUnicode
				b[2] = '0'
				b[3] = '0'
				b[4] = jsHex[ch>>4]
				b[5] = jsHex[ch&0xF]
			} else {
				continue
			}
		}
		buf.WriteString(s[last:i])
		buf.Write(b)
		last = i + 1
	}
	buf.WriteString(s[last:])
	return buf.Bytes()
}
