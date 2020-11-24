//-----------------------------------------------------------------------------
// Copyright (c) 2020 Detlef Stern
//
// This file is part of zettelstore.
//
// Zettelstore is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

// Package adapter provides handlers for web requests.
package adapter

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"

	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
	"zettelstore.de/z/place"
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
	query := r.URL.Query()
	filter, sorter := getFilterSorter(query, false)
	ctx := r.Context()
	renderWebUIMetaList(
		ctx, w, te, sorter,
		func(sorter *place.Sorter) ([]*domain.Meta, error) {
			return listMeta.Run(ctx, filter, sorter)
		},
		func(offset int) string { return newPageURL('h', query, offset, "_offset", "_limit") })
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

// MakeSearchHandler creates a new HTTP handler for the use case "search".
func MakeSearchHandler(te *TemplateEngine, search usecase.Search, getMeta usecase.GetMeta, getZettel usecase.GetZettel) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		filter, sorter := getFilterSorter(query, true)
		if filter == nil || len(filter.Expr) == 0 {
			http.Redirect(w, r, newURLBuilder('h').String(), http.StatusFound)
			return
		}

		ctx := r.Context()
		renderWebUIMetaList(
			ctx, w, te, sorter,
			func(sorter *place.Sorter) ([]*domain.Meta, error) {
				return search.Run(ctx, filter, sorter)
			},
			func(offset int) string { return newPageURL('s', query, offset, "offset", "limit") })
	}
}

func renderWebUIMetaList(
	ctx context.Context, w http.ResponseWriter, te *TemplateEngine,
	sorter *place.Sorter,
	ucMetaList func(sorter *place.Sorter) ([]*domain.Meta, error),
	pageURL func(int) string) {

	var metaList []*domain.Meta
	var err error
	var prevURL, nextURL string
	if lps := config.GetListPageSize(); lps > 0 {
		sorter = ensureSorter(sorter)
		if sorter.Limit < lps {
			sorter.Limit = lps + 1
		}

		metaList, err = ucMetaList(sorter)
		if err != nil {
			checkUsecaseError(w, err)
			return
		}
		if offset := sorter.Offset; offset > 0 {
			offset -= lps
			if offset < 0 {
				offset = 0
			}
			prevURL = pageURL(offset)
		}
		if len(metaList) >= sorter.Limit {
			nextURL = pageURL(sorter.Offset + lps)
			metaList = metaList[:len(metaList)-1]
		}
	} else {
		metaList, err = ucMetaList(sorter)
		if err != nil {
			checkUsecaseError(w, err)
			return
		}
	}
	user := session.GetUser(ctx)
	metas, err := buildHTMLMetaList(metaList)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	te.renderTemplate(ctx, w, domain.ListTemplateID, struct {
		baseData
		Metas       []metaInfo
		HasPrevNext bool
		HasPrev     bool
		PrevURL     template.URL
		HasNext     bool
		NextURL     template.URL
	}{
		baseData:    te.makeBaseData(ctx, config.GetDefaultLang(), config.GetSiteName(), user),
		Metas:       metas,
		HasPrevNext: len(prevURL) > 0 || len(nextURL) > 0,
		HasPrev:     len(prevURL) > 0,
		PrevURL:     template.URL(prevURL),
		HasNext:     len(nextURL) > 0,
		NextURL:     template.URL(nextURL),
	})
}

func newPageURL(key byte, query url.Values, offset int, offsetKey, limitKey string) string {
	urlBuilder := newURLBuilder(key)
	for key, values := range query {
		if key != offsetKey && key != limitKey {
			for _, val := range values {
				urlBuilder.AppendQuery(key, val)
			}
		}
	}
	if offset > 0 {
		urlBuilder.AppendQuery(offsetKey, strconv.Itoa(offset))
	}
	return urlBuilder.String()
}
