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

// Package htmlenc encodes the abstract syntax tree into HTML5.
package htmlenc

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"zettelstore.de/z/ast"
	"zettelstore.de/z/domain"
	"zettelstore.de/z/encoder"
)

func init() {
	encoder.Register("html", createEncoder)
}

type htmlEncoder struct {
	lang       string // default language
	xhtml      bool   // use XHTML syntax instead of HTML syntax
	material   string // Symbol after link to (external) material.
	newWindow  bool   // open link in new window
	adaptLink  func(*ast.LinkNode) *ast.LinkNode
	adaptImage func(*ast.ImageNode) *ast.ImageNode
	adaptCite  func(*ast.CiteNode) (cn *ast.CiteNode, url string)

	footnotes []*ast.FootnoteNode
}

func createEncoder() encoder.Encoder {
	return &htmlEncoder{}
}

func (he *htmlEncoder) SetOption(option encoder.Option) {
	switch opt := option.(type) {
	case *encoder.StringOption:
		switch opt.Key {
		case "lang":
			he.lang = opt.Value
		case "material":
			he.material = opt.Value
		}
	case *encoder.BoolOption:
		switch opt.Key {
		case "newwindow":
			he.newWindow = opt.Value
		case "xhtml":
			he.xhtml = opt.Value
		}
	case *encoder.AdaptLinkOption:
		he.adaptLink = opt.Adapter
	case *encoder.AdaptImageOption:
		he.adaptImage = opt.Adapter
	case *encoder.AdaptCiteOption:
		he.adaptCite = opt.Adapter
	default:
		fmt.Println("HESO", option, option.Name())
	}
}

// WriteZettel encodes a full zettel as HTML5.
func (he *htmlEncoder) WriteZettel(w io.Writer, zettel *ast.Zettel) (int, error) {
	v := newVisitor(he, w)
	if !he.xhtml {
		v.b.WriteString("<!DOCTYPE html>\n")
	}
	v.b.WriteString("<html lang=\"", he.lang, "\">\n<head>\n<meta charset=\"utf-8\">\n")
	v.acceptMeta(zettel.Meta, zettel.Title)
	v.b.WriteString("\n</head>\n<body>\n")
	v.acceptBlockSlice(zettel.Ast)
	v.writeEndnotes()
	v.b.WriteString("</body>\n</html>")
	length, err := v.b.Flush()
	return length, err
}

// WriteMeta encodes meta data as HTML5.
func (he *htmlEncoder) WriteMeta(w io.Writer, meta *domain.Meta, title ast.InlineSlice) (int, error) {
	v := newVisitor(he, w)
	v.acceptMeta(meta, title)
	length, err := v.b.Flush()
	return length, err
}

// WriteBlocks encodes a block slice.
func (he *htmlEncoder) WriteBlocks(w io.Writer, bs ast.BlockSlice) (int, error) {
	v := newVisitor(he, w)
	v.acceptBlockSlice(bs)
	v.writeEndnotes()
	length, err := v.b.Flush()
	return length, err
}

// WriteInlines writes an inline slice to the writer
func (he *htmlEncoder) WriteInlines(w io.Writer, is ast.InlineSlice) (int, error) {
	v := newVisitor(he, w)
	v.acceptInlineSlice(is)
	length, err := v.b.Flush()
	return length, err
}

// visitor writes the abstract syntax tree to an io.Writer.
type visitor struct {
	enc          *htmlEncoder
	b            encoder.BufWriter
	visibleSpace bool // Show space character in raw text
	inVerse      bool // In verse block
	xhtml        bool // copied from enc.xhtml
}

func newVisitor(he *htmlEncoder, w io.Writer) *visitor {
	return &visitor{enc: he, b: encoder.NewBufWriter(w), xhtml: he.xhtml}
}

func (v *visitor) acceptMeta(meta *domain.Meta, title ast.InlineSlice) {
	textEnc := encoder.Create("text")
	var sb strings.Builder
	textEnc.WriteInlines(&sb, title)
	v.b.WriteString("<title>", sb.String(), "</title>")

	for i, pair := range meta.Pairs() {
		if i == 0 { // "title" is number 0...
			continue
		}
		v.b.WriteString("\n<meta name=\"zettel-", pair.Key, "\" content=\"")
		v.writeEscaped(pair.Value)
		v.b.WriteString("\">")
	}
}

