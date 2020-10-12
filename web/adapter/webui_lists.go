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

// Package adapter provides handlers for web requests.
package adapter

import (
	"log"
	"net/http"
	"sort"
	"strconv"

	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
	"zettelstore.de/z/usecase"
	"zettelstore.de/z/web/session"
)

// MakeListHTMLMetaHandler creates a HTTP handler for rendering the list of zettel as HTML.
func MakeListHTMLMetaHandler(te *TemplateEngine, listMeta usecase.ListMeta) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		renderWebUIZettelList(w, r, te, listMeta)
	}
}

type tagInfo struct {
	Name  string
	Count int
	Size  int
}

var fontSizes = [...]int{75, 83, 100, 117, 150, 200}

// MakeWebUIListsHandler creates a new HTTP handler for the use case "list some zettel".
func MakeWebUIListsHandler(te *TemplateEngine, listMeta usecase.ListMeta, listRole usecase.ListRole, listTags usecase.ListTags) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		zid, err := domain.ParseZettelID(r.URL.Path[1:])
		if err != nil {
			http.NotFound(w, r)
			return
		}
		switch zid.Format() {
		case "00000000000001":
			renderWebUIZettelList(w, r, te, listMeta)
		case "00000000000002":
			renderWebUIRolesList(w, r, te, listRole)
		case "00000000000003":
			renderWebUITagsList(w, r, te, listTags)
		}
	}
}

func renderWebUIZettelList(w http.ResponseWriter, r *http.Request, te *TemplateEngine, listMeta usecase.ListMeta) {
	ctx := r.Context()
	filter, sorter := getFilterSorter(r)
	metaList, err := listMeta.Run(ctx, filter, sorter)
	if err != nil {
		checkUsecaseError(w, err)
		return
	}

	user := session.GetUser(ctx)
	metas, err := buildHTMLMetaList(metaList)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	te.renderTemplate(r.Context(), w, domain.ListTemplateID, struct {
		Lang  string
		Title string
		User  userWrapper
		Metas []metaInfo
	}{
		Lang:  config.GetDefaultLang(),
		Title: config.GetSiteName(),
		User:  wrapUser(user),
		Metas: metas,
	})
}

func renderWebUIRolesList(w http.ResponseWriter, r *http.Request, te *TemplateEngine, listRole usecase.ListRole) {
	ctx := r.Context()
	roleList, err := listRole.Run(ctx)
	if err != nil {
		checkUsecaseError(w, err)
		return
	}

	user := session.GetUser(ctx)
	te.renderTemplate(ctx, w, domain.RolesTemplateID, struct {
		Lang  string
		Title string
		User  userWrapper
		Roles []string
	}{
		Lang:  config.GetDefaultLang(),
		Title: config.GetSiteName(),
		User:  wrapUser(user),
		Roles: roleList,
	})
}

func renderWebUITagsList(w http.ResponseWriter, r *http.Request, te *TemplateEngine, listTags usecase.ListTags) {
	ctx := r.Context()
	iMinCount, _ := strconv.Atoi(r.URL.Query().Get("min"))
	tagData, err := listTags.Run(ctx, iMinCount)
	if err != nil {
		checkUsecaseError(w, err)
		return
	}

	user := session.GetUser(ctx)
	tagsList := make([]tagInfo, 0, len(tagData))
	countMap := make(map[int]int)
	for tag, ml := range tagData {
		count := len(ml)
		countMap[count]++
		tagsList = append(tagsList, tagInfo{tag, count, 100})
	}
	sort.Slice(tagsList, func(i, j int) bool { return tagsList[i].Name < tagsList[j].Name })

	countList := make([]int, 0, len(countMap))
	for count := range countMap {
		countList = append(countList, count)
	}
	sort.Ints(countList)
	for pos, count := range countList {
		countMap[count] = fontSizes[(pos*len(fontSizes))/len(countList)]
	}
	for i := 0; i < len(tagsList); i++ {
		tagsList[i].Size = countMap[tagsList[i].Count]
	}

	te.renderTemplate(ctx, w, domain.TagsTemplateID, struct {
		Lang   string
		Title  string
		User   userWrapper
		Tags   []tagInfo
		Counts []int
	}{
		Lang:   config.GetDefaultLang(),
		Title:  config.GetSiteName(),
		User:   wrapUser(user),
		Tags:   tagsList,
		Counts: countList,
	})
}
