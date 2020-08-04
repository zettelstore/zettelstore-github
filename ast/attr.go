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

// Package ast provides the abstract syntax tree.
package ast

import (
	"strings"
)

// Attributes store additional information about some node types.
type Attributes struct {
	Attrs map[string]string
}

// HasDefault returns true, if the default attribute "-" has been set.
func (a *Attributes) HasDefault() bool {
	if a != nil {
		_, ok := a.Attrs["-"]
		return ok
	}
	return false
}

// RemoveDefault removes the default attribute
func (a *Attributes) RemoveDefault() {
	if a != nil {
		delete(a.Attrs, "-")
	}
}

// Get returns the attribute value of the given key and a succes value.
func (a *Attributes) Get(key string) (string, bool) {
	if a != nil {
		value, ok := a.Attrs[key]
		return value, ok
	}
	return "", false
}

// Clone returns a duplicate of the attribute.
func (a *Attributes) Clone() *Attributes {
	if a == nil {
		return nil
	}
	attrs := make(map[string]string, len(a.Attrs))
	for k, v := range a.Attrs {
		attrs[k] = v
	}
	return &Attributes{attrs}
}

// Set changes the attribute that a given key has now a given value.
func (a *Attributes) Set(key string, value string) *Attributes {
	if a == nil {
		return &Attributes{map[string]string{key: value}}
	}
	if a.Attrs == nil {
		a.Attrs = make(map[string]string)
	}
	a.Attrs[key] = value
	return a
}

// AddClass adds a value to the class attribute.
func (a *Attributes) AddClass(class string) *Attributes {
	if a == nil {
		return &Attributes{map[string]string{"class": class}}
	}
	classes := a.GetClasses()
	for _, cls := range classes {
		if cls == class {
			return a
		}
	}
	classes = append(classes, class)
	a.Attrs["classes"] = strings.Join(classes, " ")
	return a
}

// GetClasses returns the class values as a string slice
func (a *Attributes) GetClasses() []string {
	if a == nil {
		return nil
	}
	classes, ok := a.Attrs["classes"]
	if !ok {
		return nil
	}
	return strings.Fields(classes)
}