// VisitPara emits HTML code for a paragraph: <p>...</p>
func (v *visitor) VisitPara(pn *ast.ParaNode) {
	v.b.WriteString("<p>")
	v.acceptInlineSlice(pn.Inlines)
	v.b.WriteString("</p>\n")
}

// VisitVerbatim emits HTML code for verbatim lines.
func (v *visitor) VisitVerbatim(vn *ast.VerbatimNode) {
	switch vn.Code {
	case ast.VerbatimProg:
		oldVisible := v.visibleSpace
		if vn.Attrs != nil {
			v.visibleSpace = vn.Attrs.HasDefault()
		}
		v.b.WriteString("<pre><code")
		v.visitAttributes(vn.Attrs)
		v.b.WriteString(">")
		for _, line := range vn.Lines {
			v.writeEscaped(line)
			v.b.WriteByte('\n')
		}
		v.b.WriteString("</code></pre>\n")
		v.visibleSpace = oldVisible

	case ast.VerbatimComment:
		v.b.WriteString("<!-- ")
		for _, line := range vn.Lines {
			v.writeEscaped(line)
			v.b.WriteByte('\n')
		}
		v.b.WriteString("-->\n")

	case ast.VerbatimHTML:
		for _, line := range vn.Lines {
			v.b.WriteString(line)
			v.b.WriteByte('\n')
		}
	default:
		panic(fmt.Sprintf("Unknown verbatim code %v", vn.Code))
	}
}

var regionCode = map[ast.RegionCode][]byte{
	ast.RegionSpan:  []byte("div"),
	ast.RegionQuote: []byte("blockquote"),
	ast.RegionVerse: []byte("div"),
}

// VisitRegion writes HTML code for block regions.
func (v *visitor) VisitRegion(rn *ast.RegionNode) {
	oldVerse := v.inVerse
	if rn.Code == ast.RegionVerse {
		v.inVerse = true
	}
	code, ok := regionCode[rn.Code]
	if !ok {
		panic(fmt.Sprintf("Unknown region code %v", rn.Code))
	}
	v.b.WriteString("<")
	v.b.Write(code)
	v.visitAttributes(rn.Attrs)
	v.b.WriteString(">\n")
	v.acceptBlockSlice(rn.Blocks)
	if len(rn.Inlines) > 0 {
		v.b.WriteString("<cite>")
		v.acceptInlineSlice(rn.Inlines)
		v.b.WriteString("</cite>\n")
	}
	v.b.WriteString("</")
	v.b.Write(code)
	v.b.WriteString(">\n")
	v.inVerse = oldVerse
}

// VisitHeading writes the HTML code for a heading.
func (v *visitor) VisitHeading(hn *ast.HeadingNode) {
	lvl := hn.Level
	if lvl > 6 {
		lvl = 6 // HTML has H1..H6
	}
	v.b.WriteString("<h", strconv.Itoa(lvl))
	v.visitAttributes(hn.Attrs)
	v.b.WriteString(">")
	v.acceptInlineSlice(hn.Inlines)
	v.b.WriteString("</h", strconv.Itoa(lvl), ">\n")
}

// VisitHRule writes HTML code for a horizontal rule: <hr>.
func (v *visitor) VisitHRule(hn *ast.HRuleNode) {
	v.b.WriteString("<hr")
	v.visitAttributes(hn.Attrs)
	if v.xhtml {
		v.b.WriteString(" />\n")
	} else {
		v.b.WriteString(">\n")
	}
}

var listCode = map[ast.ListCode][]byte{
	ast.ListOrdered:   []byte("ol"),
	ast.ListUnordered: []byte("ul"),
}

// VisitList writes HTML code for lists and blockquotes.
func (v *visitor) VisitList(ln *ast.ListNode) {
	if ln.Code == ast.ListQuote {
		// ListQuote -> HTML <blockquote> doesn't use <li>...</li>
		v.b.WriteString("<blockquote>\n")
		for _, item := range ln.Items {
			v.acceptItemSlice(item)
		}
		v.b.WriteString("</blockquote>\n")
		return
	}

	compact := isCompactList(ln.Items)
	v.b.WriteString("<")
	v.b.Write(listCode[ln.Code])
	v.visitAttributes(ln.Attrs)
	v.b.WriteString(">\n")
	for _, item := range ln.Items {
		v.b.WriteString("<li>")
		v.writeItemSliceOrPara(item, compact)
		v.b.WriteString("</li>\n")
	}
	v.b.WriteString("</")
	v.b.Write(listCode[ln.Code])
	v.b.WriteString(">\n")
}

