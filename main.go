package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/exp/slices"
)

var ignore = []string{
	"node_modules",
	".git",
	".next",
}

type commentString struct {
	inline string
	start  string
	end    string
}

var commentStrings = map[string]commentString{
	".c":    {"//", "/*", "*/"},
	".css":  {"", "/*", "*/"},
	".go":   {"//", "/*", "*/"},
	".html": {"", "<!--", "-->"},
	".js":   {"//", "/*", "*/"},
	".jsx":  {"//", "/*", "*/"}, // TODO: comments within JSX blocks
	".lua":  {"--", "--[[", "]]"},
	".py":   {"#", "\"\"\"", "\"\"\""},
	".scss": {"//", "/*", "*/"},
	".sh":   {"#", "", ""},
	".ts":   {"//", "/*", "*/"},
	".tsx":  {"//", "/*", "*/"},
}

var errUnsupportedFiletype = errors.New("unsupported filetype")

func sloc(filePath string) (int, error) {
	commStr, ok := commentStrings[filepath.Ext(filePath)]
	if !ok {
		return 0, fmt.Errorf("%s: %w", filePath, errUnsupportedFiletype)
	}

	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to read file: %w", filePath, err)
	}

	content := string(bytes)

	start := commStr.start
	end := commStr.end

	if start != "" && end != "" {
		pattern := regexp.QuoteMeta(start) + `[\s\S]*?` + regexp.QuoteMeta(end)
		regex := regexp.MustCompile(pattern)
		content = regex.ReplaceAllString(content, "")
	}

	lines := strings.Split(content, "\n")

	total := 0

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		if trimmedLine != "" && (commStr.inline == "" || !strings.HasPrefix(trimmedLine, commStr.inline)) {
			total++
		}
	}

	return total, nil
}

type item struct {
	path string
	sloc int
}

func walk(root string) ([]item, int, int, error) {
	items := []item{}
	total := 0
	pathMaxLen := 0

	err := filepath.WalkDir(root, func(path string, dirEntry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if slices.Contains(ignore, filepath.Base(path)) {
			return filepath.SkipDir
		}

		if dirEntry.Type().IsRegular() && !dirEntry.IsDir() {
			lines, err := sloc(path)
			if err != nil && !errors.Is(err, errUnsupportedFiletype) {
				return err
			}
			// TODO: improve this
			if !errors.Is(err, errUnsupportedFiletype) {
				relativePath, _ := filepath.Rel(root, path) // TODO: handle error
				items = append(items, item{relativePath, lines})
				total += lines
				if l := len(relativePath); l > pathMaxLen {
					pathMaxLen = l
				}

			}
		}
		return nil
	})
	if err != nil {
		return []item{}, 0, 0, err
	}

	return items, total, pathMaxLen, nil
}

func intlen(n int) int {
	count := 0
	if n == 0 {
		return 1
	}
	for n != 0 {
		n /= 10
		count++
	}
	return count
}

func print(items []item, total int, pathMaxLen int) {
	slocMaxLen := intlen(total)

	printItem := func(path string, sloc int) {
		pathPad := strings.Repeat(" ", pathMaxLen-len(path))
		slocPad := strings.Repeat(" ", slocMaxLen-intlen(sloc))
		fmt.Printf("│ %s%s │ %d%s │\n", path, pathPad, sloc, slocPad)
	}

	pathHLine := strings.Repeat("─", pathMaxLen)
	slocHLine := strings.Repeat("─", slocMaxLen)
	fmt.Printf("┌─%s─┬─%s─┐\n", pathHLine, slocHLine)

	for _, item := range items {
		printItem(item.path, item.sloc)
	}

	fmt.Printf("├─%s─┼─%s─┤\n", pathHLine, slocHLine)

	printItem("Total", total)

	fmt.Printf("└─%s─┴─%s─┘\n", pathHLine, slocHLine)
}

func main() {
	ignoreFlag := flag.String("i", "", "List of entries to ignore (comma separated)")
	flag.Parse()

	for _, entry := range strings.Split(*ignoreFlag, ",") {
		ignore = append(ignore, strings.TrimSpace(entry))
	}

	root := "."

	argc := flag.NArg()
	if argc == 1 {
		root = flag.Arg(0)
	}

	items, total, pathMaxLen, err := walk(root)
	if err != nil {
		log.Fatal(err)
	}

	print(items, total, pathMaxLen)
}
