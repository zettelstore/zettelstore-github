//-----------------------------------------------------------------------------
// Copyright (c) 2020 Detlef Stern
//
// This file is part of zettelstore.
//
// Zettelstore is licensed under the latest version of the EUPL (European Union
// Public License). Please see file LICENSE.txt for your rights and obligations
// under this license.
//-----------------------------------------------------------------------------

// Package progplace provides zettel that inform the user about the internal Zettelstore state.
package progplace

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	"zettelstore.de/z/domain"
)

func genRuntimeM(zid domain.ZettelID) *domain.Meta {
	if myPlace.startConfig == nil {
		return nil
	}
	meta := domain.NewMeta(zid)
	meta.Set(domain.MetaKeyTitle, "Zettelstore Runtime Values")
	meta.Set(domain.MetaKeyRole, "configuration")
	meta.Set(domain.MetaKeySyntax, "zmk")
	meta.Set(domain.MetaKeyVisibility, domain.MetaValueVisibilityOwner)
	meta.Set(domain.MetaKeyReadOnly, "true")
	return meta
}

func (pp *progPlace) genRuntimeC(meta *domain.Meta) string {
	var sb strings.Builder
	sb.WriteString("|=Name|=Value>\n")
	fmt.Fprintf(&sb, "|Number of CPUs|%v\n", runtime.NumCPU())
	fmt.Fprintf(&sb, "|Number of goroutines|%v\n", runtime.NumGoroutine())
	fmt.Fprintf(&sb, "|Number of Cgo calls|%v\n", runtime.NumCgoCall())
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Fprintf(&sb, "|Total allocates bytes|%v\n", m.TotalAlloc)
	fmt.Fprintf(&sb, "|Memory from OS|%v\n", m.Sys)
	fmt.Fprintf(&sb, "|Objects allocated|%v\n", m.Mallocs)
	fmt.Fprintf(&sb, "|Objects freed|%v\n", m.Frees)
	fmt.Fprintf(&sb, "|Objects active|%v\n", m.Mallocs-m.Frees)
	fmt.Fprintf(&sb, "|Heap alloc|%v\n", m.HeapAlloc)
	fmt.Fprintf(&sb, "|Heap sys|%v\n", m.HeapSys)
	fmt.Fprintf(&sb, "|Heap idle|%v\n", m.HeapIdle)
	fmt.Fprintf(&sb, "|Heap in use|%v\n", m.HeapInuse)
	fmt.Fprintf(&sb, "|Heap released|%v\n", m.HeapReleased)
	fmt.Fprintf(&sb, "|Heap objects|%v\n", m.HeapObjects)
	fmt.Fprintf(&sb, "|Stack in use|%v\n", m.StackInuse)
	fmt.Fprintf(&sb, "|Stack sys|%v\n", m.StackSys)
	fmt.Fprintf(&sb, "|Garbage collection metadata|%v\n", m.GCSys)
	fmt.Fprintf(&sb, "|Last garbage collection|%v\n", time.Unix((int64)(m.LastGC/1000000000), 0))
	fmt.Fprintf(&sb, "|Garbage collection goal|%v\n", m.NextGC)
	fmt.Fprintf(&sb, "|Garbage collections|%v\n", m.NumGC)
	fmt.Fprintf(&sb, "|Forced garbage collections|%v\n", m.NumForcedGC)
	fmt.Fprintf(&sb, "|Garbage collection fraction|%.3f%%\n", m.GCCPUFraction*100.0)
	if p := pp.startPlace; p != nil {
		p.Reload(context.Background())
	}
	return sb.String()
}