func isCompactList(insl []ast.ItemSlice) bool {
	for _, ins := range insl {
		if !isCompactSlice(ins) {
			return false
		}
	}
	return true
}

func isCompactSlice(ins ast.ItemSlice) bool {
	if len(ins) < 1 {
		return true
	}
	if len(ins) == 1 {
		switch ins[0].(type) {
		case *ast.ParaNode, *ast.VerbatimNode, *ast.HRuleNode:
			return true
		case *ast.ListNode:
			return false
		}
	}
	return false
}

// writeItemSliceOrPara emits the content of a paragraph if the paragraph is
// the only element of the block slice and if compact mode is true. Otherwise,
// the item slice is emitted normally.
func (v *visitor) writeItemSliceOrPara(ins ast.ItemSlice, compact bool) {
	if compact && len(ins) == 1 {
		if para, ok := ins[0].(*ast.ParaNode); ok {
			v.acceptInlineSlice(para.Inlines)
			return
		}
	}
	v.acceptItemSlice(ins)
}

func (v *visitor) writeDescriptionsSlice(ds ast.DescriptionSlice) {
	if len(ds) == 1 {
		if para, ok := ds[0].(*ast.ParaNode); ok {
			v.acceptInlineSlice(para.Inlines)
			return
		}
	}
	for _, dn := range ds {
		dn.Accept(v)
	}
}

// VisitDefinition emits a HTML definition list.
func (v *visitor) VisitDefinition(dn *ast.DefinitionNode) {
	v.b.WriteString("<dl>\n")
	for _, def := range dn.Definitions {
		v.b.WriteString("<dt>")
		v.acceptInlineSlice(def.Term)
		v.b.WriteString("</dt>\n")

		for _, b := range def.Descriptions {
			v.b.WriteString("<dd>")
			v.writeDescriptionsSlice(b)
			v.b.WriteString("</dd>\n")
		}
	}
	v.b.WriteString("</dl>\n")
}

// VisitTable emits a HTML table.
func (v *visitor) VisitTable(tn *ast.TableNode) {
	v.b.WriteString("<table>\n")
	if len(tn.Header) > 0 {
		v.b.WriteString("<thead>\n")
		v.writeRow(tn.Header, "<th", "</th>")
		v.b.WriteString("</thead>\n")
	}
	if len(tn.Rows) > 0 {
		v.b.WriteString("<tbody>\n")
		for _, row := range tn.Rows {
			v.writeRow(row, "<td", "</td>")
		}
		v.b.WriteString("</tbody>\n")
	}
	v.b.WriteString("</table>\n")
}

var alignStyle = map[ast.Alignment]string{
	ast.AlignDefault: ">",
	ast.AlignLeft:    " style=\"text-align:left\">",
	ast.AlignCenter:  " style=\"text-align:center\">",
	ast.AlignRight:   " style=\"text-align:right\">",
}

func (v *visitor) writeRow(row ast.TableRow, cellStart, cellEnd string) {
	v.b.WriteString("<tr>")
	for _, cell := range row {
		v.b.WriteString(cellStart)
		if len(cell.Inlines) == 0 {
			v.b.WriteByte('>')
		} else {
			v.b.WriteString(alignStyle[cell.Align])
			v.acceptInlineSlice(cell.Inlines)
		}
		v.b.WriteString(cellEnd)
	}
	v.b.WriteString("</tr>\n")
}

// VisitBLOB writes the binary object as a value.
func (v *visitor) VisitBLOB(bn *ast.BLOBNode) {
	switch bn.Syntax {
	case "gif", "jpeg", "png":
		v.b.WriteString("<img src=\"data:image/", bn.Syntax, ";base64,")
		v.b.WriteBase64(bn.Blob)
		v.b.WriteString("\" title=\"")
		v.writeEscaped(bn.Title)
		v.b.WriteString("\">\n")
	default:
		v.b.WriteString("<p class=\"error\">Unable to display BLOB with syntax '", bn.Syntax, "'.</p>\n")
	}
}

