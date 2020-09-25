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
	"net/url"
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

func init() {
	store.Register("dir", useStore)
}

// fileStore uses a directory to store zettel as files.
type fileStore struct {
	u          *url.URL
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

func useStore(u *url.URL) (store.Store, error) {
	var path string
	if u.Opaque != "" {
		path = u.Opaque
	} else {
		path = u.Path
	}
	path = filepath.Clean(path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}
	fs := &fileStore{
		u:         u,
		dir:       path,
		dirReload: 600 * time.Second, // TODO: make configurable
		fSrvs:     17,                // TODO: make configurable
	}
	fs.cacheChange(true, domain.InvalidZettelID)
	return fs, nil
}

func (fs *fileStore) isStopped() bool {
	return fs.dirSrv == nil
}

// Location returns the directory path of the file store.
func (fs *fileStore) Location() string {
	return fs.u.String()
}

// Start the file store.
func (fs *fileStore) Start(ctx context.Context) error {
	if !fs.isStopped() {
		panic("Calling filestore.Start() twice.")
	}
	fs.mxCmds.Lock()
	fs.fCmds = make([]chan fileCmd, 0, fs.fSrvs)
	for i := uint32(0); i < fs.fSrvs; i++ {
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

func (fs *fileStore) notifyChanged(all bool, zid domain.ZettelID) {
	fs.cacheChange(all, zid)
	fs.mxObserver.RLock()
	observers := fs.observers
	fs.mxObserver.RUnlock()
	for _, ob := range observers {
		ob(all, zid)
	}
}

func (fs *fileStore) getFileChan(zid domain.ZettelID) chan fileCmd {
	/* Based on https://en.wikipedia.org/wiki/Fowler%E2%80%93Noll%E2%80%93Vo_hash_function */
	var sum uint32 = 2166136261 ^ uint32(zid)
	sum *= 16777619
	sum ^= uint32(zid >> 32)
	sum *= 16777619

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
// if a zettel was found to be changed.
// possibly changed.
func (fs *fileStore) RegisterChangeObserver(f store.ObserverFunc) {
	fs.mxObserver.Lock()
	fs.observers = append(fs.observers, f)
	fs.mxObserver.Unlock()
}

func (fs *fileStore) CreateZettel(ctx context.Context, zettel domain.Zettel) (domain.ZettelID, error) {
	if fs.isStopped() {
		return domain.InvalidZettelID, store.ErrStopped
	}

	meta := zettel.Meta

	entry := fs.dirSrv.GetNew()
	meta.Zid = entry.Zid
	fs.updateEntryFromMeta(&entry, meta)

	rc := make(chan resSetZettel)
	fs.getFileChan(meta.Zid) <- &fileSetZettel{&entry, zettel, rc}
	err := <-rc
	close(rc)
	if err == nil {
		fs.dirSrv.UpdateEntry(&entry)

		// Make meta available, because file store may need some time to update directory.
		fs.cacheSetMeta(zettel.Meta)
	}
	return meta.Zid, err
}

// GetZettel reads the zettel from a file.
func (fs *fileStore) GetZettel(ctx context.Context, zid domain.ZettelID) (domain.Zettel, error) {
	if fs.isStopped() {
		return domain.Zettel{}, store.ErrStopped
	}

	entry := fs.dirSrv.GetEntry(zid)
	if !entry.IsValid() {
		return domain.Zettel{}, &store.ErrUnknownID{Zid: zid}
	}

	rc := make(chan resGetMetaContent)
	fs.getFileChan(zid) <- &fileGetMetaContent{&entry, rc}
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
func (fs *fileStore) GetMeta(ctx context.Context, zid domain.ZettelID) (*domain.Meta, error) {
	if fs.isStopped() {
		return nil, store.ErrStopped
	}
	meta, ok := fs.cacheGetMeta(zid)
	if ok {
		return meta, nil
	}
	entry := fs.dirSrv.GetEntry(zid)
	if !entry.IsValid() {
		return nil, &store.ErrUnknownID{Zid: zid}
	}

	rc := make(chan resGetMeta)
	fs.getFileChan(zid) <- &fileGetMeta{&entry, rc}
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
		meta, ok := fs.cacheGetMeta(entry.Zid)
		if !ok {
			fs.getFileChan(entry.Zid) <- &fileGetMeta{&entry, rc}

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

func (fs *fileStore) UpdateZettel(ctx context.Context, zettel domain.Zettel) error {
	if fs.isStopped() {
		return store.ErrStopped
	}

	meta := zettel.Meta
	if !meta.Zid.IsValid() {
		return &store.ErrInvalidID{Zid: meta.Zid}
	}
	entry := fs.dirSrv.GetEntry(meta.Zid)
	if !entry.IsValid() {
		// Existing zettel, but new in this store.
		entry.Zid = meta.Zid
		fs.updateEntryFromMeta(&entry, meta)
	}
	fs.notifyChanged(false, meta.Zid)

	rc := make(chan resSetZettel)
	fs.getFileChan(meta.Zid) <- &fileSetZettel{&entry, zettel, rc}
	err := <-rc
	close(rc)
	return err
}

func (fs *fileStore) updateEntryFromMeta(entry *directory.Entry, meta *domain.Meta) {
	entry.MetaSpec, entry.ContentExt = calcSpecExt(meta)
	basePath := filepath.Join(fs.dir, entry.Zid.Format())
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
	for _, s := range config.GetZettelFileSyntax() {
		if s == syntax {
			return directory.MetaSpecHeader, "zettel"
		}
	}
	return directory.MetaSpecFile, syntax
}

// Rename changes the current id to a new id.
func (fs *fileStore) RenameZettel(ctx context.Context, curZid, newZid domain.ZettelID) error {
	if fs.isStopped() {
		return store.ErrStopped
	}
	curEntry := fs.dirSrv.GetEntry(curZid)
	if !curEntry.IsValid() {
		return &store.ErrUnknownID{Zid: curZid}
	}
	if curZid == newZid {
		return nil
	}
	newEntry := directory.Entry{
		Zid:         newZid,
		MetaSpec:    curEntry.MetaSpec,
		MetaPath:    renamePath(curEntry.MetaPath, curZid, newZid),
		ContentPath: renamePath(curEntry.ContentPath, curZid, newZid),
		ContentExt:  curEntry.ContentExt,
	}
	fs.notifyChanged(false, curZid)
	if err := fs.dirSrv.RenameEntry(&curEntry, &newEntry); err != nil {
		return err
	}

	rc := make(chan resRenameZettel)
	fs.getFileChan(newZid) <- &fileRenameZettel{&curEntry, &newEntry, rc}
	err := <-rc
	close(rc)
	return err
}

// DeleteZettel removes the zettel from the store.
func (fs *fileStore) DeleteZettel(ctx context.Context, zid domain.ZettelID) error {
	if fs.isStopped() {
		return store.ErrStopped
	}

	entry := fs.dirSrv.GetEntry(zid)
	if !entry.IsValid() {
		fs.notifyChanged(false, zid)
		return nil
	}
	fs.dirSrv.DeleteEntry(zid)
	rc := make(chan resDeleteZettel)
	fs.getFileChan(zid) <- &fileDeleteZettel{&entry, rc}
	err := <-rc
	close(rc)
	fs.notifyChanged(false, zid)
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
		meta.Set(domain.MetaKeySyntax, config.GetDefaultSyntax())
	}
	if role, ok := meta.Get(domain.MetaKeyRole); !ok || role == "" {
		meta.Set(domain.MetaKeyRole, config.GetDefaultRole())
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

func (fs *fileStore) cacheChange(all bool, zid domain.ZettelID) {
	fs.mxCache.Lock()
	if all {
		fs.metaCache = make(map[domain.ZettelID]*domain.Meta, len(fs.metaCache))
	} else {
		delete(fs.metaCache, zid)
	}
	fs.mxCache.Unlock()
}

func (fs *fileStore) cacheSetMeta(meta *domain.Meta) {
	fs.mxCache.Lock()
	meta.Freeze()
	fs.metaCache[meta.Zid] = meta
	fs.mxCache.Unlock()
}

func (fs *fileStore) cacheGetMeta(zid domain.ZettelID) (*domain.Meta, bool) {
	fs.mxCache.RLock()
	meta, ok := fs.metaCache[zid]
	fs.mxCache.RUnlock()
	return meta, ok
}
