package main

import (
	"go/build"
	"go/types"
	"log"
	"sort"

	"golang.org/x/tools/go/buildutil"
	"golang.org/x/tools/go/loader"
)

func Main() error {
	var paths []string
	for path := range buildutil.ExpandPatterns(&build.Default, []string{"github.com/juju/juju/..."}) {
		paths = append(paths, path)
	}

	log.Printf("Parsing/type-checking %d package paths...", len(paths))
	const xtest = true
	var loaderConfig loader.Config
	if _, err := loaderConfig.FromArgs(paths, xtest); err != nil {
		return err
	}
	program, err := loaderConfig.Load()
	if err != nil {
		return err
	}

	// Find the state.Application.AddUnit method.
	statePkg := program.Package("github.com/juju/juju/state")
	applicationTypeName := statePkg.Pkg.Scope().Lookup("Application").(*types.TypeName)
	addUnitMethodObj, _, _ := types.LookupFieldOrMethod(applicationTypeName.Type(), true, statePkg.Pkg, "AddUnit")

	initialPackages := program.InitialPackages()
	log.Printf("Filtering %d packages...", len(initialPackages))
	filteredPackages := make(map[string]*loader.PackageInfo)
	var filteredPackagePaths []string
	for _, pkginfo := range initialPackages {
		for _, used := range pkginfo.Uses {
			if used == addUnitMethodObj {
				filteredPackages[pkginfo.Pkg.Path()] = pkginfo
				filteredPackagePaths = append(filteredPackagePaths, pkginfo.Pkg.Path())
				break
			}
		}
	}

	log.Printf("Processing %d packages...", len(filteredPackages))
	sort.Strings(filteredPackagePaths)
	for _, path := range filteredPackagePaths {
		log.Printf("Processing packages %q", path)
	}

	return nil
}

func main() {
	if err := Main(); err != nil {
		log.Fatal(err)
	}
}
