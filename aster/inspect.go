// Copyright 2018 henrylee2cn. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aster

import (
	"go/ast"
	"go/types"
	"log"
)

func (p *PackageInfo) check() {
	log.Printf("Checking package %s...", p.String())
L:
	for ident, obj := range p.info.Defs {
		switch GetObjKind(obj) {
		case Bad, Lbl, Bui, Nil:
			continue L
		case Var:
			if GetTypKind(obj.Type()) == Struct {
				break
			}
			nodes, _ := p.pathEnclosingInterval(ident.Pos(), ident.End())
			for _, n := range nodes {
				if _, ok := n.(*ast.Field); ok {
					continue L
				}
			}
		}
		p.addFacade(ident, obj)
	}
}

// Inspect traverses created and imported packages in the program.
func (prog *Program) Inspect(fn func(Facade) bool) {
	for _, pkg := range prog.InitialPackages() {
		for _, fa := range pkg.facades {
			if !fn(fa) {
				return
			}
		}
	}
}

// Lookup lookups facades in the program.
//
// Match any name if name="";
// Match any ObjKind if objKindSet=0 or objKindSet=AnyObjKind;
// Match any TypKind if typKindSet=0 or typKindSet=AnyTypKind;
//
func (prog *Program) Lookup(objKindSet ObjKind, typKindSet TypKind, name string) (list []Facade) {
	prog.Inspect(func(fa Facade) bool {
		if (name == "" || fa.Name() == name) &&
			(typKindSet == 0 || fa.TypKind().In(typKindSet)) &&
			(objKindSet == 0 || fa.ObjKind().In(objKindSet)) {
			list = append(list, fa)
		}
		return true
	})
	return
}

// FindFacade finds Facade by types.Type in the program.
func (prog *Program) FindFacade(typ types.Type) (fa Facade, found bool) {
	for _, pkg := range prog.allPackages {
		fa, found = pkg.FindFacade(typ)
		if found {
			return
		}
	}
	return
}

// Inspect traverses facades in the package.
func (p *PackageInfo) Inspect(fn func(Facade) bool) {
	for _, fa := range p.facades {
		if !fn(fa) {
			return
		}
	}
}

// Lookup lookups facades in the package.
//
// Match any name if name="";
// Match any ObjKind if objKindSet=0 or objKindSet=AnyObjKind;
// Match any TypKind if typKindSet=0 or typKindSet=AnyTypKind;
//
func (p *PackageInfo) Lookup(objKindSet ObjKind, typKindSet TypKind, name string) (list []Facade) {
	p.Inspect(func(fa Facade) bool {
		if (name == "" || fa.Name() == name) &&
			(typKindSet == 0 || fa.TypKind().In(typKindSet)) &&
			(objKindSet == 0 || fa.ObjKind().In(objKindSet)) {
			list = append(list, fa)
		}
		return true
	})
	return
}

// FindFacade finds Facade by types.Type in the package.
func (p *PackageInfo) FindFacade(typ types.Type) (fa Facade, found bool) {
	facade, idx := p.getFacadeByTyp(typ)
	return facade, idx != -1
}

func (p *PackageInfo) getFacade(ident *ast.Ident) (facade *facade, idx int) {
	for idx, facade = range p.facades {
		if facade.ident == ident {
			return
		}
	}
	return nil, -1
}

func (p *PackageInfo) getFacadeByObj(obj types.Object) (facade *facade, idx int) {
	for idx, facade = range p.facades {
		if facade.obj == obj {
			return
		}
	}
	return nil, -1
}

func (p *PackageInfo) getFacadeByTyp(t types.Type) (facade *facade, idx int) {
	for idx, facade = range p.facades {
		if facade.obj.Type() == t || facade.typ() == t {
			return
		}
	}
	return nil, -1
}

func (p *PackageInfo) addFacade(ident *ast.Ident, obj types.Object) {
	p.facades = append(p.facades, &facade{
		obj:   obj,
		pkg:   p,
		ident: ident,
		doc:   p.docComment(ident),
	})
}

func (p *PackageInfo) removeFacade(ident *ast.Ident) {
	_, idx := p.getFacade(ident)
	if idx >= 0 {
		p.facades = append(p.facades[:idx], p.facades[idx+1:]...)
	}
}
