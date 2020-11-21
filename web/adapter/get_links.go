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
	"io"
	"net/http"
	"strconv"

	"zettelstore.de/z/ast"
	"zettelstore.de/z/collect"
	"zettelstore.de/z/domain"
	"zettelstore.de/z/usecase"
)

// MakeGetLinksHandler creates a new API handler to return links to other material.
func MakeGetLinksHandler(parseZettel usecase.ParseZettel) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		zid, err := domain.ParseZettelID(r.URL.Path[1:])
		if err != nil {
			http.NotFound(w, r)
			return
		}
		ctx := r.Context()
		q := r.URL.Query()
		zn, err := parseZettel.Run(ctx, zid, q.Get("syntax"))
		if err != nil {
			checkUsecaseError(w, err)
			return
		}
		summary := collect.References(zn)

		kind := getKindFromValue(q.Get("kind"))
		matter := getMatterFromValue(q.Get("matter"))
		if !validKindMatter(kind, matter) {
			http.Error(w, "Invalid kind/matter", http.StatusBadRequest)
			return
		}

		err = writeJSONHeader(w, zid, "json")
		if err == nil && kind&kindLink != 0 {
			doComma := false
			_, err = w.Write([]byte(",\"link\":{"))
			if err == nil && matter&matterIncoming != 0 {
				_, err = w.Write([]byte("\"incoming\":["))
				// backlink
				err = closeList(w, err)
				doComma = true
			}
			zetRefs, locRefs, extRefs := collect.DivideReferences(summary.Links, false)
			if err == nil && matter&matterOutgoing != 0 {
				err = emitComma(w, doComma)
				if err == nil {
					_, err = w.Write([]byte("\"outgoing\":["))
				}
				err = emitZettelRefs(w, err, zetRefs)
				err = closeList(w, err)
				doComma = true
			}
			if err == nil && matter&matterLocal != 0 {
				err = emitComma(w, doComma)
				if err == nil {
					_, err = w.Write([]byte("\"local\":["))
				}
				err = emitStringRefs(w, err, locRefs)
				err = closeList(w, err)
				doComma = true
			}
			if err == nil && matter&matterExternal != 0 {
				err = emitComma(w, doComma)
				if err == nil {
					_, err = w.Write([]byte("\"external\":["))
				}
				err = emitStringRefs(w, err, extRefs)
				err = closeList(w, err)
			}
			err = closeObject(w, err)
		}
		if err == nil && kind&kindImage != 0 {
			doComma := false
			_, err = w.Write([]byte(",\"image\":{"))
			zetRefs, locRefs, extRefs := collect.DivideReferences(summary.Images, false)
			if err == nil && matter&matterOutgoing != 0 {
				_, err = w.Write([]byte("\"outgoing\":["))
				err = emitZettelRefs(w, err, zetRefs)
				err = closeList(w, err)
				doComma = true
			}
			if err == nil && matter&matterLocal != 0 {
				err = emitComma(w, doComma)
				if err == nil {
					_, err = w.Write([]byte("\"local\":["))
				}
				err = emitStringRefs(w, err, locRefs)
				err = closeList(w, err)
				doComma = true
			}
			if err == nil && matter&matterExternal != 0 {
				err = emitComma(w, doComma)
				if err == nil {
					_, err = w.Write([]byte("\"external\":["))
				}
				err = emitStringRefs(w, err, extRefs)
				err = closeList(w, err)
			}
			err = closeObject(w, err)
		}
		if err == nil && kind&kindCite != 0 {
			_, err = w.Write([]byte(",\"cite\":["))
			err = emitStringCites(w, err, summary.Cites)
			err = closeList(w, err)
		}

		if err == nil {
			err = writeJSONFooter(w)
		}
	}
}

func emitComma(w io.Writer, doComma bool) error {
	if doComma {
		_, err := w.Write([]byte{','})
		return err
	}
	return nil
}
func closeList(w io.Writer, err error) error {
	if err == nil {
		_, err = w.Write([]byte{']'})
	}
	return err
}
func closeObject(w io.Writer, err error) error {
	if err == nil {
		_, err = w.Write([]byte{'}'})
	}
	return err
}

func emitZettelRefs(w io.Writer, err error, refs []*ast.Reference) error {
	if err != nil {
		return err
	}
	for i, ref := range refs {
		if i > 0 && err == nil {
			_, err = w.Write([]byte{','})
		}
		if err == nil {
			zid, err1 := domain.ParseZettelID(ref.Value)
			if err1 == nil {
				err = writeJSONID(w, zid, "json")
			} else {
				err = err1
			}
		}
		err = closeObject(w, err)
	}
	return err
}
func emitStringRefs(w io.Writer, err error, refs []*ast.Reference) error {
	if err != nil {
		return err
	}
	for i, ref := range refs {
		if i > 0 && err == nil {
			_, err = w.Write([]byte{','})
		}
		if err == nil {
			_, err = w.Write([]byte{'"'})
		}
		if err == nil {
			_, err = w.Write([]byte(ref.String()))
		}
		if err == nil {
			_, err = w.Write([]byte{'"'})
		}
	}
	return err
}

func emitStringCites(w io.Writer, err error, cites []*ast.CiteNode) error {
	if err != nil {
		return err
	}
	mapKey := make(map[string]bool)
	for i, cn := range cites {
		if i > 0 && err == nil {
			_, err = w.Write([]byte{','})
		}
		if err == nil {
			_, err = w.Write([]byte{'"'})
		}
		if err == nil {
			if _, ok := mapKey[cn.Key]; !ok {
				_, err = w.Write([]byte(cn.Key))
				mapKey[cn.Key] = true
			}
		}
		if err == nil {
			_, err = w.Write([]byte{'"'})
		}
	}
	return err
}

type kindType int

const (
	_ kindType = 1 << iota
	kindLink
	kindImage
	kindCite
)

var mapKind = map[string]kindType{
	"":      kindLink,
	"link":  kindLink,
	"image": kindImage,
	"cite":  kindCite,
	"both":  kindLink | kindImage,
	"all":   kindLink | kindImage | kindCite,
}

func getKindFromValue(value string) kindType {
	if k, ok := mapKind[value]; ok {
		return k
	}
	if n, err := strconv.Atoi(value); err == nil && n > 0 {
		return kindType(n)
	}
	return 0
}

type matterType int

const (
	_ matterType = 1 << iota
	matterIncoming
	matterOutgoing
	matterLocal
	matterExternal
)

var mapMatter = map[string]matterType{
	"":         matterOutgoing,
	"incoming": matterIncoming,
	"outgoing": matterOutgoing,
	"local":    matterLocal,
	"external": matterExternal,
	"zettel":   matterIncoming | matterOutgoing,
	"material": matterLocal | matterExternal,
	"all":      matterIncoming | matterOutgoing | matterLocal | matterExternal,
}

func getMatterFromValue(value string) matterType {
	if m, ok := mapMatter[value]; ok {
		return m
	}
	if n, err := strconv.Atoi(value); err == nil && n > 0 {
		return matterType(n)
	}
	return 0
}

func validKindMatter(kind kindType, matter matterType) bool {
	if kind == 0 {
		return false
	}
	if kind&kindLink != 0 {
		if matter == 0 {
			return false
		}
		return true
	}
	if kind&kindImage != 0 {
		if matter == 0 || matter == matterIncoming {
			return false
		}
		return true
	}
	if kind&kindCite != 0 {
		return matter == matterOutgoing
	}
	return false
}
