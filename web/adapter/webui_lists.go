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

// MakeWebUIListsHandler creates a new HTTP handler for the use case "list some zettel".
func MakeWebUIListsHandler(te *TemplateEngine, listMeta usecase.ListMeta, listRole usecase.ListRole, listTags usecase.ListTags) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		zid, err := domain.ParseZettelID(r.URL.Path[1:])
		if err != nil {
			http.NotFound(w, r)
			return
		}
		switch zid {
		case 1:
			renderWebUIZettelList(w, r, te, listMeta)
		case 2:
			renderWebUIRolesList(w, r, te, listRole)
		case 3:
			renderWebUITagsList(w, r, te, listTags)
		}
	}
}

func renderWebUIZettelList(w http.ResponseWriter, r *http.Request, te *TemplateEngine, listMeta usecase.ListMeta) {
	ctx := r.Context()
	filter, sorter := getFilterSorter(r.URL.Query())
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
		baseData
		Metas []metaInfo
	}{
		baseData: te.makeBaseData(ctx, config.GetDefaultLang(), config.GetSiteName(), user),
		Metas:    metas,
	})
}

type roleInfo struct {
	Text string
	URL  string
}

func renderWebUIRolesList(w http.ResponseWriter, r *http.Request, te *TemplateEngine, listRole usecase.ListRole) {
	ctx := r.Context()
	roleList, err := listRole.Run(ctx)
	if err != nil {
		checkUsecaseError(w, err)
		return
	}

	roleInfos := make([]roleInfo, 0, len(roleList))
	for _, r := range roleList {
		roleInfos = append(roleInfos, roleInfo{r, newURLBuilder('h').AppendQuery("role", r).String()})
	}

	user := session.GetUser(ctx)
	te.renderTemplate(ctx, w, domain.RolesTemplateID, struct {
		baseData
		Roles []roleInfo
	}{
		baseData: te.makeBaseData(ctx, config.GetDefaultLang(), config.GetSiteName(), user),
		Roles:    roleInfos,
	})
}

type countInfo struct {
	Count string
	URL   string
}

type tagInfo struct {
	Name  string
	URL   string
	count int
	Count string
	Size  string
}

var fontSizes = [...]int{75, 83, 100, 117, 150, 200}

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
	baseTagListURL := newURLBuilder('h')
	for tag, ml := range tagData {
		count := len(ml)
		countMap[count]++
		tagsList = append(tagsList, tagInfo{tag, baseTagListURL.AppendQuery("tags", tag).String(), count, "", ""})
		baseTagListURL.ClearQuery()
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
		count := tagsList[i].count
		tagsList[i].Count = strconv.Itoa(count)
		tagsList[i].Size = strconv.Itoa(countMap[count])
	}

	base := te.makeBaseData(ctx, config.GetDefaultLang(), config.GetSiteName(), user)
	minCounts := make([]countInfo, 0, len(countList))
	for _, c := range countList {
		sCount := strconv.Itoa(c)
		minCounts = append(minCounts, countInfo{sCount, base.ListTagsURL + "?min=" + sCount})
	}

	te.renderTemplate(ctx, w, domain.TagsTemplateID, struct {
		baseData
		MinCounts []countInfo
		Tags      []tagInfo
	}{
		baseData:  base,
		MinCounts: minCounts,
		Tags:      tagsList,
	})
}
