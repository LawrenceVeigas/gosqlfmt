package gosqlfmt

import (
	"regexp"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
)

// map of everything within ()
var mapper = make(map[string]string)

var (
	// regexp definitions
	whitespace = regexp.MustCompile(`\s{2,}`)
	firstline  = regexp.MustCompile(`\A\s*`)

	selectcolumns = regexp.MustCompile(`(?i)select(.*?)from`)
	fmtcolumns    = regexp.MustCompile(`(?i),\s`)

	fromtables  = regexp.MustCompile(`(?i)from(.*)(where|group by|order by|having|qualify)(.*)`)
	fromtables2 = regexp.MustCompile(`(?i)from(.*)`)
	fmttables   = regexp.MustCompile(`(?i)(^[a-zA-Z][a-zA-Z\._0-9]*[\s]*[a-zA-Z0-9]*)\s*((left|right|full|join|lateral)*(.*))`)
	joins_etc   = regexp.MustCompile(`(?i)(lateral|left|right|full)`)

	wherecond  = regexp.MustCompile(`(?i)where(.*)(group by|order by|having|qualify)`)
	wherecond2 = regexp.MustCompile(`(?i)where(.*)`)
	fmtwhere   = regexp.MustCompile(`(?i)(and|or)`)
)

func CleanQuery(query string) string {
	// logger.Println("Cleaning query...")
	// remove all new lines
	query = strings.Replace(query, "\n", " ", -1)

	// remove extra white spaces
	query = string(whitespace.ReplaceAll([]byte(query), []byte(" ")))

	// remove leading/trailing whitespace
	query = strings.TrimSpace(query)

	return query
}

func findPos(text string, openPos int) int {
	closePos := openPos
	counter := 1
	for {
		if counter == 0 {
			break
		}

		closePos += 1

		if closePos > len(text) {
			panic("No corresponding bracket found")
		}

		c := text[closePos]
		if c == '(' {
			counter++
		} else if c == ')' {
			counter--
		}
	}

	return closePos
}

func ReplaceBrackets(query string) string {

	var openBracketIndex []int
	var closeBracketIndex []int

	openBracketIndices := regexp.MustCompile(`\(`).FindAllStringIndex(query, -1)

	for i := range openBracketIndices {
		openBracketIndex = append(openBracketIndex, openBracketIndices[i][1])
		closeBracketIndex = append(closeBracketIndex, findPos(query, openBracketIndex[i]))
	}

	log.Debugf("Opening bracket indices:\t%v\n", openBracketIndex)
	log.Debugf("Closing bracket indices:\t%v\n", closeBracketIndex)

	swapopen := make([]int, len(openBracketIndex))
	copy(swapopen, openBracketIndex)
	swapclose := make([]int, len(closeBracketIndex))
	sortindex := make([]int, len(swapopen))

	sort.Sort(sort.Reverse(sort.IntSlice(swapopen)))

	// Order brackets such that inmost bracket is picked first
	for i := range openBracketIndex {
		for j := range swapopen {
			if openBracketIndex[i] == swapopen[j] {
				sortindex[i] = j
			}
		}
	}
	for i := range sortindex {
		swapclose[i] = closeBracketIndex[sortindex[i]]
	}

	log.Debugf("Sorted open:\t%v\n", swapopen)
	log.Debugf("Sorted close:\t%v\n", swapclose)

	returnquery := query

	for i := range swapopen {
		text := query[swapopen[i]:swapclose[i]]
		log.Info(text)
		for key, value := range mapper {
			if strings.Contains(text, value) {
				text = strings.ReplaceAll(text, value, key)
			}
		}

		randvar := RandStringBytes(10)
		mapper[randvar] = text

		returnquery = strings.ReplaceAll(returnquery, text, randvar)
	}
	return returnquery
}

