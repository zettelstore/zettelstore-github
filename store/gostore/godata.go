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

// Package gostore stores zettel inside the executable.
package gostore

import (
	"fmt"

	"zettelstore.de/z/domain"
)

const (
	syntaxTemplate    = "template"
	roleConfiguration = "configuration"
)

var goData = goStore{
	name: "globals://",
	zettel: map[domain.ZettelID]goZettel{
		domain.ConfigurationID: goZettel{
			goHeader{
				domain.MetaKeyTitle:  "Configuration",
				domain.MetaKeySyntax: "meta",
				domain.MetaKeyRole:   roleConfiguration,
			},
			"# Zettelstore Configuration",
		},

		domain.BaseTemplateID: goZettel{
			goHeader{
				domain.MetaKeyTitle:  "Base HTML Template",
				domain.MetaKeySyntax: syntaxTemplate,
				domain.MetaKeyRole:   roleConfiguration,
			},
			domain.NewContent(fmt.Sprintf(
				`<!DOCTYPE html>
<html{{if .Lang}} lang="{{.Lang}}"{{end}}>
<head>
<meta charset="utf-8">
<meta name="referrer" content="same-origin">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<meta name="generator" content="Zettelstore {{config.GetVersion.Release}}, build {{config.GetVersion.Build}}">
{{- block "meta-header" .}}{{end}}
<link rel="stylesheet" href="{{url 'c' %q}}">
{{- block "header" .}}{{end}}
<title>{{.Title}}</title>
</head>
<body>
<nav class="zs-menu">
<a href="{{url '/' ""}}">Home</a>
<div class="zs-dropdown">
<button>Lists</button>
<nav class="zs-dropdown-content">
<a href="{{url 'h' ""}}">List Zettel</a>
<a href="{{url 'r' ""}}">List Roles</a>
<a href="{{url 't' ""}}">List Tags</a>
<a href="{{url 'c' ""}}">Reload</a>
</nav>
</div>
{{if not config.IsReadOnly}}<a href="{{url 'n' ""}}">New</a>{{end}}
{{- block "menu" .}}{{end -}}
<form action="{{url 's' ""}}">
<input type="text" placeholder="Search.." name="s">
</form>
</nav>
<main class="content">
{{- block "content" .}}TODO{{end -}}
</main>
</body>
</html>`,
				domain.BaseCSSID,
			)),
		},

		domain.LoginTemplateID: goZettel{
			goHeader{
				domain.MetaKeyTitle:  "Login Form HTML Template",
				domain.MetaKeySyntax: syntaxTemplate,
				domain.MetaKeyRole:   roleConfiguration,
			},
			domain.NewContent(
				`{{define "content"}}
<article>
<header>
<h1>{{.Title}}</h1>
</header>
<form method="POST">
<div>
<label for="username">User name</label>
<input class="zs-input" type="text" id="username" name="username" placeholder="Your user name..">
</div>
<div>
<label for="password">Password</label>
<input class="zs-input" type="password" id="password" name="password" placeholder="Your password..">
</div>
<input class="zs-button" type="submit" value="Login">
</form>
</article>
{{end}}`,
			)},

		domain.ListTemplateID: goZettel{
			goHeader{
				domain.MetaKeyTitle:  "List Meta HTML Template",
				domain.MetaKeySyntax: syntaxTemplate,
				domain.MetaKeyRole:   roleConfiguration,
			},
			domain.NewContent(fmt.Sprintf(
				`{{define "content"}}
<h1>{{.Title}}</h1>
<ul>
{{range .Metas}}<li><a href="{{url $.Key .Meta.ID}}">{{.Title}}</a><span class="zs-meta">{{range .Meta.GetListOrNil %q}} <a href="{{url $.Key ""}}?tags={{.}}">{{.}}</a>{{end}}</span></li>{{end}}
</ul>
<p>Items: {{len .Metas}}</p>
{{end}}`,
				domain.MetaKeyTags))},

		domain.DetailTemplateID: goZettel{
			goHeader{
				domain.MetaKeyTitle:  "Detail HTML Template",
				domain.MetaKeySyntax: syntaxTemplate,
				domain.MetaKeyRole:   roleConfiguration,
			},
			domain.NewContent(fmt.Sprintf(
				`{{define "meta-header"}}
{{- .MetaHeader}}
{{- end}}
{{define "content"}}
<article>
<header>
<h1>{{.HTMLTitle}}</h1>
<div class="zs-meta">
{{if not config.IsReadOnly}}<a href="{{url 'e' .Meta.ID}}">Edit</a> &#183;
{{ .Meta.ID}} &#183;{{end}}
<a href="{{url 'i' .Meta.ID}}">Info</a> &#183;
{{- with .Meta.GetDefault "role" "*"}} (<a href="{{url $.Key ""}}?role={{.}}">{{.}}</a>){{end}}
{{- with .Meta.GetListOrNil %q}}
{{- if .}}:{{range .}} <a href="{{url $.Key ""}}?tags={{.}}">{{.}}</a>{{end}}{{end}}
{{- end}}
{{if not config.IsReadOnly}}&#183; <a href="{{url 'n' .Meta.ID}}">Clone</a>{{end}}
{{with .Meta.GetDefault "url" ""}}{{if .}}<br>URL: <a href="{{.}}" target="_blank">{{.}}</a>{{HTML config.GetIconMaterial}}{{end}}{{end}}
</div>
</header>
{{- .Content -}}
</article>
{{- end}}`,
				domain.MetaKeyTags)),
		},

		domain.InfoTemplateID: goZettel{
			goHeader{
				domain.MetaKeyTitle:  "Info HTML Template",
				domain.MetaKeySyntax: syntaxTemplate,
				domain.MetaKeyRole:   roleConfiguration,
			},
			domain.NewContent(
				`{{define "content"}}
<article>
<header>
<h1>Information for Zettel {{.Meta.ID}}</h1>
<div class="zs-meta">
<a href="{{url 'h' $.Meta.ID}}">Web</a>{{range $f := .Formats}} &#183; <a href="{{url 'z' $.Meta.ID}}?_format={{$f}}">{{$f}}</a>{{end}}
</div>
{{if not config.IsReadOnly}}<a href="{{url 'e' .Meta.ID}}">Edit</a> &#183;
<a href="{{url 'n' .Meta.ID}}">Clone</a> &#183;{{end}}
<a href="{{url 'm' .Meta.ID}}">Meta</a> &#183;
<a href="{{url 'c' .Meta.ID}}">Content</a> {{if not config.IsReadOnly}}&#183;
<a href="{{url 'r' .Meta.ID}}">Rename</a> &#183;
<a href="{{url 'd' .Meta.ID}}">Delete</a>{{end}}
</header>
<h2>Interpreted Meta Data</h2>
<table>
{{- range .Meta.Pairs}}
<tr><td>{{.Key}}</td><td>{{htmlMetaValue $.Meta .Key}}</td></tr>
{{- end -}}
</table>
{{if or .IntLinks .ExtLinks}}
<h2>Outgoing Links</h2>
{{if .IntLinks}}
<h3>Internal</h3>
<ul>
{{range .IntLinks}}<li>{{if .Found}}<a href="{{url 'h' .ID}}">{{.Title}}</a>{{else}}{{.ID}}{{end}}</li>{{end}}
</ul>
{{end}}
{{if .ExtLinks}}
<h3>External</h3>
<ul>
{{range .ExtLinks}}<li><a href="{{.}}" target="_blank">{{.}}</a>{{HTML config.GetIconMaterial}}</li>{{end}}
</ul>
{{end}}
{{end}}
</article>
{{- end}}`),
		},

		domain.FormTemplateID: goZettel{
			goHeader{
				domain.MetaKeyTitle:  "Form HTML Template",
				domain.MetaKeySyntax: syntaxTemplate,
				domain.MetaKeyRole:   roleConfiguration,
			},
			`{{define "content"}}
<article>
<header>
<h1>{{.Title}}</h1>
</header>
<form method="POST">
<div>
<label for="title">Title</label>
<input class="zs-input" type="text" id="title" name="title" placeholder="Title.." value="{{.Meta.GetDefault "title" ""}}">
</div>
<div>
<label for="tags">Tags</label>
<input class="zs-input" type="text" id="tags" name="tags" placeholder="#tag" value="{{.Meta.GetDefault "tags" ""}}">
</div>
<div>
<label for="role">Role</label>
<input class="zs-input" type="text" id="role" name="role" placeholder="role.." value="{{.Meta.GetDefault "role" ""}}">
</div>
<div>
<label for="syntax">Syntax</label>
<input class="zs-input" type="text" id="syntax" name="syntax" placeholder="syntax.." value="{{.Meta.GetDefault "syntax" ""}}">
</div>
<div>
<label for="meta">Meta</label>
<textarea class="zs-input" id="meta" name="meta" rows="4" placeholder="key: value">
{{- range .Meta.PairsRest}}
{{.Key}}: {{.Value}}
{{- end -}}
</textarea>
</div>
<div>
<label for="content">Content</label>
<textarea class="zs-input zs-content" id="meta" name="content" rows="20" placeholder="Your content..">
{{- .Content -}}
</textarea>
</div>
<input class="zs-button" type="submit" value="Submit">
</form>
</article>
{{end}}`,
		},

		domain.RenameTemplateID: goZettel{
			goHeader{
				domain.MetaKeyTitle:  "Rename Form HTML Template",
				domain.MetaKeySyntax: syntaxTemplate,
				domain.MetaKeyRole:   roleConfiguration,
			},
			`{{define "content"}}
<article>
<header>
<h1>Rename Zettel {{.Meta.ID}}</h1>
</header>
<p>Do you really want to rename this zettel?</p>
<form method="POST">
<div>
<label for="newid">New ID</label>
<input class="zs-input" type="text" id="newid" name="newid" placeholder="ID.." value="{{.Meta.ID}}">
</div>
<input type="hidden" id="curid" name="curid" value="{{.Meta.ID}}">
<input class="zs-button" type="submit" value="Rename">
</form>
<dl>
{{- range .Meta.Pairs}}
<dt>{{.Key}}:</dt><dd>{{.Value}}</dd>
{{- end -}}
</dl>
</article>
{{end}}`,
		},

		domain.DeleteTemplateID: goZettel{
			goHeader{
				domain.MetaKeyTitle:  "Delete HTML Template",
				domain.MetaKeySyntax: syntaxTemplate,
				domain.MetaKeyRole:   roleConfiguration,
			},
			`{{define "content"}}
<article>
<header>
<h1>Delete Zettel {{.Meta.ID}}</h1>
</header>
<p>Do you really want to delete this zettel?</p>
<dl>
{{- range .Meta.Pairs}}
<dt>{{.Key}}:</dt><dd>{{.Value}}</dd>
{{- end -}}
</dl>
<form method="POST">
<input class="zs-button" type="submit" value="Delete">
</form>
</article>
{{end}}`,
		},

		domain.RolesTemplateID: goZettel{
			goHeader{
				domain.MetaKeyTitle:  "List Roles HTML Template",
				domain.MetaKeySyntax: syntaxTemplate,
				domain.MetaKeyRole:   roleConfiguration,
			},
			`{{define "content"}}
<h1>Currently used roles</h1>
<ul>
{{range .Roles}}<li><a href="{{url 'h' ""}}?role={{.}}">{{.}}</a></li>{{end}}
</ul>
{{end}}`,
		},

		domain.TagsTemplateID: goZettel{
			goHeader{
				domain.MetaKeyTitle:  "List Tags HTML Template",
				domain.MetaKeySyntax: syntaxTemplate,
				domain.MetaKeyRole:   roleConfiguration,
			},
			`{{define "content"}}
<h1>Currently used tags</h1>
<div class="zs-meta">
<a href="{{url 't' ""}}">All</a>{{range .Counts}}, <a href="{{url 't' ""}}?min={{.}}">{{.}}</a>{{end}}
</div>
{{range .Tags}} <a href="{{url 'h' ""}}?tags={{.Name}}" style="font-size:{{.Size}}%">{{.Name}}</a><sup>{{.Count}}</sup>{{end}}
{{end}}`,
		},

		domain.BaseCSSID: goZettel{
			goHeader{
				domain.MetaKeyTitle:  "Base CSS",
				domain.MetaKeySyntax: "css",
				domain.MetaKeyRole:   roleConfiguration,
			},
			`/* Default CSS */
*,*::before,*::after {
  box-sizing: border-box;
}
html {
  font-size: 1rem;
  font-family: serif;
  scroll-behavior: smooth;
  height: 100%;
}
body {
  margin: 0;
  min-height: 100vh;
  text-rendering: optimizeSpeed;
  line-height: 1.6;
  overflow-x: hidden;
  background-color: #f8f8f8 ;
  height: 100%;
}
nav.zs-menu {
  background-color: hsl(210, 28%, 90%);
  overflow: auto;
  white-space: nowrap;
  font-family: sans-serif;
}

nav.zs-menu > a {
  float:left;
  display: inline-block;
  text-align: center;
  padding:.83rem 1rem;
  text-decoration: none;
  color:black;
}
nav.zs-menu > a:hover, .zs-dropdown:hover button {
  background-color: hsl(210, 28%, 80%);
}
nav.zs-menu form {
  float: right;
}
nav.zs-menu form input[type=text] {
  padding: .25rem;
  border: none;
  margin-top: .5rem;
  margin-right: 1rem;
}
.zs-dropdown {
  float: left;
  overflow: hidden;
}
.zs-dropdown > button {
  font-size: 16px;
  border: none;
  outline: none;
  color: black;
  padding:.83rem 1rem;
  background-color: inherit;
  font-family: inherit;
  margin: 0;
}
.zs-dropdown-content {
  display: none;
  position: absolute;
  background-color: #f9f9f9;
  min-width: 160px;
  box-shadow: 0px 8px 16px 0px rgba(0,0,0,0.2);
  z-index: 1;
}
.zs-dropdown-content > a {
  float: none;
  color: black;
  padding:.83rem 1rem;
  text-decoration: none;
  display: block;
  text-align: left;
}
.zs-dropdown-content > a:hover {
  background-color: hsl(210, 28%, 75%);
}
.zs-dropdown:hover > .zs-dropdown-content {
  display: block;
}
main {
  padding: 0 1rem;
}
article > * + * {
  margin-top: 1rem;
}
article {
  padding: 0;
  margin: 0;
}
article header {
  padding: 0;
  margin: 0;
}
h1 { font-size:2rem;    margin:.67rem 0 }
h2 { font-size:1.5rem;  margin:.75rem 0 }
h3 { font-size:1.17rem; margin:.83rem 0 }
h4 { font-size:1rem;    margin:1.12rem 0 }
h5 { font-size:.83rem;  margin:1.5rem 0 }
h6 { font-size:.75rem;  margin:1.67rem 0 }
p {
  margin: 1rem 0 0 0;
}
li,figure,figcaption,dl {
  margin: 0;
}
dt {
  margin: 1rem 0 0 0;
}
dt+dd {
  margin-top: 0;
}
dd {
  margin: 1rem 0 0 2rem;
}
dd > p:first-child {
  margin: 0 0 0 0;
}
blockquote {
  border-left: 0.5rem solid lightgray;
  padding-left: 1rem;
  margin-left: 1rem;
  margin-right: 2rem;
  font-style: italic;
}
blockquote p {
  margin-bottom: 1rem;
}
blockquote cite {
  font-style: normal;
}
table {
  border-collapse: collapse;
  border-spacing: 0;
  max-width: 100%;
}
th,td {
  text-align: left;
  padding: 0.5rem;
}
td { border-bottom: 1px solid hsl(0, 0%, 85%); }
thead th { border-bottom: 2px solid hsl(0, 0%, 70%); }
tfoot th { border-top: 2px solid hsl(0, 0%, 70%); }
main form {
  padding: 0 .5em;
  margin: .5em 0 0 0;
}
main form:after {
  content: ".";
  display: block;
  height: 0;
  clear: both;
  visibility: hidden;
}
main form div {
  margin: .5em 0 0 0
}
input,button,select {
  font: inherit;
}
label {
  font-family: serif;
  font-weight: bold;
}
textarea {
  font-family: monospace;
  resize: vertical;
  width: 100%;
}
.zs-input {
  padding: .5em;
  display:block;
  border:none;
  border-bottom:1px solid #ccc;
  width:100%;
}
.zs-button {
  float:right;
  margin: .5em 0 .5em 1em;
}
a:not([class]) {
  text-decoration-skip-ink: auto;
}
.zs-broken {
  text-decoration: line-through;
}
.zs-text-icon {
  height:1rem;
  vertical-align:text-bottom;
}
img {
  max-width: 100%;
}
.zs-endnotes {
  padding-top: 1rem;
  border-top: 1px solid;
}
code,pre,kbd {
  font-family: monospace;
  font-size: 85%;
}
p code {
  padding: .1rem .2rem;
  background: #f0f0f0;
  border: 1px solid #ccc;
  border-radius: .25rem;
}
pre {
  padding: .5rem .7rem;
  max-width: 100%;
  overflow: auto;
  border: 1px solid #ccc;
  border-radius: .5rem;
  background: #f0f0f0;
}
pre code {
  font-size: 95%;
  position: relative;
}
div.zs-indication {
  padding: .5rem .7rem;
  max-width: 100%;
  border-radius: .5rem;
  border: 1px solid black;
}
div.zs-indication p:first-child {
  margin-top: 0;
}
span.zs-indication {
  border: 1px solid black;
  border-radius: .25rem;
  padding: .1rem .2rem;
  font-size: 95%;
}
.zs-example { border-style: dotted }
kbd {
  background: hsl(210, 5%, 100%);
  border: 1px solid hsl(210, 5%, 70%);
  border-radius: .25rem;
  padding: .1rem .2rem;
  font-size: 75%;
}
.zs-meta {
  font-size:.75rem;
  color:#888;
  margin-bottom:1rem;
}
.zs-meta a {
  color:#888;
}
h1+.zs-meta {
  margin-top:-1rem;
}
@media (prefers-reduced-motion: reduce) {
  * {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
    scroll-behavior: auto !important;
  }
}
`,
		},

		domain.MaterialIconID: goZettel{
			goHeader{
				domain.MetaKeyTitle:  "Text icon for external material",
				domain.MetaKeySyntax: "svg",
				domain.MetaKeyRole:   roleConfiguration,
				domain.MetaKeyURL:    "https://icons8.com/icon/43738/external-link",
			},
			`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16"><path d="M 9 2 L 9 3 L 12.292969 3 L 6.023438 9.273438 L 6.726563 9.976563 L 13 3.707031 L 13 7 L 14 7 L 14 2 Z M 4 4 C 2.894531 4 2 4.894531 2 6 L 2 12 C 2 13.105469 2.894531 14 4 14 L 10 14 C 11.105469 14 12 13.105469 12 12 L 12 7 L 11 8 L 11 12 C 11 12.550781 10.550781 13 10 13 L 4 13 C 3.449219 13 3 12.550781 3 12 L 3 6 C 3 5.449219 3.449219 5 4 5 L 8 5 L 9 4 Z"/></svg>`,
		},
	},
}
