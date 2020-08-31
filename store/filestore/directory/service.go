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

// Package directory manages the directory part of a file store.
package directory

import (
	"log"
	"time"

	"zettelstore.de/z/domain"
	"zettelstore.de/z/store"
)

// ping sends every tick a signal to reload the directory list
func ping(done chan<- struct{}, tick *time.Ticker) {
	defer close(done)
	for {
		select {
		case _, ok := <-tick.C:
			if !ok {
				return
			}
			done <- struct{}{}
		}
	}
}

func newEntry(ev *fileEvent) *Entry {
	de := new(Entry)
	de.ID = ev.id
	updateEntry(de, ev)
	return de
}

func updateEntry(de *Entry, ev *fileEvent) {
	if ev.ext == "meta" {
		de.MetaSpec = MetaSpecFile
		de.MetaPath = ev.path
		return
	}
	if len(de.ContentExt) != 0 && de.ContentExt != ev.ext {
		de.Duplicates = true
		return
	}
	if de.MetaSpec != MetaSpecFile {
		if ev.ext == "zettel" {
			de.MetaSpec = MetaSpecHeader
		} else {
			de.MetaSpec = MetaSpecNone
		}
	}
	de.ContentPath = ev.path
	de.ContentExt = ev.ext
}

type dirMap map[domain.ZettelID]*Entry

func dirMapUpdate(dm dirMap, ev *fileEvent) {
	de := dm[ev.id]
	if de == nil {
		dm[ev.id] = newEntry(ev)
		return
	}
	updateEntry(de, ev)
}

// directoryService is the main service.
func (srv *Service) directoryService(events <-chan *fileEvent, ready chan<- int) {
	curMap := make(dirMap)
	var newMap dirMap
	for {
		select {
		case ev, ok := <-events:
			if !ok {
				return
			}
			switch ev.status {
			case fileStatusReloadStart:
				newMap = make(dirMap)
			case fileStatusReloadEnd:
				curMap = newMap
				newMap = nil
				if ready != nil {
					ready <- len(curMap)
					close(ready)
					ready = nil
				}
				srv.notifyChange(true, domain.InvalidZettelID)
			case fileStatusError:
				log.Println("FILESTORE", "ERROR", ev.err)
			case fileStatusUpdate:
				if newMap != nil {
					dirMapUpdate(newMap, ev)
				} else {
					dirMapUpdate(curMap, ev)
					srv.notifyChange(false, ev.id)
				}
			case fileStatusDelete:
				if newMap != nil {
					delete(newMap, ev.id)
				} else {
					delete(curMap, ev.id)
					srv.notifyChange(false, ev.id)
				}
			}
		case cmd, ok := <-srv.cmds:
			if ok {
				cmd.run(curMap)
			}
		}
	}
}

type dirCmd interface {
	run(m dirMap)
}

type cmdGetEntries struct {
	result chan<- resGetEntries
}
type resGetEntries []Entry

func (cmd *cmdGetEntries) run(m dirMap) {
	res := make([]Entry, 0, len(m))
	for _, de := range m {
		res = append(res, *de)
	}
	cmd.result <- res
}

type cmdGetEntry struct {
	id     domain.ZettelID
	result chan<- resGetEntry
}
type resGetEntry = Entry

func (cmd *cmdGetEntry) run(m dirMap) {
	entry := m[cmd.id]
	if entry == nil {
		cmd.result <- Entry{ID: domain.InvalidZettelID}
	} else {
		cmd.result <- *entry
	}
}

type cmdNewEntry struct {
	result chan<- resNewEntry
}
type resNewEntry = Entry

func (cmd *cmdNewEntry) run(m dirMap) {
	id := domain.NewZettelID(false)
	if _, ok := m[id]; !ok {
		entry := &Entry{ID: id, MetaSpec: MetaSpecUnknown}
		m[id] = entry
		cmd.result <- *entry
		return
	}
	for {
		id = domain.NewZettelID(true)
		if _, ok := m[id]; !ok {
			entry := &Entry{ID: id, MetaSpec: MetaSpecUnknown}
			m[id] = entry
			cmd.result <- *entry
			return
		}
		// TODO: do not wait here, but in a non-blocking goroutine.
		time.Sleep(100 * time.Millisecond)
	}
}

type cmdUpdateEntry struct {
	entry  *Entry
	result chan<- struct{}
}

func (cmd *cmdUpdateEntry) run(m dirMap) {
	entry := *cmd.entry
	m[entry.ID] = &entry
	cmd.result <- struct{}{}
}

type cmdRenameEntry struct {
	curEntry *Entry
	newEntry *Entry
	result   chan<- resRenameEntry
}

type resRenameEntry = error

func (cmd *cmdRenameEntry) run(m dirMap) {
	newEntry := *cmd.newEntry
	newID := newEntry.ID
	if _, found := m[newID]; found {
		cmd.result <- &store.ErrInvalidID{ID: newID}
		return
	}
	delete(m, cmd.curEntry.ID)
	m[newID] = &newEntry
	cmd.result <- nil
}

type cmdDeleteEntry struct {
	id     domain.ZettelID
	result chan<- struct{}
}

func (cmd *cmdDeleteEntry) run(m dirMap) {
	delete(m, cmd.id)
	cmd.result <- struct{}{}
}
