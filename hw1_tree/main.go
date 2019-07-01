package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
)

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		fmt.Println("usage go run main.go . [-f]")
		os.Exit(2)
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		log.Fatal(err.Error())
	}
}

var treeSymbol string
var level int

func dirTree(out io.Writer, path string, printFiles bool) (err error) {
	dir, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open dir error: %s %s", err, path)
	}
	objects, err := dir.Readdir(-1)
	if err != nil {
		return fmt.Errorf("read dir error: %s %s", err, path)
	}
	defer dir.Close()

	objects = sortObjects(&objects, printFiles)
	var tabSymbol string

	for j := 0; j < strings.Count(path, string(os.PathSeparator)); j++ {
		if treeSymbol == "├───" {
			tabSymbol += "│"
		}
		tabSymbol += "\t"
	}
	for i, file := range objects {
		if i == len(objects)-1 {
			treeSymbol = "└───"
		} else {
			treeSymbol = "├───"
		}
		if level > 2 && strings.Count(path, "static") == 1 && strings.Count(path, "z_lorem") == 0 {
			tabSymbol = strings.Replace(tabSymbol, "\t", "│\t", 2)
		}
		if level == 1 && strings.Count(path, "z_lorem") == 1 && file.IsDir() && strings.Count(tabSymbol, "│") == 0 {
			tabSymbol = strings.Replace(tabSymbol, "\t", "│\t", 1)
		}
		if strings.Count(tabSymbol, "│") == 0 && !file.IsDir() {
			tabSymbol = strings.Replace(tabSymbol, "\t", "│\t", 1)
		}
		if file.IsDir() {
			_, err = fmt.Fprintln(out, tabSymbol+treeSymbol+file.Name())
			level++
			err = dirTree(out, path+string(os.PathSeparator)+file.Name(), printFiles)
			if err != nil {
				return err
			}
			//
		} else {
			if printFiles {
				if file.Size() > 0 {
					_, err = fmt.Fprintf(out, "%s (%db)\n", tabSymbol+treeSymbol+file.Name(), file.Size())
				} else {
					_, err = fmt.Fprintf(out, "%s (empty)\n", tabSymbol+treeSymbol+file.Name())
				}
			}
		}
	}
	level--
	return nil
}

func sortObjects(objects *[]os.FileInfo, printFiles bool) []os.FileInfo {
	var newObjects []os.FileInfo
	var names []string

	for _, file := range *objects {
		if printFiles {
			names = append(names, file.Name())
		} else {
			if file.IsDir() {
				names = append(names, file.Name())
			}
		}
	}
	sort.Strings(names)

	for i := 0; i < len(names); i++ {
		for _, file := range *objects {
			if file.Name() == names[i] {
				newObjects = append(newObjects, file)
			}
		}
	}
	return newObjects
}
