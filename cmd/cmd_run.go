//-----------------------------------------------------------------------------
// Copyright (c) 2020 Detlef Stern
//
// This file is part of zettelstore.
//
// Zettelstore is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"zettelstore.de/z/auth/policy"
	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
	"zettelstore.de/z/place"
	"zettelstore.de/z/place/progplace"
	"zettelstore.de/z/usecase"
	"zettelstore.de/z/web/adapter"
	"zettelstore.de/z/web/adapter/api"
	"zettelstore.de/z/web/adapter/webui"
	"zettelstore.de/z/web/router"
	"zettelstore.de/z/web/session"
)

// ---------- Subcommand: run ------------------------------------------------

func runFunc(cfg *domain.Meta) (int, error) {
	p, exitCode, err := setupPlaces(cfg)
	if p == nil {
		return exitCode, err
	}
	readonlyMode := config.IsReadOnlyMode()
	router := setupRouting(p, readonlyMode)

	listenAddr, _ := cfg.Get(config.StartupKeyListenAddress)
	v := config.GetVersion()
	log.Printf("%v %v (%v@%v/%v)", v.Prog, v.Build, v.GoVersion, v.Os, v.Arch)
	if cfg.GetBool(config.StartupKeyVerbose) {
		log.Println("Configuration")
		cfg.Write(os.Stderr)
	} else {
		log.Printf("Listening on %v", listenAddr)
		log.Printf("Zettel location [%v]", fullLocation(p))
		if readonlyMode {
			log.Println("Read-only mode")
		}
	}
	log.Fatal(http.ListenAndServe(listenAddr, router))
	return 0, nil
}

func setupPlaces(cfg *domain.Meta) (place.Place, int, error) {
	p, err := connectPlaces(getPlaceURIs(cfg))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to connect to specified places")
		return nil, 2, err
	}
	if err := p.Start(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, "Unable to start zettel place")
		return nil, 2, err
	}
	config.SetupConfiguration(p)
	progplace.Setup(cfg, p)
	return p, 0, nil
}

func getPlaceURIs(cfg *domain.Meta) []string {
	readonlyMode := config.IsReadOnlyMode()
	hasConst := false
	var result []string = nil
	for cnt := 1; ; cnt++ {
		key := fmt.Sprintf("place-%v-uri", cnt)
		uri, ok := cfg.Get(key)
		if !ok || uri == "" {
			break
		}
		if uri == "const:" {
			hasConst = true
		}
		if readonlyMode {
			if u, err := url.Parse(uri); err == nil {
				// TODO: the following is wrong under some circumstances:
				// 1. query parameter "readonly" is already set
				// 2. fragment is set
				if len(u.Query()) == 0 {
					uri += "?readonly"
				} else {
					uri += "&readonly"
				}
			}
		}
		result = append(result, uri)
	}
	if !hasConst {
		result = append(result, "const:")
	}
	return result
}

func connectPlaces(placeURIs []string) (place.Place, error) {
	if len(placeURIs) == 0 {
		return progplace.Get(), nil
	}
	next, err := connectPlaces(placeURIs[1:])
	if err != nil {
		return nil, err
	}
	p, err := place.Connect(placeURIs[0], next)
	return p, err
}

