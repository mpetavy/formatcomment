package main

import (
	"bufio"
	"bytes"
	"embed"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mpetavy/common"
)

var (
	input     = flag.String("i", "", "input file or path")
	output    = flag.String("o", "", "output file or path")
	recursive = flag.Bool("r", false, "recursive file scanning")
)

//go:embed go.mod
var resources embed.FS

func init() {
	common.Init("", "", "", "", "Line of code counter", "", "", "", &resources, nil, nil, run, 0)
}

func processJavaFile(filename string, f os.FileInfo) error {
	if f.IsDir() {
		return nil
	}

	if !strings.HasSuffix(f.Name(), ".java") {
		return nil
	}

	inputFile, err := os.ReadFile(filename)
	if common.Error(err) {
		return err
	}

	var (
		buf      bytes.Buffer
		modified bool
	)

	scanner := bufio.NewScanner(bytes.NewReader(inputFile))
	for scanner.Scan() {
		oldLine := scanner.Text()
		newLine := oldLine

		p := strings.Index(newLine, "/*")
		if p != -1 && strings.HasSuffix(newLine, "*/") {
			comment := newLine[p+2 : len(newLine)-2]

			if strings.HasPrefix(comment, "*") {
				comment = comment[1:]
			}

			if strings.HasSuffix(comment, "*") {
				comment = comment[:len(comment)-1]
			}

			comment = strings.TrimSpace(comment)

			newLine = strings.TrimSpace(newLine[:p])

			if len(newLine) == 0 {
				newLine = strings.Repeat(" ", p)
			} else {
				newLine = fmt.Sprintf("%s ", newLine)
			}

			newLine = fmt.Sprintf("%s// %s", newLine, comment)
		} else {
			pComment := strings.Index(newLine, "//")
			pComma := strings.Index(newLine, ";")
			pComma = -1

			if pComment != -1 && pComma != -1 && pComma < pComment {
				comment := strings.TrimSpace(newLine[pComment+2:])
				newLine = strings.TrimRight(newLine[:pComment], " ")

				newLine = fmt.Sprintf("%s // %s", newLine, comment)
			}
		}

		if oldLine != newLine {
			modified = true
		}

		buf.WriteString(newLine + "\n")
	}

	if modified {
		outputFile := *output + filename[len(*input):]
		outputDir := filepath.Dir(outputFile)

		if !common.FileExists(outputDir) {
			err := os.MkdirAll(outputDir, common.DefaultDirMode)
			if common.Error(err) {
				return err
			}
		}

		err := os.WriteFile(outputFile, buf.Bytes(), common.DefaultFileMode)
		if common.Error(err) {
			return err
		}
	}

	return nil
}

func run() error {
	*input = common.CleanPath(*input)
	*output = common.CleanPath(*output)

	err := common.WalkFiles(*input, *recursive, false, func(file string, f os.FileInfo) error {
		if f.IsDir() {
			return nil
		}

		common.Debug("found file: %s", file)

		err := processJavaFile(file, f)
		if common.Error(err) {
			return err
		}

		return nil
	})
	if common.Error(err) {
		return err
	}

	return nil
}

func main() {
	common.Run([]string{"i"})
}
