//-----------------------------------------------------------------------------
// Copyright (c) 2020 Detlef Stern
//
// This file is part of zettelstore.
//
// Zettelstore is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

// Package dirplace provides a directory-based zettel place.
package dirplace

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
	"zettelstore.de/z/place"
	"zettelstore.de/z/place/dirplace/directory"
)

func init() {
	place.Register("dir", func(u *url.URL, next place.Place) (place.Place, error) {
		path := getDirPath(u)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return nil, err
		}
		dp := dirPlace{
			u:         u,
			readonly:  getQueryBool(u, "readonly"),
			next:      next,
			dir:       path,
			dirRescan: time.Duration(getQueryInt(u, "rescan", 60, 600, 30*24*60*60)) * time.Second,
			fSrvs:     uint32(getQueryInt(u, "worker", 1, 17, 1499)),
		}
		dp.cacheChange(true, domain.InvalidZettelID)
		return &dp, nil
	})
}

func getDirPath(u *url.URL) string {
	if u.Opaque != "" {
		return filepath.Clean(u.Opaque)
	}
	return filepath.Clean(u.Path)
}

func getQueryBool(u *url.URL, key string) bool {
	_, ok := u.Query()[key]
	return ok
}

func getQueryInt(u *url.URL, key string, min, def, max int) int {
	sVal := u.Query().Get(key)
	if sVal == "" {
		return def
	}
	iVal, err := strconv.Atoi(sVal)
	if err != nil {
		return def
	}
	if iVal < min {
		return min
	}
	if iVal > max {
		return max
	}
	return iVal
}

// dirPlace uses a directory to store zettel as files.
type dirPlace struct {
	u          *url.URL
	readonly   bool
	next       place.Place
	observers  []place.ObserverFunc
	mxObserver sync.RWMutex
	dir        string
	dirRescan  time.Duration
	dirSrv     *directory.Service
	fSrvs      uint32
	fCmds      []chan fileCmd
	mxCmds     sync.RWMutex
	metaCache  map[domain.ZettelID]*domain.Meta
	mxCache    sync.RWMutex
}

func (dp *dirPlace) isStopped() bool {
	return dp.dirSrv == nil
}

func (dp *dirPlace) Next() place.Place { return dp.next }

func (dp *dirPlace) Location() string {
	return dp.u.String()
}

func (dp *dirPlace) Start(ctx context.Context) error {
	if !dp.isStopped() {
		panic("Calling dirplace.Start() twice.")
	}
	if dp.next != nil {
		if err := dp.next.Start(ctx); err != nil {
			return err
		}
	}
	return dp.localStart(ctx)
}

func (dp *dirPlace) localStart(ctx context.Context) error {
	dp.mxCmds.Lock()
	dp.fCmds = make([]chan fileCmd, 0, dp.fSrvs)
	for i := uint32(0); i < dp.fSrvs; i++ {
		cc := make(chan fileCmd)
		go fileService(i, cc)
		dp.fCmds = append(dp.fCmds, cc)
	}
	dp.dirSrv = directory.NewService(dp.dir, dp.dirRescan)
	dp.mxCmds.Unlock()
	dp.dirSrv.Subscribe(dp.notifyChanged)
	dp.dirSrv.Start()
	return nil
}

func (dp *dirPlace) notifyChanged(all bool, zid domain.ZettelID) {
	dp.cacheChange(all, zid)
	dp.mxObserver.RLock()
	observers := dp.observers
	dp.mxObserver.RUnlock()
	for _, ob := range observers {
		ob(all, zid)
	}
}

func (dp *dirPlace) getFileChan(zid domain.ZettelID) chan fileCmd {
	/* Based on https://en.wikipedia.org/wiki/Fowler%E2%80%93Noll%E2%80%93Vo_hash_function */
	var sum uint32 = 2166136261 ^ uint32(zid)
	sum *= 16777619
	sum ^= uint32(zid >> 32)
	sum *= 16777619

	dp.mxCmds.RLock()
	defer dp.mxCmds.RUnlock()
	return dp.fCmds[sum%dp.fSrvs]
}