func setupRouting(up place.Place, readonlyMode bool) http.Handler {
	pp, pol := policy.PlaceWithPolicy(
		up, config.WithAuth, readonlyMode, config.GetExpertMode,
		config.IsOwner, config.GetVisibility)
	te := webui.NewTemplateEngine(up, pol)

	ucAuthenticate := usecase.NewAuthenticate(up)
	ucGetMeta := usecase.NewGetMeta(pp)
	ucGetZettel := usecase.NewGetZettel(pp)
	ucParseZettel := usecase.NewParseZettel(ucGetZettel)
	ucListMeta := usecase.NewListMeta(pp)
	ucListRoles := usecase.NewListRole(pp)
	ucListTags := usecase.NewListTags(pp)
	listHTMLMetaHandler := webui.MakeListHTMLMetaHandler(te, ucListMeta)
	getHTMLZettelHandler := webui.MakeGetHTMLZettelHandler(te, ucParseZettel, ucGetMeta)

	router := router.NewRouter()
	router.Handle("/", webui.MakeGetRootHandler(pp, listHTMLMetaHandler, getHTMLZettelHandler))
	router.AddListRoute('a', http.MethodGet, webui.MakeGetLoginHandler(te))
	router.AddListRoute('a', http.MethodPost, adapter.MakePostLoginHandler(
		api.MakePostLoginHandlerAPI(ucAuthenticate),
		webui.MakePostLoginHandlerHTML(te, ucAuthenticate)))
	router.AddListRoute('a', http.MethodPut, api.MakeRenewAuthHandler())
	router.AddZettelRoute('a', http.MethodGet, webui.MakeGetLogoutHandler())
	router.AddListRoute('c', http.MethodGet, adapter.MakeReloadHandler(
		usecase.NewReload(pp),
		api.ReloadHandlerAPI,
		webui.ReloadHandlerHTML))
	if !readonlyMode {
		router.AddZettelRoute('c', http.MethodGet, webui.MakeGetCloneZettelHandler(te, ucGetZettel, usecase.NewCloneZettel()))
		router.AddZettelRoute('c', http.MethodPost, webui.MakePostCreateZettelHandler(usecase.NewCreateZettel(pp)))
		router.AddZettelRoute('d', http.MethodGet, webui.MakeGetDeleteZettelHandler(te, ucGetZettel))
		router.AddZettelRoute('d', http.MethodPost, webui.MakePostDeleteZettelHandler(usecase.NewDeleteZettel(pp)))
		router.AddZettelRoute('e', http.MethodGet, webui.MakeEditGetZettelHandler(te, ucGetZettel))
		router.AddZettelRoute('e', http.MethodPost, webui.MakeEditSetZettelHandler(usecase.NewUpdateZettel(pp)))
	}
	router.AddListRoute('h', http.MethodGet, listHTMLMetaHandler)
	router.AddZettelRoute('h', http.MethodGet, getHTMLZettelHandler)
	router.AddZettelRoute('i', http.MethodGet, webui.MakeGetInfoHandler(te, ucParseZettel, ucGetMeta))
	router.AddZettelRoute('k', http.MethodGet, webui.MakeWebUIListsHandler(te, ucListMeta, ucListRoles, ucListTags))
	router.AddZettelRoute('l', http.MethodGet, api.MakeGetLinksHandler(ucParseZettel))
	if !readonlyMode {
		router.AddZettelRoute('n', http.MethodGet, webui.MakeGetNewZettelHandler(te, ucGetZettel, usecase.NewNewZettel()))
		router.AddZettelRoute('n', http.MethodPost, webui.MakePostCreateZettelHandler(usecase.NewCreateZettel(pp)))
	}
	router.AddListRoute('r', http.MethodGet, api.MakeListRoleHandler(ucListRoles))
	if !readonlyMode {
		router.AddZettelRoute('r', http.MethodGet, webui.MakeGetRenameZettelHandler(te, ucGetMeta))
		router.AddZettelRoute('r', http.MethodPost, webui.MakePostRenameZettelHandler(usecase.NewRenameZettel(pp)))
	}
	router.AddListRoute('t', http.MethodGet, api.MakeListTagsHandler(ucListTags))
	router.AddListRoute('s', http.MethodGet, webui.MakeSearchHandler(te, usecase.NewSearch(pp), ucGetMeta, ucGetZettel))
	router.AddListRoute('z', http.MethodGet, api.MakeListMetaHandler(usecase.NewListMeta(pp), ucGetMeta, ucParseZettel))
	router.AddZettelRoute('z', http.MethodGet, api.MakeGetZettelHandler(ucParseZettel, ucGetMeta))
	return session.NewHandler(router, usecase.NewGetUserByZid(up))
}

func fullLocation(p place.Place) string {
	if n := p.Next(); n != nil {
		if rest := fullLocation(n); len(rest) > 0 {
			return p.Location() + ", " + fullLocation(n)
		}
		return p.Location()
	}
	return p.Location()
}
