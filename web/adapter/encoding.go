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
	"html/template"
	"io"
	"strings"

	"zettelstore.de/z/ast"
	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
	"zettelstore.de/z/encoder"
	"zettelstore.de/z/parser"
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

func writeBlocks(w io.Writer, bs ast.BlockSlice, format string, options ...encoder.Option) error {
	enc := encoder.Create(format, options...)
	if enc == nil {
		return errNoSuchFormat
	}

	_, err := enc.WriteBlocks(w, bs)
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

func makeLinkAdapter(ctx context.Context, key byte, getMeta usecase.GetMeta) func(*ast.LinkNode) *ast.LinkNode {
	return func(origLink *ast.LinkNode) *ast.LinkNode {
		if origRef := origLink.Ref; origRef.IsZettel() {
			id, err := domain.ParseZettelID(origRef.Value)
			if err != nil {
				panic(err)
			}
			_, err = getMeta.Run(ctx, id)
			newLink := *origLink
			if err == nil {
				newRef := ast.ParseReference(urlForZettel(key, id))
				newRef.State = ast.RefStateZettelFound
				newLink.Ref = newRef
			} else {
				newRef := ast.ParseReference(origRef.Value)
				newRef.State = ast.RefStateZettelBroken
				newLink.Ref = newRef
			}
			return &newLink
		}
		return origLink
	}
}

func makeImageAdapter() func(*ast.ImageNode) *ast.ImageNode {
	return func(origImage *ast.ImageNode) *ast.ImageNode {
		if origImage.Ref == nil {
			return origImage
		}
		newImage := *origImage
		if newImage.Ref.IsZettel() {
			id, err := domain.ParseZettelID(newImage.Ref.Value)
			if err != nil {
				panic(err)
			}
			newImage.Ref = ast.ParseReference(urlForZettel('c', id))
		}
		return &newImage
	}
}

type metaInfo struct {
	Meta  *domain.Meta
	Title template.HTML
}

// buildHTMLMetaList builds a zettel list based on a meta list for HTML rendering.
func buildHTMLMetaList(metaList []*domain.Meta) ([]metaInfo, error) {
	defaultLang := config.Config.GetDefaultLang()
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
		metas = append(metas, metaInfo{meta, template.HTML(htmlTitle)})
	}
	return metas, nil
}
