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
	"net/http"
	"sort"
	"strconv"

	"zettelstore.de/z/config"
	"zettelstore.de/z/domain"
	"zettelstore.de/z/encoder"
	"zettelstore.de/z/encoder/jsonenc"
	"zettelstore.de/z/usecase"
	"zettelstore.de/z/web/session"
)

type tagInfo struct {
	Name  string
	Count int
	Size  int
}

var fontSizes = [...]int{75, 83, 100, 117, 150, 200}

// MakeListTagsHandler creates a new HTTP handler for the use case "list some zettel".
func MakeListTagsHandler(te *TemplateEngine, listTags usecase.ListTags) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		iMinCount, _ := strconv.Atoi(r.URL.Query().Get("min"))
		tagData, err := listTags.Run(ctx, iMinCount)
		if err != nil {
			checkUsecaseError(w, err)
			return
		}

		user := session.GetUser(ctx)
		if format := getFormat(r, encoder.GetDefaultFormat()); format != "html" {
			w.Header().Set("Content-Type", format2ContentType(format))
			switch format {
			case "json":
				renderListTagsJSON(w, tagData)
				return
			}
		}

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
}

func renderListTagsJSON(w http.ResponseWriter, tagData usecase.TagData) {
	buf := encoder.NewBufWriter(w)

	tagList := make([]string, 0, len(tagData))
	for tag := range tagData {
		tagList = append(tagList, tag)
	}
	sort.Strings(tagList)

	buf.WriteString("{\"tags\":{")
	first := true
	for _, tag := range tagList {
		if first {
			buf.WriteByte('"')
			first = false
		} else {
			buf.WriteString(",\"")
		}
		buf.Write(jsonenc.Escape(tag))
		buf.WriteString("\":[")
		for i, meta := range tagData[tag] {
			if i > 0 {
				buf.WriteByte(',')
			}
			buf.WriteByte('"')
			buf.WriteString(meta.Zid.Format())
			buf.WriteByte('"')
		}
		buf.WriteString("]")

	}
	buf.WriteString("}}")
	buf.Flush()
}
