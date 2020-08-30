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

// Package filestore provides a file based zettel store.
package filestore

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
	"zettelstore.de/z/store"
	"zettelstore.de/z/store/filestore/directory"
)

// NewStore creates and returns a new Store.
func NewStore(dir string) (store.Store, error) {
	var path string
	if len(dir) == 0 {
		path = "." // TODO: make absolute path of current working directory
	} else {
		path = dir
	}
	path = filepath.Clean(path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}
	fs := &fileStore{
		dir:       path,
		dirReload: 600 * time.Second, // TODO: make configurable
		fSrvs:     17,                // TODO: make configurable
	}
	fs.cacheChange(true, domain.ZettelID(0))
	return fs, nil
}

// fileStore uses a directory to store zettel as files.
type fileStore struct {
	parent     store.Store
	observers  []store.ObserverFunc
	mxObserver sync.RWMutex
	dir        string
	dirReload  time.Duration
	dirSrv     *directory.Service
	fSrvs      uint32
	fCmds      []chan fileCmd
	mxCmds     sync.RWMutex
	metaCache  map[domain.ZettelID]*domain.Meta
	mxCache    sync.RWMutex
}

func (fs *fileStore) isStopped() bool {
	return fs.dirSrv == nil
}

// SetParentStore is called when the store is part of a bigger store.
func (fs *fileStore) SetParentStore(parent store.Store) {
	fs.parent = parent
}

// Location returns the directory path of the file store.
func (fs *fileStore) Location() string {
	return fmt.Sprintf("dir://%s", fs.dir)
}

// Start the file store.
func (fs *fileStore) Start(ctx context.Context) error {
	if !fs.isStopped() {
		panic("Calling filestore.Start() twice.")
	}
	fs.mxCmds.Lock()
	fs.fCmds = make([]chan fileCmd, 0, fs.fSrvs)
	var i uint32
	for ; i < fs.fSrvs; i++ {
		cc := make(chan fileCmd)
		go fileService(i, cc)
		fs.fCmds = append(fs.fCmds, cc)
	}
	fs.mxCmds.Unlock()

	fs.dirSrv = directory.NewService(fs.dir, fs.dirReload)
	fs.dirSrv.Subscribe(fs.notifyChanged)
	fs.dirSrv.Start()
	return nil
}

func (fs *fileStore) notifyChanged(all bool, id domain.ZettelID) {
	fs.cacheChange(false, id)
	fs.mxObserver.RLock()
	observers := fs.observers
	fs.mxObserver.RUnlock()
	for _, ob := range observers {
		ob(all, id)
	}
}

func (fs *fileStore) getFileChan(zid domain.ZettelID) chan fileCmd {
	/* Based on https://en.wikipedia.org/wiki/Fowler%E2%80%93Noll%E2%80%93Vo_hash_function */
	id := zid.Format() // TODO: just hash the two uin32
	var sum uint32 = 2166136261
	for i := 0; i < len(id); i++ {
		sum ^= uint32(id[i])
		sum *= 16777619
	}
	fs.mxCmds.RLock()
	defer fs.mxCmds.RUnlock()
	return fs.fCmds[sum%fs.fSrvs]
}

// Stop the file store.
func (fs *fileStore) Stop(ctx context.Context) error {
	if fs.isStopped() {
		return store.ErrStopped
	}
	dirSrv := fs.dirSrv
	fs.dirSrv = nil
	dirSrv.Stop()
	for _, c := range fs.fCmds {
		close(c)
	}
	return nil
}

// RegisterChangeObserver registers an observer that will be notified
// if a zettel was found to be changed. If the id is empty, all zettel are
// possibly changed.
func (fs *fileStore) RegisterChangeObserver(f store.ObserverFunc) {
	fs.mxObserver.Lock()
	fs.observers = append(fs.observers, f)
	fs.mxObserver.Unlock()
}

