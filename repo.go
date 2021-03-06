// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

// Package repoctl provides functions for managing Arch repositories.
package repoctl

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/goulash/osutil"
	"github.com/goulash/pacman/pkgutil"
)

type Repo struct {
	// Directory is the absolute path to the directory where the
	// packages are stored.
	Directory string
	// Database is the relative path to the repository database,
	// relative from Directory.
	Database string

	// Backup specifies whether to backup old packages.
	Backup bool
	// BackupDir specifies where old packages are backed up to.
	// If the path is not absolute, then it is interpreted as
	// relative to the repository directory.
	BackupDir string
	// IgnoreUpgrades specifies which packages to ignore when looking
	// for upgrades. Explicitely specifying the file will override the
	// ignore however.
	IgnoreAUR []string

	// AddParameters are parameters to add to the repo-add
	// command line.
	AddParameters []string
	// RemoveParameters are parameters to add to the repo-remove
	// command line.
	RemoveParameters []string

	// Error, Info, and Debug is where output is written to. If set to
	// nil, no output is written.
	Error io.Writer
	Info  io.Writer
	Debug io.Writer
}

// New creates a new default configuration with repo as the repository
// database. It is assumed that the database resides in the same
// directory as the packages.
//
// If repo is invalid (because it is absolute), nil is returned.
// We don't check for database existance, because at this point,
// it might not exist yet.
func New(repo string) *Repo {
	if !path.IsAbs(repo) {
		return nil
	}

	return &Repo{
		Directory: path.Dir(repo),
		Database:  path.Base(repo),
		BackupDir: `backup`,

		Error: os.Stderr,
		Info:  os.Stdout,
		Debug: nil,

		IgnoreAUR:        make([]string, 0),
		AddParameters:    make([]string, 0),
		RemoveParameters: make([]string, 0),
	}
}

// Name returns the name of the repository, which is interpreted to be
// the name of the database up to the first period.
func (r *Repo) Name() string {
	base := path.Base(r.Database)
	return base[:strings.IndexByte(base, '.')]
}

// IgnoreFltr returns a FilterFunc for filtering out packages that should
// be ignored. For example, for a list of meta.Packages:
//
//  pkgs = pkgutil.Filter(pkgs, r.ignoreFltr()).(meta.Packages)
//
func (r *Repo) IgnoreFltr() pkgutil.FilterFunc {
	return pkgutil.NameFltr(r.IgnoreAUR).Not()
}

// IgnoreMap returns a map of packages to ignore.
func (r *Repo) IgnoreMap() map[string]bool {
	m := make(map[string]bool)
	for _, i := range r.IgnoreAUR {
		m[i] = true
	}
	return m
}

// AssertSetup returns nil if a normal repository setup is present:
// the directory exists.
//
// While it would make sense to check for readability and writability,
// in modern systems there are so many ways to achieve this, that to
// test all of them is more effort than it is worth.
func (r *Repo) AssertSetup() error {
	if !path.IsAbs(r.Directory) {
		return ErrRepoDirRelative
	}

	ex, err := osutil.DirExists(r.Directory)
	if err != nil {
		return err
	}
	if !ex {
		return ErrRepoDirMissing
	}

	return nil
}

// Setup creates the directory and returns an error if not possible.
func (r *Repo) Setup() error {
	if err := r.AssertSetup(); err != ErrRepoDirMissing {
		return err
	}

	return os.MkdirAll(r.Directory, os.ModePerm)
}

func (r *Repo) printf(format string, obj ...interface{}) {
	if r.Info != nil {
		fmt.Fprintf(r.Info, format, obj...)
	}
}

func (r *Repo) errorf(format string, obj ...interface{}) {
	if r.Error != nil {
		fmt.Fprintf(r.Error, format, obj...)
	}
}

func (r *Repo) debugf(format string, obj ...interface{}) {
	if r.Debug != nil {
		fmt.Fprintf(r.Debug, format, obj...)
	}
}