// VisitText writes text content.
func (v *visitor) VisitText(tn *ast.TextNode) {
	v.writeEscaped(tn.Text)
}

// VisitTag writes tag content.
func (v *visitor) VisitTag(tn *ast.TagNode) {
	// TODO: erst mal als span. Link wäre gut, muss man vermutlich via Callback lösen.
	v.b.WriteString("<span class=\"zettel-tag\">")
	v.writeEscaped(tn.Tag)
	v.b.WriteString("</span>")
}

// VisitSpace emits a white space.
func (v *visitor) VisitSpace(sn *ast.SpaceNode) {
	if v.inVerse || v.xhtml {
		v.b.WriteString(sn.Lexeme)
	} else {
		v.b.WriteByte(' ')
	}
}

// VisitBreak writes HTML code for line breaks.
func (v *visitor) VisitBreak(bn *ast.BreakNode) {
	if bn.Hard {
		if v.xhtml {
			v.b.WriteString("<br />\n")
		} else {
			v.b.WriteString("<br>\n")
		}
	} else {
		v.b.WriteByte('\n')
	}
}

// VisitLink writes HTML code for links.
func (v *visitor) VisitLink(ln *ast.LinkNode) {
	if adapt := v.enc.adaptLink; adapt != nil {
		ln = adapt(ln)
	}
	if ln == nil {
		return
	}
	switch ln.Ref.State {
	case ast.RefStateZettelFound:
		v.writeAHref(ln.Ref, ln.Attrs, ln.Inlines)
	case ast.RefStateZettelBroken:
		attrs := ln.Attrs.Clone()
		attrs = attrs.Set("class", "zs-broken")
		attrs = attrs.Set("title", "Zettel not found") // l10n
		v.writeAHref(ln.Ref, attrs, ln.Inlines)
	case ast.RefStateMaterial:
		attrs := ln.Attrs.Clone()
		attrs = attrs.Set("class", "zs-external")
		if v.enc.newWindow {
			attrs = attrs.Set("target", "_blank")
		}
		v.writeAHref(ln.Ref, attrs, ln.Inlines)
		v.b.WriteString(v.enc.material)
	default:
		v.b.WriteString("<a href=\"")
		v.writeEscaped(ln.Ref.Value)
		v.b.WriteByte('"')
		v.visitAttributes(ln.Attrs)
		v.b.WriteByte('>')
		v.acceptInlineSlice(ln.Inlines)
		v.b.WriteString("</a>")
	}
}

func (v *visitor) writeAHref(ref *ast.Reference, attrs *ast.Attributes, ins ast.InlineSlice) {
	v.b.WriteString("<a href=\"")
	v.writeReference(ref)
	v.b.WriteByte('"')
	v.visitAttributes(attrs)
	v.b.WriteByte('>')
	v.acceptInlineSlice(ins)
	v.b.WriteString("</a>")
}

// VisitImage writes HTML code for images.
func (v *visitor) VisitImage(in *ast.ImageNode) {
	if adapt := v.enc.adaptImage; adapt != nil {
		in = adapt(in)
	}
	if in == nil {
		return
	}

	if in.Ref == nil {
		v.b.WriteString("<img src=\"data:image/")
		switch in.Syntax {
		case "svg":
			v.b.WriteString("svg+xml;utf8,")
			v.writeEscaped(string(in.Blob))
		default:
			v.b.WriteString(in.Syntax, ";base64,")
			v.b.WriteBase64(in.Blob)
		}
	} else {
		v.b.WriteString("<img src=\"")
		v.writeReference(in.Ref)
	}
	v.b.WriteString("\" alt=\"")
	v.acceptInlineSlice(in.Inlines)
	v.b.WriteByte('"')
	v.visitAttributes(in.Attrs)
	if v.xhtml {
		v.b.WriteString(" />")
	} else {
		v.b.WriteByte('>')
	}
}