// GetZettel reads the zettel from a file.
func (fs *fileStore) GetZettel(ctx context.Context, id domain.ZettelID) (domain.Zettel, error) {
	if fs.isStopped() {
		return domain.Zettel{}, store.ErrStopped
	}

	entry := fs.dirSrv.GetEntry(id)
	if !entry.IsValid() {
		return domain.Zettel{}, &store.ErrUnknownID{ID: id}
	}

	rc := make(chan resGetMetaContent)
	fs.getFileChan(id) <- &fileGetMetaContent{&entry, rc}
	res := <-rc
	close(rc)

	if res.err != nil {
		return domain.Zettel{}, res.err
	}
	fs.cleanupMeta(ctx, res.meta)
	zettel := domain.Zettel{Meta: res.meta, Content: domain.NewContent(res.content)}
	fs.cacheSetMeta(res.meta)
	return zettel, nil
}

// GetMeta retrieves just the meta data of a specific zettel.
func (fs *fileStore) GetMeta(ctx context.Context, id domain.ZettelID) (*domain.Meta, error) {
	if fs.isStopped() {
		return nil, store.ErrStopped
	}
	meta, ok := fs.cacheGetMeta(id)
	if ok {
		return meta, nil
	}
	entry := fs.dirSrv.GetEntry(id)
	if !entry.IsValid() {
		return nil, &store.ErrUnknownID{ID: id}
	}

	rc := make(chan resGetMeta)
	fs.getFileChan(id) <- &fileGetMeta{&entry, rc}
	res := <-rc
	close(rc)

	if res.err != nil {
		return nil, res.err
	}
	fs.cleanupMeta(ctx, res.meta)
	fs.cacheSetMeta(res.meta)
	return res.meta, nil
}

// SelectMeta returns all zettel meta data that match the selection
// criteria. The result is ordered by descending zettel id.
func (fs *fileStore) SelectMeta(ctx context.Context, f *store.Filter, s *store.Sorter) (res []*domain.Meta, err error) {
	if fs.isStopped() {
		return nil, store.ErrStopped
	}

	hasMatch := store.CreateFilterFunc(f)
	entries := fs.dirSrv.GetEntries()
	rc := make(chan resGetMeta)
	res = make([]*domain.Meta, 0, len(entries))
	for _, entry := range entries {
		meta, ok := fs.cacheGetMeta(entry.ID)
		if !ok {
			fs.getFileChan(entry.ID) <- &fileGetMeta{&entry, rc}

			// Response processing could be done by separate goroutine, so that
			// requests can be executed concurrently.
			res := <-rc

			if res.err != nil {
				continue
			}
			meta = res.meta
			fs.cleanupMeta(ctx, meta)
			fs.cacheSetMeta(meta)
		}

		if hasMatch(meta) {
			res = append(res, meta)
		}
	}
	close(rc)
	return store.ApplySorter(res, s), err
}

// SetZettel stores new data for a zettel.
func (fs *fileStore) SetZettel(ctx context.Context, zettel domain.Zettel) error {
	if fs.isStopped() {
		return store.ErrStopped
	}

	var entry directory.Entry
	newEntry := false
	meta := zettel.Meta
	if meta.ID.IsValid() {
		// Update existing zettel or create a new one with given ID.
		entry = fs.dirSrv.GetEntry(meta.ID)
		if !entry.IsValid() {
			// Existing zettel, but new in this store.
			entry.ID = meta.ID
			fs.updateEntryFromMeta(&entry, meta)
		}
		fs.notifyChanged(false, meta.ID)
	} else {
		// Calculate a new ID, because of new zettel.
		entry = fs.dirSrv.GetNew()
		meta.ID = entry.ID
		fs.updateEntryFromMeta(&entry, meta)
		newEntry = true
	}

	rc := make(chan resSetZettel)
	fs.getFileChan(meta.ID) <- &fileSetZettel{&entry, zettel, rc}
	err := <-rc
	close(rc)
	if newEntry && err == nil {
		fs.dirSrv.UpdateEntry(&entry)

		// Make meta available, because file store may need some time to update directory.
		fs.cacheSetMeta(zettel.Meta)
	}
	return err
}

func (fs *fileStore) updateEntryFromMeta(entry *directory.Entry, meta *domain.Meta) {
	entry.MetaSpec, entry.ContentExt = calcSpecExt(meta)
	basePath := filepath.Join(fs.dir, entry.ID.Format())
	if entry.MetaSpec == directory.MetaSpecFile {
		entry.MetaPath = basePath + ".meta"
	}
	entry.ContentPath = basePath + "." + entry.ContentExt
	entry.Duplicates = false
}