func (dp *dirPlace) Stop(ctx context.Context) error {
	if dp.isStopped() {
		return place.ErrStopped
	}
	if err := dp.localStop(ctx); err != nil {
		return err
	}
	if dp.next != nil {
		return dp.next.Stop(ctx)
	}
	return nil
}

func (dp *dirPlace) localStop(ctx context.Context) error {

	dirSrv := dp.dirSrv
	dp.dirSrv = nil
	dirSrv.Stop()
	for _, c := range dp.fCmds {
		close(c)
	}
	return nil
}

// RegisterChangeObserver registers an observer that will be notified
// if a zettel was found to be changed.
// possibly changed.
func (dp *dirPlace) RegisterChangeObserver(f place.ObserverFunc) {
	if dp.next != nil {
		dp.next.RegisterChangeObserver(f)
	}

	dp.mxObserver.Lock()
	dp.observers = append(dp.observers, f)
	dp.mxObserver.Unlock()
}

func (dp *dirPlace) CanCreateZettel(ctx context.Context) bool {
	return !dp.isStopped() && !dp.readonly
}

func (dp *dirPlace) CreateZettel(ctx context.Context, zettel domain.Zettel) (domain.ZettelID, error) {
	if dp.isStopped() {
		return domain.InvalidZettelID, place.ErrStopped
	}
	if dp.readonly {
		return domain.InvalidZettelID, place.ErrReadOnly
	}

	meta := zettel.Meta
	entry := dp.dirSrv.GetNew()
	meta.Zid = entry.Zid
	dp.updateEntryFromMeta(&entry, meta)

	rc := make(chan resSetZettel)
	dp.getFileChan(meta.Zid) <- &fileSetZettel{&entry, zettel, rc}
	err := <-rc
	close(rc)
	if err == nil {
		dp.dirSrv.UpdateEntry(&entry)

		// Make meta available, because place may need some time to update directory.
		dp.cacheSetMeta(zettel.Meta)
	}
	return meta.Zid, err
}

// GetZettel reads the zettel from a file.
func (dp *dirPlace) GetZettel(ctx context.Context, zid domain.ZettelID) (domain.Zettel, error) {
	if dp.isStopped() {
		return domain.Zettel{}, place.ErrStopped
	}

	entry := dp.dirSrv.GetEntry(zid)
	if !entry.IsValid() {
		if dp.next != nil {
			return dp.next.GetZettel(ctx, zid)
		}
		return domain.Zettel{}, &place.ErrUnknownID{Zid: zid}
	}

	rc := make(chan resGetMetaContent)
	dp.getFileChan(zid) <- &fileGetMetaContent{&entry, rc}
	res := <-rc
	close(rc)

	if res.err != nil {
		return domain.Zettel{}, res.err
	}
	dp.cleanupMeta(ctx, res.meta)
	zettel := domain.Zettel{Meta: res.meta, Content: domain.NewContent(res.content)}
	dp.cacheSetMeta(res.meta)
	return zettel, nil
}

// GetMeta retrieves just the meta data of a specific zettel.
func (dp *dirPlace) GetMeta(ctx context.Context, zid domain.ZettelID) (*domain.Meta, error) {
	if dp.isStopped() {
		return nil, place.ErrStopped
	}
	meta, ok := dp.cacheGetMeta(zid)
	if ok {
		return meta, nil
	}
	entry := dp.dirSrv.GetEntry(zid)
	if !entry.IsValid() {
		if dp.next != nil {
			return dp.next.GetMeta(ctx, zid)
		}
		return nil, &place.ErrUnknownID{Zid: zid}
	}

	rc := make(chan resGetMeta)
	dp.getFileChan(zid) <- &fileGetMeta{&entry, rc}
	res := <-rc
	close(rc)

	if res.err != nil {
		return nil, res.err
	}
	dp.cleanupMeta(ctx, res.meta)
	dp.cacheSetMeta(res.meta)
	return res.meta, nil
}