func FormatSelect(query string) string {
	// logger.Println("Formatting columns...")
	// FORMAT COLUMNS
	log.Debugf("Query under consideration:\n%v\n\n", query)
	log.Debugf("Checking for columns in select:\t%v\n\n", selectcolumns.Match([]byte(query)))

	columns := selectcolumns.FindStringSubmatch(query)[1]
	log.Debugf("Columns:\n%v\n", columns)

	columns = string(fmtcolumns.ReplaceAll([]byte(columns), []byte(",\n\t")))
	columns = string(firstline.ReplaceAll([]byte(columns), []byte("\t$1")))

	returnquery := "\nselect\n" + columns

	// logger.Println("Formatting from clause...")
	// FORMAT TABLES
	table_li := fromtables.FindStringSubmatch(query)
	var tables string
	if len(table_li) == 0 {
		tables = fromtables2.FindStringSubmatch(query)[1]
	} else {
		tables = table_li[1]
	}

	log.Info("Parsing from clause...")
	tables = string(firstline.ReplaceAll([]byte(tables), []byte("")))
	joins_li := fmttables.FindStringSubmatch(tables)
	primary_table := joins_li[1]
	log.Debugf("Primary table:%v\n", primary_table)
	log.Debugf("Joins found:%v\n", joins_li)

	returnquery = returnquery + "\nfrom\n\t" + primary_table

	var joins string
	if len(joins_li) >= 2 {
		joins = joins_li[2]
		joins = string(joins_etc.ReplaceAll([]byte(joins), []byte("\n\t$1")))
		returnquery = returnquery + joins
	}

	log.Info("Formatting where clause...")
	// FORMAT WHERE
	conditions := wherecond.FindStringSubmatch(query)
	var conds string
	if len(conditions) == 0 {
		where_li := wherecond2.FindStringSubmatch(query)
		if len(where_li) > 0 {
			for t := range where_li {
				log.Debugf("Conditions in WHERE:%v: %v\n", t, where_li[t])
			}

			conds = where_li[1]
			conds = string(firstline.ReplaceAll([]byte(conds), []byte("")))
			conds = string(fmtwhere.ReplaceAll([]byte(conds), []byte("\n\t$1")))
			returnquery = returnquery + "\nwhere\n\t" + conds
		}
	} else {
		for t := range conditions {
			log.Debugf("Conditions in WHERE:%v: %v\n", t, conditions[t])
		}
		conds = conditions[1]
		conds = string(firstline.ReplaceAll([]byte(conds), []byte("")))
		conds = string(fmtwhere.ReplaceAll([]byte(conds), []byte("\n\t$1")))
		returnquery = returnquery + "\nwhere\n\t" + conds
	}

	return returnquery
}

func FormatCTE(query string) string {
	withre := regexp.MustCompile(`(?i)^(with\s[a-zA-Z0-9_]*\sas\s\()(.*?)(\))`)
	nextctere := regexp.MustCompile(`(?i)(.*)(\),\s[a-zA-Z0-9_]*\sas\s\()(.*?)(\))`)
	cteselect := regexp.MustCompile(`(?i)(\))\s(.*)`)

	// log.Debugf("CTE Query:\n%v\n\n", query)
	query = ReplaceBrackets(query)

	// match := nextctere.Match([]byte(query))
	// log.Info(match)

	returnquery := string(withre.ReplaceAll([]byte(query), []byte("$1\n\t$2\n$3")))
	// log.Debugf("Formatted with:\n%v\n", returnquery)

	returnquery = string(nextctere.ReplaceAll([]byte(returnquery), []byte("$1$2\n\t$3\n$4")))

	// log.Debugf("Formatted with:\n%v\n", returnquery)

	selectstmt := cteselect.FindStringSubmatch(returnquery)[2]
	fmtselectstmt := FormatSelect(selectstmt)
	log.Debugf("SELECT stmt in CTE:\n%v\n", fmtselectstmt)

	returnquery = strings.ReplaceAll(returnquery, selectstmt, "")
	log.Debugf("Query before sending for UNWRAP:\n%v\n", returnquery)

	returnquery = unwrapbrackets(returnquery)
	returnquery = returnquery + fmtselectstmt

	return returnquery
}

func unwrapbrackets(query string) string {
	log.Debugf("Unwrap func\nQuery:\n%v\n\n", query)
	returnquery := query
	match := true

	for {
		if !match {
			break
		}

		for k, v := range mapper {
			// log.Debugf("key:\t%v\nvalue:\t%v\n", k, v)
			match = strings.Contains(returnquery, k)

			if match {
				log.Debugf("Matched %v in %v...\n", k, CleanQuery(returnquery))
				q := matcher(CleanQuery(v))
				log.Debugf("Formatted query in unwrap func:\n%v\n\n", q)

				q = unwrapbrackets(q)
				// match = false
				// break
				returnquery = strings.ReplaceAll(returnquery, k, q)
				log.Debugf("Returned query after final unwrap\n%v\n", returnquery)
			}
		}
	}

	return returnquery
}
