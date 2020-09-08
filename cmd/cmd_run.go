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

package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"net/http"
	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
	"zettelstore.de/z/store"
	"zettelstore.de/z/store/chainstore"
	"zettelstore.de/z/store/filestore"
	"zettelstore.de/z/store/gostore"
	"zettelstore.de/z/usecase"
	"zettelstore.de/z/web/adapter"
	"zettelstore.de/z/web/router"
	"zettelstore.de/z/web/session"
)

// ---------- Subcommand: run ------------------------------------------------

func runFunc(cfg *domain.Meta) (int, error) {
	cs, exitCode, err := setupStores(cfg)
	if cs == nil {
		return exitCode, err
	}
	readonly := cfg.GetBool("readonly")
	router := setupRouting(cs, readonly)

	listenAddr, _ := cfg.Get("listen-addr")
	v := config.GetVersion()
	log.Printf("%v %v", v.Prog, v.Build)
	if cfg.GetBool("verbose") {
		log.Println("Configuration")
		cfg.Write(os.Stderr)
	} else {
		log.Printf("Listening on %v", listenAddr)
		log.Printf("Zettel location %q", cs.Location())
		if readonly {
			log.Println("Read-only mode")
		}
	}
	log.Fatal(http.ListenAndServe(listenAddr, router))
	return 0, nil
}

func setupStores(cfg *domain.Meta) (store.Store, int, error) {
	var stores []store.Store
	cnt := 1
	for {
		dir, ok := cfg.Get(fmt.Sprintf("store-%v-dir", cnt))
		if !ok {
			break
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Unable to create zettel directory %q\n", dir)
			return nil, 2, err
		}
		fs, err := filestore.NewStore(dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to create filestore for %q\n", dir)
			return nil, 2, err
		}
		stores = append(stores, fs)
		cnt++
	}
	if len(stores) == 0 {
		fmt.Fprintln(os.Stderr, "No directory for storing zettel specified.")
		return nil, 2, nil
	}
	stores = append(stores, gostore.NewStore())
	cs := chainstore.NewStore(stores...)
	if err := cs.Start(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, "Unable to start zettel store")
		return nil, 2, err
	}
	config.SetupConfiguration(cs)
	return cs, 0, nil
}

func setupRouting(s store.Store, readonly bool) http.Handler {
	te := adapter.NewTemplateEngine(s)

	ucGetMeta := usecase.NewGetMeta(s)
	ucGetZettel := usecase.NewGetZettel(s)
	listHTMLMetaHandler := adapter.MakeListHTMLMetaHandler('h', te, usecase.NewListMeta(s))
	getHTMLZettelHandler := adapter.MakeGetHTMLZettelHandler('h', te, ucGetZettel, ucGetMeta)
	getNewZettelHandler := adapter.MakeGetNewZettelHandler(te, ucGetZettel)
	postNewZettelHandler := adapter.MakePostNewZettelHandler(usecase.NewNewZettel(s))

	router := router.NewRouter()
	router.Handle("/", adapter.MakeGetRootHandler(s, listHTMLMetaHandler, getHTMLZettelHandler))
	if !readonly {
		router.AddListRoute('a', http.MethodGet, adapter.MakeGetLoginHandler(te))
		router.AddListRoute('a', http.MethodPost, adapter.MakePostLoginHandler(te, usecase.NewAuthenticate(s)))
	}
	router.AddZettelRoute('b', http.MethodGet, adapter.MakeGetBodyHandler('b', te, ucGetZettel, ucGetMeta))
	router.AddListRoute('c', http.MethodGet, adapter.MakeReloadHandler(usecase.NewReload(s)))
	router.AddZettelRoute('c', http.MethodGet, adapter.MakeGetContentHandler(ucGetZettel))
	if !readonly {
		router.AddZettelRoute('d', http.MethodGet, adapter.MakeGetDeleteZettelHandler(te, ucGetZettel))
		router.AddZettelRoute('d', http.MethodPost, adapter.MakePostDeleteZettelHandler(usecase.NewDeleteZettel(s)))
		router.AddZettelRoute('e', http.MethodGet, adapter.MakeEditGetZettelHandler(te, ucGetZettel))
		router.AddZettelRoute('e', http.MethodPost, adapter.MakeEditSetZettelHandler(usecase.NewUpdateZettel(s)))
	}
	router.AddListRoute('h', http.MethodGet, listHTMLMetaHandler)
	router.AddZettelRoute('h', http.MethodGet, getHTMLZettelHandler)
	router.AddZettelRoute('i', http.MethodGet, adapter.MakeGetInfoHandler(te, ucGetZettel, ucGetMeta))
	router.AddZettelRoute('m', http.MethodGet, adapter.MakeGetMetaHandler(ucGetMeta))
	if !readonly {
		router.AddZettelRoute('n', http.MethodGet, getNewZettelHandler)
		router.AddZettelRoute('n', http.MethodPost, postNewZettelHandler)
	}
	router.AddListRoute('r', http.MethodGet, adapter.MakeListRoleHandler(te, usecase.NewListRole(s)))
	if !readonly {
		router.AddZettelRoute('r', http.MethodGet, adapter.MakeGetRenameZettelHandler(te, ucGetMeta))
		router.AddZettelRoute('r', http.MethodPost, adapter.MakePostRenameZettelHandler(usecase.NewRenameZettel(s)))
	}
	router.AddListRoute('t', http.MethodGet, adapter.MakeListTagsHandler(te, usecase.NewListTags(s)))
	router.AddListRoute('s', http.MethodGet, adapter.MakeSearchHandler(te, usecase.NewSearch(s)))
	router.AddListRoute('z', http.MethodGet, adapter.MakeListMetaHandler('z', te, usecase.NewListMeta(s)))
	router.AddZettelRoute('z', http.MethodGet, adapter.MakeGetZettelHandler('z', te, ucGetZettel, ucGetMeta))
	return session.NewHandler(router, usecase.NewGetUserByZid(s))
}
