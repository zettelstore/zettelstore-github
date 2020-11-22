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
	"net/url"
	"os"

	"zettelstore.de/z/auth/policy"
	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
	"zettelstore.de/z/place"
	"zettelstore.de/z/usecase"
	"zettelstore.de/z/web/adapter"
	"zettelstore.de/z/web/router"
	"zettelstore.de/z/web/session"
)

// ---------- Subcommand: run ------------------------------------------------

func runFunc(cfg *domain.Meta) (int, error) {
	p, exitCode, err := setupPlaces(cfg)
	if p == nil {
		return exitCode, err
	}
	readonly := cfg.GetBool("readonly")
	router := setupRouting(p, readonly)

	listenAddr, _ := cfg.Get("listen-addr")
	v := config.GetVersion()
	log.Printf("%v %v (%v@%v/%v)", v.Prog, v.Build, v.GoVersion, v.Os, v.Arch)
	if cfg.GetBool("verbose") {
		log.Println("Configuration")
		cfg.Write(os.Stderr)
	} else {
		log.Printf("Listening on %v", listenAddr)
		log.Printf("Zettel location [%v]", fullLocation(p))
		if readonly {
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
	return p, 0, nil
}

func getPlaceURIs(cfg *domain.Meta) []string {
	readonly := cfg.GetBool("readonly")
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
		if readonly {
			if u, err := url.Parse(uri); err == nil {
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
		return nil, nil
	}
	next, err := connectPlaces(placeURIs[1:])
	if err != nil {
		return nil, err
	}
	p, err := place.Connect(placeURIs[0], next)
	return p, err
}

func wrapPolicyPlace(p place.Place, pol policy.Policy) place.Place {
	if n := p.Next(); n != nil {
		return policy.NewPlace(p, pol, wrapPolicyPlace(n, pol))
	}
	return policy.NewPlace(p, pol, nil)
}
func setupRouting(up place.Place, readonly bool) http.Handler {
	pp := up
	var pol policy.Policy
	if config.WithAuth() || readonly {
		pol = policy.NewPolicy("default")
		pp = wrapPolicyPlace(up, pol)
	} else {
		pol = policy.NewPolicy("all")
	}
	te := adapter.NewTemplateEngine(up, pol)

	ucGetMeta := usecase.NewGetMeta(pp)
	ucGetZettel := usecase.NewGetZettel(pp)
	ucParseZettel := usecase.NewParseZettel(ucGetZettel)
	ucListMeta := usecase.NewListMeta(pp)
	ucListRoles := usecase.NewListRole(pp)
	ucListTags := usecase.NewListTags(pp)
	listHTMLMetaHandler := adapter.MakeListHTMLMetaHandler(te, ucListMeta)
	getHTMLZettelHandler := adapter.MakeGetHTMLZettelHandler(te, ucParseZettel, ucGetMeta)

	router := router.NewRouter()
	router.Handle("/", adapter.MakeGetRootHandler(pp, listHTMLMetaHandler, getHTMLZettelHandler))
	router.AddListRoute('a', http.MethodGet, adapter.MakeGetLoginHandler(te))
	router.AddListRoute('a', http.MethodPost, adapter.MakePostLoginHandler(te, usecase.NewAuthenticate(up)))
	router.AddListRoute('a', http.MethodPut, adapter.MakeRenewAuthHandler())
	router.AddZettelRoute('a', http.MethodGet, adapter.MakeGetLogoutHandler())
	router.AddListRoute('c', http.MethodGet, adapter.MakeReloadHandler(usecase.NewReload(pp)))
	if !readonly {
		router.AddZettelRoute('c', http.MethodGet, adapter.MakeGetCloneZettelHandler(te, ucGetZettel, usecase.NewCloneZettel()))
		router.AddZettelRoute('c', http.MethodPost, adapter.MakePostCreateZettelHandler(usecase.NewCreateZettel(pp)))
		router.AddZettelRoute('d', http.MethodGet, adapter.MakeGetDeleteZettelHandler(te, ucGetZettel))
		router.AddZettelRoute('d', http.MethodPost, adapter.MakePostDeleteZettelHandler(usecase.NewDeleteZettel(pp)))
		router.AddZettelRoute('e', http.MethodGet, adapter.MakeEditGetZettelHandler(te, ucGetZettel))
		router.AddZettelRoute('e', http.MethodPost, adapter.MakeEditSetZettelHandler(usecase.NewUpdateZettel(pp)))
	}
	router.AddListRoute('h', http.MethodGet, listHTMLMetaHandler)
	router.AddZettelRoute('h', http.MethodGet, getHTMLZettelHandler)
	router.AddZettelRoute('i', http.MethodGet, adapter.MakeGetInfoHandler(te, ucParseZettel, ucGetMeta))
	router.AddZettelRoute('k', http.MethodGet, adapter.MakeWebUIListsHandler(te, ucListMeta, ucListRoles, ucListTags))
	router.AddZettelRoute('l', http.MethodGet, adapter.MakeGetLinksHandler(ucParseZettel))
	if !readonly {
		router.AddZettelRoute('n', http.MethodGet, adapter.MakeGetNewZettelHandler(te, ucGetZettel, usecase.NewNewZettel()))
		router.AddZettelRoute('n', http.MethodPost, adapter.MakePostCreateZettelHandler(usecase.NewCreateZettel(pp)))
	}
	router.AddListRoute('r', http.MethodGet, adapter.MakeListRoleHandler(te, ucListRoles))
	if !readonly {
		router.AddZettelRoute('r', http.MethodGet, adapter.MakeGetRenameZettelHandler(te, ucGetMeta))
		router.AddZettelRoute('r', http.MethodPost, adapter.MakePostRenameZettelHandler(usecase.NewRenameZettel(pp)))
	}
	router.AddListRoute('t', http.MethodGet, adapter.MakeListTagsHandler(te, ucListTags))
	router.AddListRoute('s', http.MethodGet, adapter.MakeSearchHandler(te, usecase.NewSearch(pp), ucGetMeta, ucGetZettel))
	router.AddListRoute('z', http.MethodGet, adapter.MakeListMetaHandler(te, usecase.NewListMeta(pp), ucGetMeta, ucParseZettel))
	router.AddZettelRoute('z', http.MethodGet, adapter.MakeGetZettelHandler(ucParseZettel, ucGetMeta))
	return session.NewHandler(router, usecase.NewGetUserByZid(up))
}

func fullLocation(p place.Place) string {
	if n := p.Next(); n != nil {
		return p.Location() + ", " + fullLocation(n)
	}
	return p.Location()
}