// VisitCite writes code for citations.
func (v *visitor) VisitCite(cn *ast.CiteNode) {
	var url string
	if adapt := v.enc.adaptCite; adapt != nil {
		cn, url = adapt(cn)
	}
	if cn != nil {
		if url == "" {
			v.b.WriteString(cn.Key)
			// TODO: Attrs
		} else {
			v.b.WriteString("<a href=\"", url)
			v.b.WriteByte('"')
			v.visitAttributes(cn.Attrs)
			v.b.WriteByte('>')
			v.b.WriteString(cn.Key, "</a>")
		}
		if len(cn.Inlines) > 0 {
			v.b.WriteString(", ")
			v.acceptInlineSlice(cn.Inlines)
		}
	}
}

// VisitFootnote write HTML code for a footnote.
func (v *visitor) VisitFootnote(fn *ast.FootnoteNode) {
	v.enc.footnotes = append(v.enc.footnotes, fn)
	n := fmt.Sprintf("%d", len(v.enc.footnotes))
	v.b.WriteString("<sup id=\"fnref:", n, "\"><a href=\"#fn:", n, "\" class=\"zs-footnote-ref\" role=\"doc-noteref\">", n, "</a></sup>")
	// TODO: what to do with Attrs?
}

// VisitMark writes HTML code to mark a position.
func (v *visitor) VisitMark(mn *ast.MarkNode) {
	if len(mn.Text) > 0 {
		v.b.WriteString("<a id=\"", mn.Text, "\"></a>")
	}
}

var formatCode = map[ast.FormatCode][]byte{
	ast.FormatItalic:    []byte("i"),
	ast.FormatEmph:      []byte("em"),
	ast.FormatBold:      []byte("b"),
	ast.FormatStrong:    []byte("strong"),
	ast.FormatUnder:     []byte("u"), // TODO: ändern in <span class="XXX">
	ast.FormatStrike:    []byte("s"),
	ast.FormatSuper:     []byte("sup"),
	ast.FormatSub:       []byte("sub"),
	ast.FormatQuotation: []byte("q"),
	ast.FormatSmall:     []byte("small"),
	ast.FormatSpan:      []byte("span"),
	ast.FormatMonospace: []byte("span"),
}

// VisitFormat write HTML code for formatting text.
func (v *visitor) VisitFormat(fn *ast.FormatNode) {
	if fn.Code == ast.FormatQuote {
		v.visitQuotes(fn)
		return
	}
	v.b.WriteByte('<')
	v.b.Write(formatCode[fn.Code])
	switch fn.Code {
	case ast.FormatMonospace:
		v.b.WriteString(" style=\"font-family:monospace\"")
	}
	v.visitAttributes(fn.Attrs)
	v.b.WriteByte('>')
	v.acceptInlineSlice(fn.Inlines)
	v.b.WriteString("</")
	v.b.Write(formatCode[fn.Code])
	v.b.WriteByte('>')
}

var langQuotes = map[string][2]string{
	"en": {"&ldquo;", "&rdquo;"},
	"de": {"&bdquo;", "&ldquo;"},
	"fr": {"&laquo;&nbsp;", "&nbsp;&raquo;"},
}

func getQuotes(lang string) (string, string) {
	langFields := strings.FieldsFunc(lang, func(r rune) bool { return r == '-' || r == '_' })
	for len(langFields) > 0 {
		langSup := strings.Join(langFields, "-")
		quotes, ok := langQuotes[langSup]
		if ok {
			return quotes[0], quotes[1]
		}
		langFields = langFields[0 : len(langFields)-1]
	}
	return "\"", "\""
}

func (v *visitor) visitQuotes(fn *ast.FormatNode) {
	lang, _ := fn.Attrs.Get("lang")
	if len(lang) == 0 {
		lang = v.enc.lang
	}
	withSpan := len(lang) > 0

	openingQ, closingQ := getQuotes(lang)
	if withSpan {
		v.b.WriteString("<span")
		v.visitAttributes(fn.Attrs)
		v.b.WriteByte('>')
	}
	v.b.WriteString(openingQ)
	v.acceptInlineSlice(fn.Inlines)
	v.b.WriteString(closingQ)
	if withSpan {
		v.b.WriteString("</span>")
	}
}

// VisitEdit write HTML code for edit markup text.
func (v *visitor) VisitEdit(en *ast.EditNode) {
	if len(en.Deletes) > 0 {
		v.b.WriteString("<del>")
		v.acceptInlineSlice(en.Deletes)
		v.b.WriteString("</del>")
	}
	if len(en.Inserts) > 0 {
		v.b.WriteString("<ins>")
		v.acceptInlineSlice(en.Inserts)
		v.b.WriteString("</ins>")
	}
}

