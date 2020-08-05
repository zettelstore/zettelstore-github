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

// Package main is the starting point for the zettel web command.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "zettelstore.de/z/cmd"
	"zettelstore.de/z/config"
	"zettelstore.de/z/parser"
	"zettelstore.de/z/store"
	"zettelstore.de/z/store/chainstore"
	"zettelstore.de/z/store/filestore"
	"zettelstore.de/z/store/gostore"
	"zettelstore.de/z/usecase"
	"zettelstore.de/z/web/adapter"
	"zettelstore.de/z/web/router"
)

// Version variables
var (
	buildVersion   string = ""
	releaseVersion string = ""
)

func setupRouting(s store.Store) *router.Router {
	te := adapter.NewTemplateEngine(s)
	p := parser.New()
	p.InitCache(s)

	ucGetMeta := usecase.NewGetMeta(s)
	ucGetZettel := usecase.NewGetZettel(s)
	listHTMLMetaHandler := adapter.MakeListHTMLMetaHandler('h', te, p, usecase.NewListMeta(s))
	getHTMLZettelHandler := adapter.MakeGetHTMLZettelHandler('h', te, p, ucGetZettel, ucGetMeta)
	getNewZettelHandler := adapter.MakeGetNewZettelHandler(te, ucGetZettel)
	postNewZettelHandler := adapter.MakePostNewZettelHandler(usecase.NewNewZettel(s))

	router := router.NewRouter()
	router.Handle("/", adapter.MakeGetRootHandler(s, listHTMLMetaHandler, getHTMLZettelHandler))
	router.AddListRoute('a', http.MethodGet, adapter.MakeGetLoginHandler(te))
	router.AddListRoute('a', http.MethodPost, adapter.MakePostLoginHandler(usecase.NewAuthenticate(s)))
	router.AddZettelRoute('b', http.MethodGet, adapter.MakeGetBodyHandler('b', te, p, ucGetZettel, ucGetMeta))
	router.AddListRoute('c', http.MethodGet, adapter.MakeReloadHandler(usecase.NewReload(s)))
	router.AddZettelRoute('c', http.MethodGet, adapter.MakeGetContentHandler(ucGetZettel))
	router.AddZettelRoute('d', http.MethodGet, adapter.MakeGetDeleteZettelHandler(te, ucGetZettel))
	router.AddZettelRoute('d', http.MethodPost, adapter.MakePostDeleteZettelHandler(usecase.NewDeleteZettel(s)))
	router.AddZettelRoute('e', http.MethodGet, adapter.MakeEditGetZettelHandler(te, ucGetZettel))
	router.AddZettelRoute('e', http.MethodPost, adapter.MakeEditSetZettelHandler(usecase.NewUpdateZettel(s)))
	router.AddListRoute('h', http.MethodGet, listHTMLMetaHandler)
	router.AddZettelRoute('h', http.MethodGet, getHTMLZettelHandler)
	router.AddZettelRoute('i', http.MethodGet, adapter.MakeGetInfoHandler(te, p, ucGetZettel, ucGetMeta))
	router.AddZettelRoute('m', http.MethodGet, adapter.MakeGetMetaHandler(p, ucGetMeta))
	router.AddListRoute('n', http.MethodGet, getNewZettelHandler)
	router.AddListRoute('n', http.MethodPost, postNewZettelHandler)
	router.AddZettelRoute('n', http.MethodGet, getNewZettelHandler)
	router.AddZettelRoute('n', http.MethodPost, postNewZettelHandler)
	router.AddListRoute('r', http.MethodGet, adapter.MakeListRoleHandler(te, usecase.NewListRole(s)))
	router.AddZettelRoute('r', http.MethodGet, adapter.MakeGetRenameZettelHandler(te, ucGetMeta))
	router.AddZettelRoute('r', http.MethodPost, adapter.MakePostRenameZettelHandler(usecase.NewRenameZettel(s)))
	router.AddListRoute('t', http.MethodGet, adapter.MakeListTagsHandler(te, usecase.NewListTags(s)))
	router.AddListRoute('s', http.MethodGet, adapter.MakeSearchHandler(te, p, usecase.NewSearch(s)))
	router.AddListRoute('z', http.MethodGet, adapter.MakeListMetaHandler('z', te, p, usecase.NewListMeta(s)))
	router.AddZettelRoute('z', http.MethodGet, adapter.MakeGetZettelHandler('z', te, p, ucGetZettel, ucGetMeta))
	return router
}

func main() {
	config.SetupVersion(releaseVersion, buildVersion)

	var port uint64
	var dir string

	flag.Uint64Var(&port, "p", 23123, "port number")
	flag.StringVar(&dir, "d", "./zettel", "zettel directory")
	flag.Parse()

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		log.Fatalf("Unable to create zettel directory %q: %v", dir, err)
	}
	fs, err := filestore.NewStore(dir)
	if err != nil {
		log.Fatalf("Unable to create filestore for %q: %v", dir, err)
	}
	cs := chainstore.NewStore(fs, gostore.NewStore())
	if err = cs.Start(context.Background()); err != nil {
		log.Fatalf("Unable to start zettel store: %v", err)
	}
	config.SetupConfiguration(cs)

	router := setupRouting(cs)

	v := config.Config.GetVersion()
	log.Printf("Release %v, Build %v", v.Release, v.Build)
	log.Printf("Listening on port %v", port)
	log.Printf("Zettel location %q", cs.Location())
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), router))
}
