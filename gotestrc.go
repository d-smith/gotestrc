package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"
	"os/exec"
	"bytes"
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


func processTestOutput(testout []byte, buffer *bytes.Buffer )  {

	bb := bytes.NewBuffer(testout)
	for {
		line, err := bb.ReadBytes('\n')
		if err != nil {
			break
		}

		linetxt := string(line)
		if strings.HasPrefix(linetxt, "coverage") /*|| strings.HasPrefix(linetxt,"ok")*/ ||
				 /*strings.HasPrefix(linetxt, "exit") ||*/ strings.HasPrefix(linetxt, "--- FAIL") ||
				strings.HasPrefix(linetxt, "PASS") ||
				 strings.HasPrefix(linetxt, "Test directory:") {
			buffer.WriteString(linetxt)
		}

		if strings.HasPrefix(linetxt, "coverage") {
			buffer.WriteRune('\n')
		}
	}

}

func walkDirsWithTestFiles(buffer *bytes.Buffer) {

	for k,_ := range dirsWithTestFiles {
		println("Running testss in ", k)

		buffer.WriteString("Test directory: ")
		buffer.WriteString(k)
		buffer.WriteRune('\n')

		err := os.Chdir(k)
		if err != nil {
			println("Error running tests", err.Error())
		}
		cmd := exec.Command("godep", "go", "test",  "-cover")
		out, err := cmd.CombinedOutput()
		if err != nil {
			println("Error running tests ", err.Error())
		}

		processTestOutput(out, buffer)
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

	buffer := bytes.NewBufferString("")
	walkDirsWithTestFiles(buffer)

	println("==================================================")
	println("Test Summary")
	println("==================================================")

	println(buffer.String())
}