func calcSpecExt(meta *domain.Meta) (directory.MetaSpec, string) {
	if meta.YamlSep {
		return directory.MetaSpecHeader, "zettel"
	}
	syntax := meta.GetDefault(domain.MetaKeySyntax, "bin")
	switch syntax {
	case "meta", "zmk":
		return directory.MetaSpecHeader, "zettel"
	}
	for _, s := range config.Config.GetZettelFileSyntax() {
		if s == syntax {
			return directory.MetaSpecHeader, "zettel"
		}
	}
	return directory.MetaSpecFile, syntax
}

// Rename changes the current ID to a new ID.
func (fs *fileStore) RenameZettel(ctx context.Context, curID, newID domain.ZettelID) error {
	if fs.isStopped() {
		return store.ErrStopped
	}
	curEntry := fs.dirSrv.GetEntry(curID)
	if !curEntry.IsValid() {
		return &store.ErrUnknownID{ID: curID}
	}
	if curID == newID {
		return nil
	}
	newEntry := directory.Entry{
		ID:          newID,
		MetaSpec:    curEntry.MetaSpec,
		MetaPath:    renamePath(curEntry.MetaPath, curID, newID),
		ContentPath: renamePath(curEntry.ContentPath, curID, newID),
		ContentExt:  curEntry.ContentExt,
	}
	fs.notifyChanged(false, curID)
	if err := fs.dirSrv.RenameEntry(&curEntry, &newEntry); err != nil {
		return err
	}

	rc := make(chan resRenameZettel)
	fs.getFileChan(newID) <- &fileRenameZettel{&curEntry, &newEntry, rc}
	err := <-rc
	close(rc)
	return err
}

// DeleteZettel removes the zettel from the store.
func (fs *fileStore) DeleteZettel(ctx context.Context, id domain.ZettelID) error {
	if fs.isStopped() {
		return store.ErrStopped
	}

	entry := fs.dirSrv.GetEntry(id)
	if !entry.IsValid() {
		fs.notifyChanged(false, id)
		return nil
	}
	fs.dirSrv.DeleteEntry(id)
	rc := make(chan resDeleteZettel)
	fs.getFileChan(id) <- &fileDeleteZettel{&entry, rc}
	err := <-rc
	close(rc)
	fs.notifyChanged(false, id)
	return err
}

// Reload clears all caches, reloads all internal data to reflect changes
// that were possibly undetected.
func (fs *fileStore) Reload(ctx context.Context) error {
	if fs.isStopped() {
		return store.ErrStopped
	}

	// Brute force: stop everything, then start everything.
	// Could be done better in the future...
	err := fs.Stop(ctx)
	if err == nil {
		err = fs.Start(ctx)
	}
	return err
}

func (fs *fileStore) cleanupMeta(ctx context.Context, meta *domain.Meta) {
	if syntax, ok := meta.Get(domain.MetaKeySyntax); !ok || syntax == "" {
		meta.Set(domain.MetaKeySyntax, config.Config.GetDefaultSyntax())
	}
	if role, ok := meta.Get(domain.MetaKeyRole); !ok || role == "" {
		meta.Set(domain.MetaKeyRole, config.Config.GetDefaultRole())
	}
}

func renamePath(path string, curID, newID domain.ZettelID) string {
	dir, file := filepath.Split(path)
	if cur := curID.Format(); strings.HasPrefix(file, cur) {
		file = newID.Format() + file[len(cur):]
		return filepath.Join(dir, file)
	}
	return path
}

func (fs *fileStore) cacheChange(all bool, id domain.ZettelID) {
	fs.mxCache.Lock()
	if all {
		fs.metaCache = make(map[domain.ZettelID]*domain.Meta, len(fs.metaCache))
	} else {
		delete(fs.metaCache, id)
	}
	fs.mxCache.Unlock()
}

func (fs *fileStore) cacheSetMeta(meta *domain.Meta) {
	fs.mxCache.Lock()
	meta.Freeze()
	fs.metaCache[meta.ID] = meta
	fs.mxCache.Unlock()
}

func (fs *fileStore) cacheGetMeta(id domain.ZettelID) (*domain.Meta, bool) {
	fs.mxCache.RLock()
	meta, ok := fs.metaCache[id]
	fs.mxCache.RUnlock()
	return meta, ok
}