// SelectMeta returns all zettel meta data that match the selection
// criteria. The result is ordered by descending zettel id.
func (dp *dirPlace) SelectMeta(ctx context.Context, f *place.Filter, s *place.Sorter) (res []*domain.Meta, err error) {
	if dp.isStopped() {
		return nil, place.ErrStopped
	}

	hasMatch := place.CreateFilterFunc(f)
	entries := dp.dirSrv.GetEntries()
	rc := make(chan resGetMeta)
	res = make([]*domain.Meta, 0, len(entries))
	for _, entry := range entries {
		meta, ok := dp.cacheGetMeta(entry.Zid)
		if !ok {
			dp.getFileChan(entry.Zid) <- &fileGetMeta{&entry, rc}

			// Response processing could be done by separate goroutine, so that
			// requests can be executed concurrently.
			res := <-rc

			if res.err != nil {
				continue
			}
			meta = res.meta
			dp.cleanupMeta(ctx, meta)
			dp.cacheSetMeta(meta)
		}

		if hasMatch(meta) {
			res = append(res, meta)
		}
	}
	close(rc)
	if err != nil {
		return nil, err
	}
	if dp.next != nil {
		other, err := dp.next.SelectMeta(ctx, f, nil)
		if err != nil {
			return nil, err
		}
		return place.MergeSorted(place.ApplySorter(res, nil), other, s), err
	}
	return place.ApplySorter(res, s), nil
}

func (dp *dirPlace) CanUpdateZettel(ctx context.Context, zettel domain.Zettel) bool {
	return !dp.isStopped() && !dp.readonly
}

func (dp *dirPlace) UpdateZettel(ctx context.Context, zettel domain.Zettel) error {
	if dp.isStopped() {
		return place.ErrStopped
	}
	if dp.readonly {
		return place.ErrReadOnly
	}

	meta := zettel.Meta
	if !meta.Zid.IsValid() {
		return &place.ErrInvalidID{Zid: meta.Zid}
	}
	entry := dp.dirSrv.GetEntry(meta.Zid)
	if !entry.IsValid() {
		// Existing zettel, but new in this place.
		entry.Zid = meta.Zid
		dp.updateEntryFromMeta(&entry, meta)
	} else if entry.MetaSpec == directory.MetaSpecNone {
		if defaultMeta := entry.CalcDefaultMeta(); !meta.Equal(defaultMeta) {
			dp.updateEntryFromMeta(&entry, meta)
			dp.dirSrv.UpdateEntry(&entry)
		}
	}
	dp.notifyChanged(false, meta.Zid)

	rc := make(chan resSetZettel)
	dp.getFileChan(meta.Zid) <- &fileSetZettel{&entry, zettel, rc}
	err := <-rc
	close(rc)
	return err
}

