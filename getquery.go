package gosqlfmt

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

func GetQuery(fileName string) []string {
	file, err := os.Open(fileName)
	if err!=nil {
		fmt.Printf("Error opening %s\n%s", fileName, err)
	}

	reader := bufio.NewReader(file)

	data, err := io.ReadAll(reader)
	if err!=nil {
		fmt.Printf("Error reading %s\n%s", fileName, err)
	}

	contents := string(data)

	queries := strings.Split(contents, ";")
	return queries
}

