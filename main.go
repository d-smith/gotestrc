package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"
	"os/exec"
)



var dirsWithTestFiles = make(map[string]struct{})

func fatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}


func walkFn(path string, info os.FileInfo, err error) error {
	if strings.Contains(path, "Godeps") {
		return nil
	}

	if strings.HasSuffix(path, "_test.go") {
		replacer := strings.NewReplacer(info.Name(), "")
		pathSansFile := replacer.Replace(path)
		dirsWithTestFiles[pathSansFile] = struct{}{}
	}
	return nil
}

func walkDirsWithTestFiles() {

	for k,_ := range dirsWithTestFiles {
		println("Running testss in ", k)
		err := os.Chdir(k)
		if err != nil {
			println("Error running tests", err.Error())
		}
		cmd := exec.Command("godep", "go", "test",  "-cover")
		out, err := cmd.CombinedOutput()
		if err != nil {
			println("Error running tests ", err.Error())
		}

		println(string(out))
	}
}

func main() {

	currentDir, err := os.Getwd()
	fatal(err)

	root := flag.String("root",currentDir, "root of tree to walk")
	flag.Parse()
	println("walk ", *root)

	err = os.Chdir(*root)
	fatal(err)

	filepath.Walk(*root, walkFn)

	walkDirsWithTestFiles()
}