func (dp *dirPlace) updateEntryFromMeta(entry *directory.Entry, meta *domain.Meta) {
	entry.MetaSpec, entry.ContentExt = calcSpecExt(meta)
	basePath := filepath.Join(dp.dir, entry.Zid.Format())
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

func (dp *dirPlace) CanRenameZettel(ctx context.Context, zid domain.ZettelID) bool {
	if dp.isStopped() || dp.readonly {
		return false
	}
	entry := dp.dirSrv.GetEntry(zid)
	canLocalRename := entry.IsValid()
	return canLocalRename && (dp.next == nil || dp.next.CanRenameZettel(ctx, zid))
}

// Rename changes the current zettel id to a new zettel id.
func (dp *dirPlace) RenameZettel(ctx context.Context, curZid, newZid domain.ZettelID) error {
	if dp.isStopped() {
		return place.ErrStopped
	}
	if dp.readonly {
		return place.ErrReadOnly
	}
	if curZid == newZid {
		return nil
	}
	curEntry := dp.dirSrv.GetEntry(curZid)
	if !curEntry.IsValid() {
		if dp.next != nil {
			return dp.next.RenameZettel(ctx, curZid, newZid)
		}
		return nil
	}

	// Check whether zettel with new ID already exists in this place or in next places
	if _, err := dp.GetMeta(ctx, newZid); err == nil {
		return &place.ErrInvalidID{Zid: newZid}
	}

	newEntry := directory.Entry{
		Zid:         newZid,
		MetaSpec:    curEntry.MetaSpec,
		MetaPath:    renamePath(curEntry.MetaPath, curZid, newZid),
		ContentPath: renamePath(curEntry.ContentPath, curZid, newZid),
		ContentExt:  curEntry.ContentExt,
	}
	dp.notifyChanged(false, curZid)
	if err := dp.dirSrv.RenameEntry(&curEntry, &newEntry); err != nil {
		return err
	}

	rc := make(chan resRenameZettel)
	dp.getFileChan(newZid) <- &fileRenameZettel{&curEntry, &newEntry, rc}
	err := <-rc
	close(rc)
	return err
}

func (dp *dirPlace) CanDeleteZettel(ctx context.Context, zid domain.ZettelID) bool {
	if dp.isStopped() || dp.readonly {
		return false
	}
	entry := dp.dirSrv.GetEntry(zid)
	return entry.IsValid() || (dp.next != nil && dp.next.CanDeleteZettel(ctx, zid))
}

// DeleteZettel removes the zettel from the place.
func (dp *dirPlace) DeleteZettel(ctx context.Context, zid domain.ZettelID) error {
	if dp.isStopped() {
		return place.ErrStopped
	}
	if dp.readonly {
		return place.ErrReadOnly
	}

	entry := dp.dirSrv.GetEntry(zid)
	if !entry.IsValid() {
		dp.notifyChanged(false, zid)
		return nil
	}
	dp.dirSrv.DeleteEntry(zid)
	rc := make(chan resDeleteZettel)
	dp.getFileChan(zid) <- &fileDeleteZettel{&entry, rc}
	err := <-rc
	close(rc)
	dp.notifyChanged(false, zid)
	return err
}

// Reload clears all caches, reloads all internal data to reflect changes
// that were possibly undetected.
func (dp *dirPlace) Reload(ctx context.Context) error {
	if dp.isStopped() {
		return place.ErrStopped
	}

	// Brute force: stop everything, then start everything.
	// Could be done better in the future...
	err := dp.localStop(ctx)
	if err == nil {
		err = dp.localStart(ctx)
	}
	if dp.next != nil {
		err1 := dp.next.Reload(ctx)
		if err == nil {
			err = err1
		}
	}
	return err
}

func (dp *dirPlace) cleanupMeta(ctx context.Context, meta *domain.Meta) {
	if role, ok := meta.Get(domain.MetaKeyRole); !ok || role == "" {
		meta.Set(domain.MetaKeyRole, config.GetDefaultRole())
	}
	if syntax, ok := meta.Get(domain.MetaKeySyntax); !ok || syntax == "" {
		meta.Set(domain.MetaKeySyntax, config.GetDefaultSyntax())
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

func (dp *dirPlace) cacheChange(all bool, zid domain.ZettelID) {
	dp.mxCache.Lock()
	if all {
		dp.metaCache = make(map[domain.ZettelID]*domain.Meta, len(dp.metaCache))
	} else {
		delete(dp.metaCache, zid)
	}
	dp.mxCache.Unlock()
}

func (dp *dirPlace) cacheSetMeta(meta *domain.Meta) {
	dp.mxCache.Lock()
	meta.Freeze()
	dp.metaCache[meta.Zid] = meta
	dp.mxCache.Unlock()
}

func (dp *dirPlace) cacheGetMeta(zid domain.ZettelID) (*domain.Meta, bool) {
	dp.mxCache.RLock()
	meta, ok := dp.metaCache[zid]
	dp.mxCache.RUnlock()
	return meta, ok
}
