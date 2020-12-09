package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

const dirArgName = "--dir"
const dictArgName = "--dicts"
const extsArgName = "--exts"

type DictNameSpace map[string]map[string]string

func main() {
	buildDir := "./build"

	args := os.Args[1:]
	err := verifyArgs(args)
	if err != nil {
		log.Fatalf(err.Error())
	}

	os.Mkdir(buildDir, os.ModePerm)
	removeContents(buildDir)

	err, dirPath := getArgValue(args, dirArgName)
	if err != nil {
		log.Fatalf(err.Error())
	}

	err, _dictPaths := getArgValue(args, dictArgName)
	if err != nil {
		log.Fatalf(err.Error())
	}
	dictPaths := strings.Split(_dictPaths, ",")

	err, _whiteListExt := getArgValue(args, extsArgName)
	if err != nil {
		log.Fatalf(err.Error())
	}
	whiteListExt := strings.Split(_whiteListExt, ",")

	fmt.Println("=========================")
	fmt.Println("Root folder: ", dirPath)
	fmt.Println("Dictionary files:", dictPaths)
	fmt.Println("Used extensions:", whiteListExt)
	fmt.Println("")

	ignorePaths := []string{}

	usedLines := make(DictNameSpace)
	ignoredLines := make(DictNameSpace)

	_, files := scanRecursive(dirPath, ignorePaths, whiteListExt)
	for _, file := range files {

		err, contents := scanFile(file)
		if err != nil {
			fmt.Sprintf("An error occurred when scannig file %s", file)
		}
		for _, dict := range dictPaths {
			filename := filepath.Base(dict)
			if usedLines[filename] == nil {
				usedLines[filename] = make(map[string]string)
			}
			if ignoredLines[filename] == nil {
				ignoredLines[filename] = make(map[string]string)
			}

			if err != nil {
				fmt.Printf("Cannot create/open new file")
				continue
			}
			err, dictContents := scanFile(dict)
			if err != nil {
				fmt.Printf("An error occurred while reading a dictionary file %s \n", dict)
				continue
			}

			lines := getDictLines(dictContents)
			for _, line := range lines {
				found, key := getKey(line)
				if !found {
					continue
				}
				re := regexp.MustCompile(key)
				usageFound := re.MatchString(contents)
				if usageFound {
					// update list of used key-values
					usedLines[filename][key] = line
				} else {
					// update list of ignored key-values
					ignoredLines[filename][key] = line
				}

				// remove record from list of ignored key-values if the same found in list of used values
				if _, ok := ignoredLines[filename][key]; ok {
					if _, ok := usedLines[filename][key]; ok {
						delete(ignoredLines[filename], key)
					}
				}
			}
		}
	}

	for _, dict := range dictPaths {
		createDictFiles(usedLines, buildDir, dict, "")
		createDictFiles(ignoredLines, buildDir, dict, "ignored_")
	}
}

func createDictFiles(d DictNameSpace, buildDir string, dict string, prefix string) {
	filename := filepath.Base(dict)
	f, err := os.OpenFile(path.Join(buildDir, prefix+filename), os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		fmt.Printf("Cannot open file %s \n", filename)
		return
	}
	for _, v := range d[filename] {
		f.WriteString(v)
		f.WriteString("\n")
	}
	f.Close()
}

func getDictLines(contents string) []string {
	lines := make([]string, 0)
	sc := bufio.NewScanner(strings.NewReader(contents))
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines
}

func getKey(line string) (bool, string) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") || !strings.Contains(trimmed, "=") {
		return false, ""
	}
	key := strings.Split(trimmed, "=")[0]
	return true, key
}

func scanFile(path string) (error, string) {
	file, err := os.Open(path)
	if err != nil {
		return nil, ""
	}
	result := ""
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	file.Close()
	for _, line := range lines {
		result += fmt.Sprintf("%s \n", line)
	}
	return nil, result
}

func scanRecursive(dir_path string, ignore []string, whiteListExt []string) ([]string, []string) {
	folders := []string{}
	files := []string{}
	filepath.Walk(dir_path, func(path string, f os.FileInfo, err error) error {
		_continue := false
		for _, i := range ignore {
			if strings.Index(path, i) != -1 {
				_continue = true
			}
		}
		if _continue == false {
			f, err = os.Stat(path)
			if err != nil {
				log.Fatal(err)
			}
			f_mode := f.Mode()
			if f_mode.IsDir() {
				folders = append(folders, path)
			} else if f_mode.IsRegular() {
				contains := false
				for _, ext := range whiteListExt {
					if strings.Contains(path, ext) {
						contains = true
					}
				}
				if contains {
					files = append(files, path)
				}
			}
		}

		return nil
	})

	return folders, files
}

func removeContents(dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return err
	}
	for _, file := range files {
		err = os.RemoveAll(file)
		if err != nil {
			return err
		}
	}
	return nil
}

func verifyArgs(args []string) error {
	dicts := false
	dir := false
	exts := false
	for _, argVal := range args {
		if !strings.Contains(argVal, "=") {
			return fmt.Errorf("invalid args")
		}
		key := strings.Split(argVal, "=")[0]
		if key == dirArgName {
			dir = true
		}
		if key == dictArgName {
			dicts = true
		}
		if key == extsArgName {
			exts = true
		}
	}
	if !dicts || !dir || !exts {
		return fmt.Errorf("required args are missing")
	}
	if len(args) != 3 {
		return fmt.Errorf("wrong argument number")
	}
	return nil
}

func getArgValue(args []string, name string) (error, string) {
	for _, argVal := range args {
		if !strings.Contains(argVal, "=") {
			return fmt.Errorf("invalid arguments"), ""
		}
		key := strings.Split(argVal, "=")[0]
		value := strings.Split(argVal, "=")[1]
		if key == name {
			return nil, value
		}
	}
	return fmt.Errorf("wrong argument name"), ""
}
