package gosqlfmt

import (
	"fmt"
	"regexp"
	"strings"
)

func matcher(query string) string {
	match, err := regexp.Match(`(?i)^set`, []byte(query))
	if err!=nil {
		fmt.Println(err)
	}
	if match {
		return query
	}

	match, err = regexp.Match(`(?i)^select`, []byte(query))
	if err!=nil {
		fmt.Println(err)
	}
	if match {
		return FormatSelect(query)
	}

	match, err = regexp.Match(`(?i)^with`, []byte(query))
	if err!=nil {
		fmt.Println(err)
	}
	if match {
		return FormatCTE(query)
	}

	return query
}

func FormatQuery(fileName string) string {

	queries := GetQuery(fileName)

	var finalQuery string 
	for i := range queries {
		query := matcher(CleanQuery(queries[i]))

		if len(finalQuery)>0 {
			finalQuery = finalQuery+";\n\n"+strings.TrimSpace(query)
		} else {
			finalQuery = strings.TrimSpace(query)
		}
	}
	return finalQuery
}
