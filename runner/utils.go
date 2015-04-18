package runner

import (
	"fmt"
	"os"
	pth "path"
	"path/filepath"
	"regexp"
	"strings"
)

func initFolders() {
	runnerLog("InitFolders")
	path := tmpPath()
	runnerLog("mkdir %s", path)
	err := os.Mkdir(path, 0755)
	if err != nil {
		runnerLog(err.Error())
	}
}

func isTmpDir(path string) bool {
	absolutePath, _ := filepath.Abs(path)
	absoluteTmpPath, _ := filepath.Abs(tmpPath())

	return absolutePath == absoluteTmpPath
}

func isWatchedFileJudgedByShellPattern(path string) bool {
	absolutePath, _ := filepath.Abs(path)

	// valid pattern first, then ext
	shellpattern := settings["valid_shell_pattern"]
	fmt.Println("judge by shellpattern: ", shellpattern)
	if shellpattern != "" {
		fmt.Println("\n")
		root, _ := filepath.Abs(settings["root"])

		for i, p := range strings.Split(shellpattern, ",") {
			fmt.Println("pattern is:", p, i, len(strings.Split(shellpattern, ",")))
			relpath := absolutePath
			fmt.Println("path:", relpath)
			fmt.Println("root:", root)
			if strings.HasPrefix(absolutePath, root) {
				relpath = relpath[len(root):]
				if relpath[0] == '/' {
					relpath = relpath[1:]
				}
			}
			fmt.Println("relpath is: ", relpath)
			if m, e := pth.Match(strings.TrimSpace(p), relpath); m && e == nil {
				fmt.Println("===Matched=== ", relpath, "(relpath) matchs to patter:", p)
				return true
			} else {
				fmt.Println("!!!Not Match!!", relpath, "(relpath), not match to patter:", p)
				return false
			}

		}
		return false
	}

	return false
}

func isWatchedFileJudgedByRegexp(path string) bool {
	// fmt.Println("\n============= regexp checking:", path, " ==============\n")
	absolutePath, _ := filepath.Abs(path)
	root, _ := filepath.Abs(settings["root"])
	// fmt.Println("abs path:", absolutePath)
	// fmt.Println("abs root:", root)

	relpath := strings.TrimPrefix(absolutePath, root)
	relpath = strings.TrimPrefix(relpath, "/")
	// fmt.Println("relpath is: ", relpath)

	// invalidpattern (balcklist) checking
	isvalid := true
	invalidpattern := settings["invalid_regexp"]
	// fmt.Println("Judging by regexp: invalidpattern is: ", invalidpattern)
	if invalidpattern != "" {
		// fmt.Println("\n------------- blacklist checking :", relpath)
		for _, p := range strings.Split(invalidpattern, ",") {
			// fmt.Println("current pattern is:", p, ", cycle index:", i, "of", len(strings.Split(invalidpattern, ","))-1)
			if m, e := regexp.MatchString(strings.TrimSpace(p), relpath); m && e == nil {
				isvalid = false
				// fmt.Println("Blacklist===Matched to invalid pattern === ", relpath, "(relpath) matchs to patter:", p, "now return!")
				break
			} else {
				// fmt.Println("Blacklist!!!Not Match to invalid pattern!!", relpath, "(relpath), not match to patter:", p)
			}
		}
	}

	// now goes to validpattern(whitelist) checking
	if !isvalid {
		// fmt.Println("-----------> Not passed: ", path)
		return false
	}

	validpattern := settings["valid_regexp"]
	// fmt.Println("Judging by regexp: validpattern is: ", validpattern)
	if validpattern != "" {
		// fmt.Println("\n************* whitelist checking :", relpath)

		for _, p := range strings.Split(validpattern, ",") {
			// fmt.Println("validpattern is:", p, i, len(strings.Split(validpattern, ",")))
			if m, e := regexp.MatchString(strings.TrimSpace(p), relpath); m && e == nil {
				isvalid = true
				// fmt.Println("Whitelist passed! ===Matched to valid pattern === ", relpath, "(relpath) matchs to patter:", p)
				// fmt.Println("-----------> Passed: ", path)
				watcherLog("%q changed. reloading......", path)
				return true
			} else {
				// fmt.Println("Whitelist!!!Not Match to valid pattern!!", relpath, "(relpath), not match to patter:", p)
			}
		}
	}

	return false
}

func isWatchedFile(path string) bool {
	absolutePath, _ := filepath.Abs(path)
	absoluteTmpPath, _ := filepath.Abs(tmpPath())

	if strings.HasPrefix(absolutePath, absoluteTmpPath) {
		return false
	}

	// valid pattern first, then ext
	// shellpattern := settings["valid_shell_pattern"]
	// fmt.Println("shellpattern is: ", shellpattern)

	switch {
	case settings["valid_regexp"] != "" || settings["invalid_regexp"] != "":
		{
			return isWatchedFileJudgedByRegexp(path)
		}

	// case settings["valid_shell_pattern"] != "":
	// 	{
	// 		return isWatchedFileJudgedByShellPattern(path)
	// 	}

	default:
		{
			ext := filepath.Ext(path)

			for _, e := range strings.Split(settings["valid_ext"], ",") {
				if strings.TrimSpace(e) == ext {
					return true
				}
			}
		}
	}

	return false
}

func createBuildErrorsLog(message string) bool {
	file, err := os.Create(buildErrorsFilePath())
	if err != nil {
		return false
	}

	_, err = file.WriteString(message)
	if err != nil {
		return false
	}

	return true
}

func removeBuildErrorsLog() error {
	err := os.Remove(buildErrorsFilePath())

	return err
}
