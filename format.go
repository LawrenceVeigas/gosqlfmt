package gosqlfmt

func FormatQuery(fileName string) string {

	queries := GetQuery(fileName)
	
	query := queries[2]

	return FormatSet(query)
}