// VisitLiteral write HTML code for literal inline text.
func (v *visitor) VisitLiteral(ln *ast.LiteralNode) {
	switch ln.Code {
	case ast.LiteralProg:
		v.writeLiteral("<code", "</code>", ln.Attrs, ln.Text)
	case ast.LiteralKeyb:
		v.writeLiteral("<kbd", "</kbd>", ln.Attrs, ln.Text)
	case ast.LiteralOutput:
		v.writeLiteral("<samp", "</samp>", ln.Attrs, ln.Text)
	case ast.LiteralComment:
		v.b.WriteString("<!-- ")
		v.writeEscaped(ln.Text)
		v.b.WriteString(" -->")
	case ast.LiteralHTML:
		v.b.WriteString(ln.Text)
	default:
		panic(fmt.Sprintf("Unknown literal code %v", ln.Code))
	}
}

func (v *visitor) writeLiteral(codeS, codeE string, attrs *ast.Attributes, text string) {
	oldVisible := v.visibleSpace
	if attrs != nil {
		v.visibleSpace = attrs.HasDefault()
	}
	v.b.WriteString(codeS)
	v.visitAttributes(attrs)
	v.b.WriteByte('>')
	v.writeEscaped(text)
	v.b.WriteString(codeE)
	v.visibleSpace = oldVisible
}

func (v *visitor) acceptBlockSlice(bns ast.BlockSlice) {
	for _, bn := range bns {
		bn.Accept(v)
	}
}
func (v *visitor) acceptItemSlice(ins ast.ItemSlice) {
	for _, in := range ins {
		in.Accept(v)
	}
}
func (v *visitor) acceptInlineSlice(ins ast.InlineSlice) {
	for _, in := range ins {
		in.Accept(v)
	}
}

func (v *visitor) writeEndnotes() {
	if len(v.enc.footnotes) > 0 {
		v.b.WriteString("<ol class=\"zs-endnotes\">\n")
		for i := 0; i < len(v.enc.footnotes); i++ {
			// Do not use a range loop above, because a footnote may contain
			// a footnote. Therefore v.enc.footnote may grow during the loop.
			fn := v.enc.footnotes[i]
			n := fmt.Sprintf("%d", i+1)
			v.b.WriteString("<li id=\"fn:", n, "\" role=\"doc-endnote\">")
			v.acceptInlineSlice(fn.Inlines)
			v.b.WriteString(" <a href=\"#fnref:", n, "\" class=\"zs-footnote-backref\" role=\"doc-backlink\">&#x21a9;&#xfe0e;</a></li>\n")
		}
		v.b.WriteString("</ol>\n")
	}
}

// visitAttributes write HTML attributes
func (v *visitor) visitAttributes(a *ast.Attributes) {
	if a == nil || len(a.Attrs) == 0 {
		return
	}
	keys := make([]string, 0, len(a.Attrs))
	for k := range a.Attrs {
		if k != "-" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	for _, k := range keys {
		if k == "" || k == "-" {
			continue
		}
		v.b.WriteByte(' ')
		v.b.WriteString(k)
		vl := a.Attrs[k]
		if len(vl) > 0 {
			v.b.WriteString("=\"")
			v.writeEscaped(vl)
			v.b.WriteByte('"')
		}
	}
}

func (v *visitor) writeEscaped(s string) {
	last := 0
	var html string
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\000':
			html = "\uFFFD"
		case ' ':
			if v.visibleSpace {
				html = "\U00002423"
			} else {
				continue
			}
		case '"':
			if v.xhtml {
				html = "&quot;"
			} else {
				html = "&#34;"
			}
		case '&':
			html = "&amp;"
		case '<':
			html = "&lt;"
		case '>':
			html = "&gt;"
		default:
			continue
		}
		v.b.WriteString(s[last:i], html)
		last = i + 1
	}
	v.b.WriteString(s[last:])
}

func (v *visitor) writeReference(ref *ast.Reference) {
	if ref.URL == nil {
		v.writeEscaped(ref.Value)
		return
	}
	v.b.WriteString(ref.URL.String())
}
