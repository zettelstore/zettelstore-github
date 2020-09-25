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
	"net/http"
	"os"

	"zettelstore.de/z/auth/policy"
	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
	"zettelstore.de/z/store"
	"zettelstore.de/z/store/chainstore"
	"zettelstore.de/z/store/policystore"
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
	var stores []store.Store = nil
	hasGlobals := false
	for cnt := 1; ; cnt++ {
		key := fmt.Sprintf("place-%v-uri", cnt)
		uri, ok := cfg.Get(key)
		if !ok || uri == "" {
			break
		}
		s, err := store.Connect(uri)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to use store with URI %q\n", uri)
			return nil, 2, err
		}
		stores = append(stores, s)
		if s.Location() == "globals:" {
			hasGlobals = true
		}
	}

	if !hasGlobals {
		globals, err := store.Connect("globals:")
		if err != nil {
			return nil, 2, err
		}
		stores = append(stores, globals)
	}
	cs := chainstore.NewStore(stores...)
	if err := cs.Start(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, "Unable to start zettel store")
		return nil, 2, err
	}
	config.SetupConfiguration(cs)
	return cs, 0, nil
}

func setupRouting(us store.Store, readonly bool) http.Handler {
	ps := us
	var pol policy.Policy
	if config.WithAuth() || readonly {
		pol = policy.NewPolicy("default")
		ps = policystore.NewStore(us, pol)
	} else {
		pol = policy.NewPolicy("all")
	}
	te := adapter.NewTemplateEngine(us, pol)

	ucGetMeta := usecase.NewGetMeta(ps)
	ucGetZettel := usecase.NewGetZettel(ps)
	listHTMLMetaHandler := adapter.MakeListHTMLMetaHandler(te, usecase.NewListMeta(ps))
	getHTMLZettelHandler := adapter.MakeGetHTMLZettelHandler(te, ucGetZettel, ucGetMeta)

	router := router.NewRouter()
	router.Handle("/", adapter.MakeGetRootHandler(ps, listHTMLMetaHandler, getHTMLZettelHandler))
	router.AddListRoute('a', http.MethodGet, adapter.MakeGetLoginHandler(te))
	router.AddListRoute('a', http.MethodPost, adapter.MakePostLoginHandler(te, usecase.NewAuthenticate(us)))
	router.AddZettelRoute('a', http.MethodGet, adapter.MakeGetLogoutHandler())
	router.AddListRoute('c', http.MethodGet, adapter.MakeReloadHandler(usecase.NewReload(ps)))
	if !readonly {
		router.AddZettelRoute('d', http.MethodGet, adapter.MakeGetDeleteZettelHandler(te, ucGetZettel))
		router.AddZettelRoute('d', http.MethodPost, adapter.MakePostDeleteZettelHandler(usecase.NewDeleteZettel(ps)))
		router.AddZettelRoute('e', http.MethodGet, adapter.MakeEditGetZettelHandler(te, ucGetZettel))
		router.AddZettelRoute('e', http.MethodPost, adapter.MakeEditSetZettelHandler(usecase.NewUpdateZettel(ps)))
	}
	router.AddListRoute('h', http.MethodGet, listHTMLMetaHandler)
	router.AddZettelRoute('h', http.MethodGet, getHTMLZettelHandler)
	router.AddZettelRoute('i', http.MethodGet, adapter.MakeGetInfoHandler(te, ucGetZettel, ucGetMeta))
	if !readonly {
		router.AddZettelRoute('n', http.MethodGet, adapter.MakeGetNewZettelHandler(te, ucGetZettel))
		router.AddZettelRoute('n', http.MethodPost, adapter.MakePostNewZettelHandler(usecase.NewNewZettel(ps)))
	}
	router.AddListRoute('r', http.MethodGet, adapter.MakeListRoleHandler(te, usecase.NewListRole(ps)))
	if !readonly {
		router.AddZettelRoute('r', http.MethodGet, adapter.MakeGetRenameZettelHandler(te, ucGetMeta))
		router.AddZettelRoute('r', http.MethodPost, adapter.MakePostRenameZettelHandler(usecase.NewRenameZettel(ps)))
	}
	router.AddListRoute('t', http.MethodGet, adapter.MakeListTagsHandler(te, usecase.NewListTags(ps)))
	router.AddListRoute('s', http.MethodGet, adapter.MakeSearchHandler(te, usecase.NewSearch(ps)))
	router.AddListRoute('z', http.MethodGet, adapter.MakeListMetaHandler(te, usecase.NewListMeta(ps)))
	router.AddZettelRoute('z', http.MethodGet, adapter.MakeGetZettelHandler(te, ucGetZettel, ucGetMeta))
	return session.NewHandler(router, usecase.NewGetUserByZid(us))
}
