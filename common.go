package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

func (cmd Common) DeleteNonRepos() {
	var err error
	srcdir := cmd.Path("src")

	var repos = map[string]bool{}
	repos[strings.ToLower(filepath.FromSlash(cmd.Package))] = true

	err = filepath.Walk(srcdir,
		func(path string, info os.FileInfo, err error) error {
			rpath, err := filepath.Rel(srcdir, path)
			if err != nil || rpath == "" || rpath == "." || rpath == ".." || !info.IsDir() {
				return err
			}
			if filepath.Base(rpath) == ".git" {
				repos[strings.ToLower(filepath.Dir(rpath))] = true
				return filepath.SkipDir
			}
			return nil
		})
	ErrFatalf(err, "collect failed: %q", err)

	err = filepath.Walk(srcdir,
		func(path string, info os.FileInfo, err error) error {
			rpath, err := filepath.Rel(srcdir, path)
			if err != nil || rpath == "" || rpath == "." || rpath == ".." {
				return err
			}

			lowpath := strings.ToLower(rpath)

			for repo := range repos {
				if repo == lowpath {
					return filepath.SkipDir
				}
			}

			for repo := range repos {
				if strings.HasPrefix(repo, lowpath) {
					return nil
				}
			}

			err = os.RemoveAll(path)
			if err != nil {
				return err
			}
			return filepath.SkipDir
		})
	ErrFatalf(err, "remove failed: %q", err)
}

func (cmd Common) VendorModules() {
	fmt.Fprintf(os.Stdout, "# Vendoring modules\n")

	workdir, err := os.Getwd()
	ErrFatalf(err, "unable to get working directory: %q\n", err)

	defer func() {
		err = os.Chdir(workdir)
		ErrFatalf(err, "unable to change directory: %q\n", err)
	}()

	err = os.Chdir(cmd.RepoDir())
	ErrFatalf(err, "unable to change directory: %q\n", err)

	err = os.RemoveAll("vendor")
	if os.IsNotExist(err) {
		err = nil
	}
	ErrFatalf(err, "unable to delete vendor: %q\n", err)

	for repeat := 2; repeat > 0; repeat-- {
		gomod := exec.Command("go", "mod", "vendor", "-v")
		gomod.Env = append(os.Environ(), "GO111MODULE=on")
		gomod.Stdout, gomod.Stderr = os.Stdout, os.Stderr
		err = gomod.Run()
		Errf(err, "go mod vendor failed, retrying: %q\n", err)
		if err == nil {
			break
		}
	}
	ErrFatalf(err, "go mod vendor failed: %q\n", err)
}

func (cmd Common) FlattenVendor() {
	fmt.Fprintf(os.Stdout, "# Flattening vendor\n")

	vendordir := filepath.Join(cmd.RepoDir(), "vendor")
	srcdir := cmd.Path("src")
	err := filepath.Walk(vendordir,
		func(path string, info os.FileInfo, err error) error {
			rpath, err := filepath.Rel(vendordir, path)
			if err != nil || rpath == "" || rpath == "." || rpath == ".." {
				return err
			}
			if err != nil {
				return err
			}
			err = os.Rename(
				filepath.Join(vendordir, rpath),
				filepath.Join(srcdir, rpath))
			if err != nil {
				return err
			}
			return filepath.SkipDir
		})
	ErrFatalf(err, "rename failed: %q", err)

	err = os.Remove(vendordir)
	ErrFatalf(err, "unable to delete vendor: %q", err)
}

func ReadModules(path string) []string {
	data, err := ioutil.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}

	if err != nil {
		ErrFatalf(err, "unable to read modules.txt: %q", err)
	}

	unsorted := []string{}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			continue
		}
		if line == "" {
			continue
		}
		unsorted = append(unsorted, line)
	}

	sort.Strings(unsorted)

	modules := []string{}
	before := "\x00"

	for _, line := range unsorted {
		if strings.HasPrefix(line, before) {
			continue
		}
		before = line
		modules = append(modules, line)
	}

	return modules
}
