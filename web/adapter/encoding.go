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

// Package adapter provides handlers for web requests.
package adapter

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"strings"

	"zettelstore.de/z/ast"
	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
	"zettelstore.de/z/encoder"
	"zettelstore.de/z/parser"
	"zettelstore.de/z/place"
	"zettelstore.de/z/usecase"
)

var errNoSuchFormat = errors.New("no such format")

func formatBlocks(bs ast.BlockSlice, format string, options ...encoder.Option) (string, error) {
	enc := encoder.Create(format, options...)
	if enc == nil {
		return "", errNoSuchFormat
	}

	var content strings.Builder
	_, err := enc.WriteBlocks(&content, bs)
	if err != nil {
		return "", err
	}
	return content.String(), nil
}

func formatMeta(meta *domain.Meta, format string, options ...encoder.Option) (string, error) {
	enc := encoder.Create(format, options...)
	if enc == nil {
		return "", errNoSuchFormat
	}

	var content strings.Builder
	_, err := enc.WriteMeta(&content, meta)
	if err != nil {
		return "", err
	}
	return content.String(), nil
}

func formatInlines(is ast.InlineSlice, format string, options ...encoder.Option) (string, error) {
	enc := encoder.Create(format, options...)
	if enc == nil {
		return "", errNoSuchFormat
	}

	var content strings.Builder
	_, err := enc.WriteInlines(&content, is)
	if err != nil {
		return "", err
	}
	return content.String(), nil
}

func writeZettel(w io.Writer, zettel *ast.Zettel, format string, options ...encoder.Option) error {
	enc := encoder.Create(format, options...)
	if enc == nil {
		return errNoSuchFormat
	}

	_, err := enc.WriteZettel(w, zettel)
	return err
}

func writeContent(w io.Writer, zettel *ast.Zettel, format string, options ...encoder.Option) error {
	enc := encoder.Create(format, options...)
	if enc == nil {
		return errNoSuchFormat
	}

	_, err := enc.WriteContent(w, zettel)
	return err
}

func writeMeta(w io.Writer, meta *domain.Meta, format string, options ...encoder.Option) error {
	enc := encoder.Create(format, options...)
	if enc == nil {
		return errNoSuchFormat
	}

	_, err := enc.WriteMeta(w, meta)
	return err
}

func makeLinkAdapter(ctx context.Context, key byte, getMeta usecase.GetMeta, part, format string) func(*ast.LinkNode) ast.InlineNode {
	return func(origLink *ast.LinkNode) ast.InlineNode {
		origRef := origLink.Ref
		if origRef == nil || origRef.State != ast.RefStateZettel {
			return origLink
		}
		zid, err := domain.ParseZettelID(origRef.Value)
		if err != nil {
			panic(err)
		}
		_, err = getMeta.Run(ctx, zid)
		newLink := *origLink
		if err == nil {
			url := urlForZettel(key, zid)
			if part != "" {
				if format != "" {
					url = fmt.Sprintf("%v?_part=%v&_format=%v", url, part, format)
				} else {
					url = fmt.Sprintf("%v?_part=%v", url, part)
				}
			} else if format != "" {
				url = fmt.Sprintf("%v?_format=%v", url, format)
			}
			newRef := ast.ParseReference(url)
			newRef.State = ast.RefStateZettelFound
			newLink.Ref = newRef
			return &newLink
		}
		if place.IsErrNotAllowed(err) {
			return &ast.FormatNode{
				Code:    ast.FormatSpan,
				Attrs:   origLink.Attrs,
				Inlines: origLink.Inlines,
			}
		}
		newRef := ast.ParseReference(origRef.Value)
		newRef.State = ast.RefStateZettelBroken
		newLink.Ref = newRef
		return &newLink
	}
}

func makeImageAdapter() func(*ast.ImageNode) ast.InlineNode {
	return func(origImage *ast.ImageNode) ast.InlineNode {
		if origImage.Ref == nil || origImage.Ref.State != ast.RefStateZettel {
			return origImage
		}
		newImage := *origImage
		zid, err := domain.ParseZettelID(newImage.Ref.Value)
		if err != nil {
			panic(err)
		}
		newImage.Ref = ast.ParseReference(urlForZettel('z', zid) + "?_part=content&_format=raw")
		newImage.Ref.State = ast.RefStateZettelFound
		return &newImage
	}
}

type metaInfo struct {
	Meta  metaWrapper
	Title template.HTML
}

// buildHTMLMetaList builds a zettel list based on a meta list for HTML rendering.
func buildHTMLMetaList(metaList []*domain.Meta) ([]metaInfo, error) {
	defaultLang := config.GetDefaultLang()
	langOption := encoder.StringOption{Key: "lang", Value: ""}
	metas := make([]metaInfo, 0, len(metaList))
	for _, meta := range metaList {
		if lang, ok := meta.Get(domain.MetaKeyLang); ok {
			langOption.Value = lang
		} else {
			langOption.Value = defaultLang
		}
		title, _ := meta.Get(domain.MetaKeyTitle)
		htmlTitle, err := formatInlines(parser.ParseTitle(title), "html", &langOption)
		if err != nil {
			return nil, err
		}
		metas = append(metas, metaInfo{wrapMeta(meta), template.HTML(htmlTitle)})
	}
	return metas, nil
}
