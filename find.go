// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package repoctl

import (
	"fmt"

	"github.com/goulash/errs"
	"github.com/goulash/osutil"
	"github.com/goulash/pacman"
	"github.com/goulash/pacman/aur"
	"github.com/goulash/pacman/pkgutil"
)

// Upgrade represents an available AUR upgrade.
type Upgrade struct {
	Old *pacman.Package
	New *aur.Package
}

// Name returns the package name of the upgrade.
func (u *Upgrade) Name() string {
	return u.New.Name
}

func (u *Upgrade) Base() string {
	return u.New.PackageBase
}

// DownloadURL returns the download URL of the upgrade.
func (u *Upgrade) DownloadURL() string {
	return u.New.DownloadURL()
}

// Versions returns the old and the new version of the packages.
func (u *Upgrade) Versions() (from string, to string) {
	if u.Old == nil {
		return "", u.New.Version
	}
	return u.Old.Version, u.New.Version
}

// String returns a representation of the available upgrade, in the
// form:
//
//  pkgname: oldver -> newver
func (u *Upgrade) String() string {
	from, to := u.Versions()
	if from == "" {
		return fmt.Sprintf("%s: %s", u.Name(), to)
	}
	return fmt.Sprintf("%s: %s -> %s", u.Name(), from, to)
}

// Upgrades is a list of Upgrade, which can be sorted using sort.Sort.
type Upgrades []*Upgrade

func (u Upgrades) Len() int           { return len(u) }
func (u Upgrades) Swap(i, j int)      { u[i], u[j] = u[j], u[i] }
func (u Upgrades) Less(i, j int) bool { return u[i].New.Name < u[j].New.Name }

// FindUpgrades finds all upgrades it finds to the given packages. If
// no package names are given, all available package names are searched.
func (r *Repo) FindUpgrades(h errs.Handler, pkgnames ...string) (Upgrades, error) {
	errs.Init(&h)
	pkgs, err := r.FindNewest(h, pkgnames...)
	if err != nil {
		return nil, err
	}
	if len(pkgnames) == 0 && len(r.IgnoreUpgrades) != 0 {
		iu := r.ignoreMap()
		pkgs = pkgutil.Filter(pkgs, func(p *pacman.Package) bool {
			return !iu[p.Name]
		})
	}
	as, err := aur.ReadAll(pkgutil.Map(pkgs, pkgutil.PkgName))
	if err != nil {
		if e, ok := err.(*aur.NotFoundError); ok {
			r.debugf(e.Error())
		} else {
			return nil, err
		}
	}

	am := make(map[string]*aur.Package)
	for _, p := range as {
		am[p.Name] = p
	}
	var upgrades Upgrades
	for _, p := range pkgs {
		a := am[p.Name]
		if a != nil && a.PacmanPkg().Newer(p) {
			upgrades = append(upgrades, &Upgrade{p, a})
		}
	}
	return upgrades, nil
}

// FindNewest returns the newest package files found for the given
// package names. If no names are given, all names are searched for.
func (r *Repo) FindNewest(h errs.Handler, pkgnames ...string) (pacman.Packages, error) {
	errs.Init(&h)

	var pkgs pacman.Packages
	var err error
	if len(pkgnames) == 0 {
		pkgs, err = r.ReadDirectory(h)
	} else {
		pkgs, err = r.ReadNames(h, pkgnames...)
	}
	if err != nil {
		return nil, err
	}

	return pkgutil.FilterNewest(pkgs), nil
}

// FindSimilar finds package files in the repository that
// have the same package name. The pkgfile given is filtered from
// the results.
//
// This makes this function useful for finding all the other
// packages given a particular package file.
//
// If no pkgfiles are given, nil is returned.
func (r *Repo) FindSimilar(h errs.Handler, pkgfiles ...string) (pacman.Packages, error) {
	errs.Init(&h)
	if len(pkgfiles) == 0 {
		r.debugf("repoctl.(Repo).FindSimilar: pkgfiles empty.\n")
		return nil, nil
	}

	pkgs, err := pacman.ReadFiles(h, pkgfiles...)
	if err != nil {
		return nil, err
	}
	similar, err := r.ReadNames(h, pkgutil.Map(pkgs, pkgutil.PkgName)...)
	if err != nil {
		return nil, err
	}

	basefiles := pkgutil.MapBool(pkgs, pkgutil.PkgFilename)
	return pkgutil.Filter(similar, func(p *pacman.Package) bool {
		return !basefiles[p.Filename]
	}), nil
}

// FindUpdates finds all given packages with updates. If no names are
// given, all names are checked.
func (r *Repo) FindUpdates(h errs.Handler, pkgnames ...string) (pacman.Packages, error) {
	errs.Init(&h)

	// Case len(pkgnames) == 0 handled by FindNewest
	pkgs, err := r.FindNewest(h, pkgnames...)
	if err != nil {
		return nil, err
	}
	dbpkgs, err := r.ReadDatabase()
	if err != nil {
		return nil, err
	}
	var updates pacman.Packages
	db := pkgutil.MapPkg(dbpkgs, pkgutil.PkgName)
	for _, p := range pkgs {
		dp := db[p.Name]
		if dp == nil || p.Newer(dp) || !r.Exists(dp) {
			updates = append(updates, p)
		}
	}
	return updates, nil
}

// FindMissing returns all packages from the database that do not
// have associated files existing.
func (r *Repo) FindMissing() (pacman.Packages, error) {
	pkgs, err := r.ReadDatabase()
	if err != nil {
		return nil, err
	}

	return pkgutil.Filter(pkgs, func(p *pacman.Package) bool {
		return !r.Exists(p)
	}), nil
}

// OnlyNames reads all possible package names from the repository.
// This includes packages from the database where the files have been
// deleted.
func (r *Repo) OnlyNames(h errs.Handler) ([]string, error) {
	errs.Init(&h)
	dbpkgs, err := r.ReadDatabase()
	if err = h(err); err != nil {
		return nil, err
	}
	filepkgs, err := r.ReadDirectory(h)
	if err = h(err); err != nil {
		return nil, err
	}

	m := pkgutil.MapBool(dbpkgs, pkgutil.PkgName)
	for _, p := range filepkgs {
		m[p.Name] = true
	}
	names := make([]string, 0, len(m))
	for k := range m {
		names = append(names, k)
	}
	return names, nil
}

// Exists checks the existance of a package file; this is only necessary
// for packages read from the database. If the file can't be read for
// any reason, then chances are any client will not be able to read it
// either, and so false is returned.
func (r *Repo) Exists(p *pacman.Package) bool {
	ex, err := osutil.FileExists(p.Filename)
	return err != nil || !ex
}
