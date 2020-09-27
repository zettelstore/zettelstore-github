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

// Package dirplace provides a directory-based zettel place.
package dirplace

import (
	"io/ioutil"
	"os"
	"strings"

	"zettelstore.de/z/domain"
	"zettelstore.de/z/input"
	"zettelstore.de/z/place/dirplace/directory"
)

func fileService(num uint32, cmds <-chan fileCmd) {
	for cmd := range cmds {
		cmd.run()
	}
}

type fileCmd interface {
	run()
}

// COMMAND: getMeta ----------------------------------------
//
// Retrieves the meta data from a zettel.

type fileGetMeta struct {
	entry *directory.Entry
	rc    chan<- resGetMeta
}
type resGetMeta struct {
	meta *domain.Meta
	err  error
}

func (cmd *fileGetMeta) run() {
	var meta *domain.Meta
	var err error
	switch cmd.entry.MetaSpec {
	case directory.MetaSpecFile:
		meta, err = parseMetaFile(cmd.entry.Zid, cmd.entry.MetaPath)
	case directory.MetaSpecHeader:
		meta, _, err = parseMetaContentFile(cmd.entry.Zid, cmd.entry.ContentPath)
	default:
		meta = calculateMeta(cmd.entry)
	}
	if err == nil {
		cleanupMeta(meta, cmd.entry)
	}
	cmd.rc <- resGetMeta{meta, err}
}

// COMMAND: getMetaContent ----------------------------------------
//
// Retrieves the meta data and the content of a zettel.

type fileGetMetaContent struct {
	entry *directory.Entry
	rc    chan<- resGetMetaContent
}
type resGetMetaContent struct {
	meta    *domain.Meta
	content string
	err     error
}

func (cmd *fileGetMetaContent) run() {
	var meta *domain.Meta
	var content string
	var err error

	switch cmd.entry.MetaSpec {
	case directory.MetaSpecFile:
		meta, err = parseMetaFile(cmd.entry.Zid, cmd.entry.MetaPath)
		content, err = readFileContent(cmd.entry.ContentPath)
	case directory.MetaSpecHeader:
		meta, content, err = parseMetaContentFile(cmd.entry.Zid, cmd.entry.ContentPath)
	default:
		meta = calculateMeta(cmd.entry)
		content, err = readFileContent(cmd.entry.ContentPath)
	}
	if err == nil {
		cleanupMeta(meta, cmd.entry)
	}
	cmd.rc <- resGetMetaContent{meta, content, err}
}

// COMMAND: setZettel ----------------------------------------
//
// Writes a new or exsting zettel.

type fileSetZettel struct {
	entry  *directory.Entry
	zettel domain.Zettel
	rc     chan<- resSetZettel
}
type resSetZettel = error

func (cmd *fileSetZettel) run() {
	var f *os.File
	var err error

	switch cmd.entry.MetaSpec {
	case directory.MetaSpecFile:
		f, err = openFileWrite(cmd.entry.MetaPath)
		if err == nil {
			_, err = cmd.zettel.Meta.Write(f)
			if err1 := f.Close(); err == nil {
				err = err1
			}

			if err == nil {
				err = writeFileContent(cmd.entry.ContentPath, cmd.zettel.Content.AsString())
			}
		}

	case directory.MetaSpecHeader:
		f, err = openFileWrite(cmd.entry.ContentPath)
		if err == nil {
			_, err = cmd.zettel.Meta.WriteAsHeader(f)
			if err == nil {
				_, err = f.WriteString(cmd.zettel.Content.AsString())
				if err1 := f.Close(); err == nil {
					err = err1
				}
			}
		}

	case directory.MetaSpecNone:
		// TODO: if meta has some additional infos: write meta to new .meta; update entry in dir

		err = writeFileContent(cmd.entry.ContentPath, cmd.zettel.Content.AsString())

	case directory.MetaSpecUnknown:
		panic("TODO: ???")
	}
	cmd.rc <- err
}

