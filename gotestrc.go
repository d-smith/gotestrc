package main

import (
	"bytes"
	"flag"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

//dirsWithTestFiles is used to build a collection of directories that contain go tests,
//e.g. filenames that end with _test.go
var dirsWithTestFiles = make(map[string]struct{})

func fatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func walkFn(walktpath string, info os.FileInfo, err error) error {

	//Skip directories with Godep in the path component.
	if strings.Contains(walktpath, "Godeps") {
		return nil
	}

	//If the directory contains a test file, add the directory to the collection of
	//directories containing test files.
	if strings.HasSuffix(walktpath, "_test.go") {
		dir, _ := path.Split(walktpath)
		dirsWithTestFiles[dir] = struct{}{}
	}
	return nil
}

//Collect the relevant output needed for the test summary from the test output.
func processTestOutput(testout []byte, buffer *bytes.Buffer) {

	bb := bytes.NewBuffer(testout)
	for {
		line, err := bb.ReadBytes('\n')
		if err != nil {
			break
		}

		linetxt := string(line)
		if strings.HasPrefix(linetxt, "coverage") ||
			strings.HasPrefix(linetxt, "--- FAIL") ||
			strings.HasPrefix(linetxt, "PASS") ||
			strings.HasPrefix(linetxt, "Test directory:") {
			buffer.WriteString(linetxt)
		}

		if strings.HasPrefix(linetxt, "coverage") {
			buffer.WriteRune('\n')
		}
	}

}

//Walk the directories containing tests, collecting test summary output in the
//provided buffer
func walkDirsWithTestFiles(buffer *bytes.Buffer) {

	for k, _ := range dirsWithTestFiles {
		println("Running tests in ", k)

		buffer.WriteString("Test directory: ")
		buffer.WriteString(k)
		buffer.WriteRune('\n')

		//Change working directory to the directory in which to execute the current tests
		err := os.Chdir(k)
		if err != nil {
			println("Error running tests", err.Error())
		}

		//Run the tests and assess coverage
		cmd := exec.Command("godep", "go", "test", "-cover")
		out, err := cmd.CombinedOutput()
		if err != nil {
			println("Error running tests ", err.Error())
		}

		//Grab output for the test summary
		processTestOutput(out, buffer)
	}
}

func main() {

	//Run tests in the current directory as the default behavior
	currentDir, err := os.Getwd()
	fatal(err)

	//Check to see if the default has been overridden via the -root command line option,
	root := flag.String("root", currentDir, "root of tree to walk")
	flag.Parse()
	println("walk ", *root)

	//Change directory to the test root directory for gathering directories with tests
	err = os.Chdir(*root)
	fatal(err)

	//Walk the directory hiearachy from the given root looking for directories containing tests
	filepath.Walk(*root, walkFn)

	//Now go to each directory with tests and execute them
	buffer := bytes.NewBufferString("")
	walkDirsWithTestFiles(buffer)

	//Output test summary
	println("==================================================")
	println("Test Summary")
	println("==================================================")

	println(buffer.String())
}
