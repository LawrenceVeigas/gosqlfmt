package gosqlfmt

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"
)

// map of everything within ()
var mapper = make(map[string]string)

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

	// remove leading/trailing whitespace
	query = strings.TrimSpace(query)

	return query
}

func ReplaceBrackets(query string) string {
	var openBracketIndex []int
	var tempCloseBracketIndex []int

	openBracketIndices := regexp.MustCompile(`\(`).FindAllStringIndex(query,-1)
	closeBracketIndices := regexp.MustCompile(`\)`).FindAllStringIndex(query,-1)

	for i:= range openBracketIndices {
		openBracketIndex = append(openBracketIndex, openBracketIndices[i][1])
	}
	for i:= range closeBracketIndices {
		tempCloseBracketIndex = append(tempCloseBracketIndex, closeBracketIndices[i][0])
	}

	closeBracketIndex  := make([]int, len(tempCloseBracketIndex))
	unused := make([]int, len(tempCloseBracketIndex))
	// arrange bracket indices
	for i:= range openBracketIndex {
		j := i+1
		if j<=len(openBracketIndex)-1 {
			for k := range tempCloseBracketIndex {
				v := tempCloseBracketIndex[k]
				if v>openBracketIndex[i] && v<openBracketIndex[j] {
					closeBracketIndex[i] = v
					break
				}
			}

		} else {
			// find unused indices
			for n:= range tempCloseBracketIndex {
				v := tempCloseBracketIndex[n]
				flag := 1
				for o:= range closeBracketIndex {
					v2 := closeBracketIndex[o]
					if v==v2 {
						flag = 0
					}
				}
				if flag==1 {
					unused[n] = v
				}
			}



			// fill nil/0 positions in reverse order
			for l := len(tempCloseBracketIndex)-1; i>=0 ; i--  {
				v := unused[l]

				if v!=0{
					for m:= range closeBracketIndex {
						x := closeBracketIndex[m]
						if x==0 {
							closeBracketIndex[m] = v
						}
					}
				}

			}
		}

	}

	swapopen := make([]int, len(openBracketIndex))
	copy(swapopen, openBracketIndex)
	swapclose := make([]int, len(closeBracketIndex))
	sortindex := make([]int, len(swapopen))

	sort.Sort(sort.Reverse(sort.IntSlice(swapopen)))
	
	// Order brackets such that inmost bracket is picked first
	for i:= range openBracketIndex {
		for j:= range swapopen {
			// fmt.Printf("Open - %v: %v\n",i, openBracketIndex[i])
			// fmt.Printf("Sort - %v: %v\n",j, swapopen[j])
			if openBracketIndex[i]==swapopen[j] {
				sortindex[i] = j
			}
		}
	}
	for i:= range sortindex {
		swapclose[i] = closeBracketIndex[sortindex[i]]
	}

	returnquery := query

	for i:=range swapopen {
		text := query[swapopen[i]:swapclose[i]]
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
	query = CleanQuery(query)

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
			for t := range where_li {
				fmt.Printf("%v: %v\n", t, where_li[t])
			}

			conds = where_li[1]
			conds = string(firstline.ReplaceAll([]byte(conds), []byte("")))
			conds = string(fmtwhere.ReplaceAll([]byte(conds), []byte("\n\t$1")))
			returnquery = returnquery + "\nwhere\n\t" + conds
		}
	} else {
		for t := range conditions {
			fmt.Printf("%v: %v\n", t, conditions[t])
		}
		conds = conditions[1]
		conds = string(firstline.ReplaceAll([]byte(conds), []byte("")))
		conds = string(fmtwhere.ReplaceAll([]byte(conds), []byte("\n\t$1")))
		returnquery = returnquery + "\nwhere\n\t" + conds
	}

	return returnquery 
}

func FormatCTE(query string) string {
	ctes := regexp.MustCompile(`\s[a-zA-Z0-9_]+\sas\s\((.*?)\)(,|\sselect)`)
	cteselect := regexp.MustCompile(`\)\s(select.*)`)

	selectsmt := cteselect.FindString(query)
	selectsmt = strings.Replace(selectsmt, ") ", "", -1)

	texts := ctes.FindAllStringSubmatch(query,-1)
	fmttexts := make([]string, len(texts))
	for i := range texts {
		fmttexts[i] = ReplaceBrackets(texts[i][1])
		fmt.Printf("Actual: %v\n\n", texts[i][1])
		fmt.Printf("Format: %v\n\n", fmttexts[i])
	}

	return query
}