// COMMAND: renameZettel ----------------------------------------
//
// Gives an existing zettel a new id.

type fileRenameZettel struct {
	curEntry *directory.Entry
	newEntry *directory.Entry
	rc       chan<- resRenameZettel
}

type resRenameZettel = error

func (cmd *fileRenameZettel) run() {
	var err error

	switch cmd.curEntry.MetaSpec {
	case directory.MetaSpecFile:
		err1 := os.Rename(cmd.curEntry.MetaPath, cmd.newEntry.MetaPath)
		err = os.Rename(cmd.curEntry.ContentPath, cmd.newEntry.ContentPath)
		if err == nil {
			err = err1
		}
	case directory.MetaSpecHeader, directory.MetaSpecNone:
		err = os.Rename(cmd.curEntry.ContentPath, cmd.newEntry.ContentPath)
	case directory.MetaSpecUnknown:
		panic("TODO: ???")
	}
	cmd.rc <- err
}

// COMMAND: deleteZettel ----------------------------------------
//
// Deletes an existing zettel.

type fileDeleteZettel struct {
	entry *directory.Entry
	rc    chan<- resDeleteZettel
}
type resDeleteZettel = error

func (cmd *fileDeleteZettel) run() {
	var err error

	switch cmd.entry.MetaSpec {
	case directory.MetaSpecFile:
		err1 := os.Remove(cmd.entry.MetaPath)
		err = os.Remove(cmd.entry.ContentPath)
		if err == nil {
			err = err1
		}
	case directory.MetaSpecHeader:
		err = os.Remove(cmd.entry.ContentPath)
	case directory.MetaSpecNone:
		err = os.Remove(cmd.entry.ContentPath)
	case directory.MetaSpecUnknown:
		panic("TODO: ???")
	}
	cmd.rc <- err
}

// Utility functions ----------------------------------------

func readFileContent(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func parseMetaFile(zid domain.ZettelID, path string) (*domain.Meta, error) {
	src, err := readFileContent(path)
	if err != nil {
		return nil, err
	}
	inp := input.NewInput(src)
	return domain.NewMetaFromInput(zid, inp), nil
}

func parseMetaContentFile(zid domain.ZettelID, path string) (*domain.Meta, string, error) {
	src, err := readFileContent(path)
	if err != nil {
		return nil, "", err
	}
	inp := input.NewInput(src)
	meta := domain.NewMetaFromInput(zid, inp)
	return meta, src[inp.Pos:], nil
}

func cleanupMeta(meta *domain.Meta, entry *directory.Entry) {
	if title, ok := meta.Get(domain.MetaKeyTitle); !ok || title == "" {
		meta.Set(domain.MetaKeyTitle, entry.Zid.Format())
	}

	switch entry.MetaSpec {
	case directory.MetaSpecFile:
		if syntax, ok := meta.Get(domain.MetaKeySyntax); !ok || syntax == "" {
			meta.Set(domain.MetaKeySyntax, calculateSyntax(entry))
		}
	}

	if entry.Duplicates {
		meta.Set("duplicates", "yes")
	}
}

var alternativeSyntax = map[string]string{
	"htm":  "html",
	"tmpl": "go-template-html",
}

func calculateSyntax(entry *directory.Entry) string {
	ext := strings.ToLower(entry.ContentExt)
	if syntax, ok := alternativeSyntax[ext]; ok {
		return syntax
	}
	return ext
}

func calculateMeta(entry *directory.Entry) *domain.Meta {
	meta := domain.NewMeta(entry.Zid)
	meta.Set(domain.MetaKeyTitle, entry.Zid.Format())
	meta.Set(domain.MetaKeySyntax, calculateSyntax(entry))
	return meta
}

func openFileWrite(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
}

func writeFileContent(path string, content string) error {
	f, err := openFileWrite(path)
	if err == nil {
		_, err = f.WriteString(content)
		if err1 := f.Close(); err == nil {
			err = err1
		}
	}
	return err
}
