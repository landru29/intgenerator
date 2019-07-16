package main

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

func parseFile(filename string, structName string) ([]string, error) {
	functionRegexp := regexp.MustCompile(fmt.Sprintf(`(?m)^func\s\([\w]*\s\*?%s\)\s*([\w]*\(.*)`, structName))
	curlyRemover := regexp.MustCompile(`\s*{\s*$`)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return []string{}, err
	}

	matcher := functionRegexp.FindAllStringSubmatch(string(data), -1)

	result := make([]string, len(matcher))
	for i, catcher := range matcher {
		result[i] = curlyRemover.ReplaceAllString(catcher[1], "")
	}

	return result, nil

}

func parsePackage(searchDir string, structName string) ([]string, error) {
	result := []string{}

	packagePath := getFullPath(searchDir)

	fileList, errFileList := getFileList(packagePath)
	if errFileList != nil {
		return result, errFileList
	}

	for _, file := range fileList {
		functions, errParse := parseFile(file, structName)
		if errParse != nil {
			return result, errParse
		}
		for _, function := range functions {
			result = append(result, function)
		}
	}

	return result, nil
}

func getFileList(searchDir string) ([]string, error) {
	fileList := []string{}

	var goFileRegExp = regexp.MustCompile(`\.go$`)
	var notHiddentDirRegExp = regexp.MustCompile(`\/\.\w+|^\.\w+`)
	var vendorRegExp = regexp.MustCompile(`\/vendor\/`)

	files, err := ioutil.ReadDir(searchDir)
	if err != nil {
		return fileList, err
	}

	for _, f := range files {
		if !f.IsDir() && goFileRegExp.MatchString(f.Name()) && !notHiddentDirRegExp.MatchString(searchDir) && !vendorRegExp.MatchString(searchDir) {
			fileList = append(
				fileList,
				strings.Replace(fmt.Sprintf("%s/%s", searchDir, f.Name()), "//", "/", -1),
			)
		}
	}

	return fileList, nil
}

func getFullPath(packageName string) string {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}
	return strings.Replace(fmt.Sprintf("%s/src/%s", gopath, packageName), "//", "/", -1)
}

func generateInterface(structName string, functions []string) string {
	capitalizeRe := regexp.MustCompile(`^[A-Z]`)
	result := []string{fmt.Sprintf("type %sInterface interface {", structName)}
	for _, function := range functions {
		if capitalizeRe.MatchString(function) {
			result = append(result, fmt.Sprintf("    func %s", function))
		}
	}
	result = append(result, "}")
	return strings.Join(result, "\n")
}

func command() *cobra.Command {
	return &cobra.Command{
		Use:  "generate <package> <struct>",
		Long: "./intgenerator generate <package> <struct>",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				fmt.Println(cmd.Usage())
				log.Fatal("You must specify a package name (ie. github.com/spf13/cobra) and a struct name (ie. Command)")
				return
			}

			functionList, errParse := parsePackage(args[0], args[1])
			if errParse != nil {
				log.Fatal(errParse)
				return
			}

			fmt.Println(generateInterface(args[1], functionList))

		},
	}
}

func main() {
	command().Execute()
}
