package gosqlfmt

import (
	"bytes"
	"log"
	"regexp"
	"strings"
)

var (
	// logger config
	buf bytes.Buffer
	logger = log.New(&buf, "logger: ", log.Lshortfile)

	// regexp definitions
	whitespace = regexp.MustCompile(`\s{2,}`)
	firstline = regexp.MustCompile(`\A\s*`)
	
	selectcolumns = regexp.MustCompile(`(?i)select(.*?)from`)
	fmtcolumns = regexp.MustCompile(`(?i),\s`)

	fromtables = regexp.MustCompile(`(?i)from(.*)(where|group by|order by|having|qualify)(.*)`)
	fromtables2 = regexp.MustCompile(`(?i)from(.*)`)
	fmttables = regexp.MustCompile(`(?i)(^[a-zA-Z][a-zA-Z\._0-9]*[\s]*[a-zA-Z0-9]*)\s((left|right|full|join|lateral)*(.*))`)
	joins_etc = regexp.MustCompile(`(?i)(lateral|left|right|full)`)

	wherecond = regexp.MustCompile(`(?i)where(.*)(group by|order by|having|qualify)`)
	wherecond2 = regexp.MustCompile(`(?i)where(.*)`)
	fmtwhere = regexp.MustCompile(`(?i)(and|or)`)
)

func CleanQuery(query string) string {
	logger.Println("Cleaning query...")
	// remove all new lines
	query = strings.Replace(query, "\n", " ", -1)

	// remove extra white spaces
	query = string(whitespace.ReplaceAll([]byte(query), []byte(" ")))

	return query
}

func FormatSelect(query string) string {
	logger.Println("Formatting columns...")
	// FORMAT COLUMNS
	columns := selectcolumns.FindStringSubmatch(query)[1]

	columns = string(fmtcolumns.ReplaceAll([]byte(columns), []byte(",\n\t")))
	columns = string(firstline.ReplaceAll([]byte(columns), []byte("\t$1")))

	returnquery := "\nselect\n"+columns

	logger.Println("Formatting from clause...")
	// FORMAT TABLES
	table_li := fromtables.FindStringSubmatch(query)
	var tables string
	if len(table_li)==0 {
		tables = fromtables2.FindStringSubmatch(query)[1]
	} else {
		tables = table_li[1]
	}

	tables = string(firstline.ReplaceAll([]byte(tables), []byte("")))
	joins_li := fmttables.FindStringSubmatch(tables)
	primary_table := joins_li[1]
	
	returnquery = returnquery+"\nfrom\n\t"+primary_table

	var joins string
	if len(joins_li)>=2 {
		joins = joins_li[2]
		joins = string(joins_etc.ReplaceAll([]byte(joins), []byte("\n\t$1")))
		returnquery = returnquery + joins
	}

	logger.Println("Formatting where clause...")
	// FORMAT WHERE
	conditions := wherecond.FindStringSubmatch(query)
	var conds string
	if len(conditions)==0 {
		where_li := wherecond2.FindStringSubmatch(query)
		if len(where_li)>0 {
			conds = where_li[1]
			conds = string(firstline.ReplaceAll([]byte(conds), []byte("")))
			conds = string(fmtwhere.ReplaceAll([]byte(conds), []byte("\n\t$1")))
			returnquery = returnquery + "\nwhere\n\t" + conds
		}
	} else {
		conds = conditions[1]
		conds = string(firstline.ReplaceAll([]byte(conds), []byte("")))
		conds = string(fmtwhere.ReplaceAll([]byte(conds), []byte("\n\t$1")))
		returnquery = returnquery + "\nwhere\n\t" + conds
	}

	return returnquery 
}

func FormatSet(query string) string {
	query = CleanQuery(query)
	
	query = FormatSelect(query)

	return query
}
