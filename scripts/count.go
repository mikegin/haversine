package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var counter int

// Increment the counter and return its value
func getNextCount() string {
	counter++
	return fmt.Sprintf("%d", counter)
}

func getLeadingWhitespace(s string) string {
	var i int
	for i = 0; i < len(s); i++ {
		if s[i] != ' ' && s[i] != '\t' {
			break
		}
	}
	return s[:i]
}

// Process a file to replace 7 with unique integers
func processFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var modifiedLines []string
	scanner := bufio.NewScanner(file)
	hasTrailingNewline := false

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "var __COUNT__ int") {
			line = getLeadingWhitespace(line) + "var __COUNT__ int = " + getNextCount()
		}
		modifiedLines = append(modifiedLines, line)
	}

	// Check if the last character of the file is a newline
	stat, err := file.Stat()
	if err != nil {
		return err
	}
	if stat.Size() > 0 {
		buf := make([]byte, 1)
		file.Seek(stat.Size()-1, 0)
		file.Read(buf)
		if buf[0] == '\n' {
			hasTrailingNewline = true
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	output := strings.Join(modifiedLines, "\n")
	if hasTrailingNewline {
		output += "\n"
	}

	return os.WriteFile(filePath, []byte(output), 0644)
}

// Walk through the directory and process files
func walkDirectory(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") && path != "scripts/count.go" {
			if err := processFile(path); err != nil {
				return err
			}
		}
		return nil
	})
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run main.go <directory>")
		return
	}

	dir := os.Args[1]
	counter = 0

	if err := walkDirectory(dir); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
