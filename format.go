package gosqlfmt

import (
	"os"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		// TimestampFormat: "2006-01-02 15:04:05",
		TimestampFormat: "15:04:05",
	})

	log.SetOutput(os.Stdout)

	log.SetLevel(log.DebugLevel)
	log.SetReportCaller(true)
}

func matcher(query string) string {
	match, err := regexp.Match(`(?i)^set`, []byte(query))
	if err != nil {
		log.Error(err)
	}
	if match {
		log.Info("Matched SET")
		return query
	}

	match, err = regexp.Match(`(?i)^select`, []byte(query))
	if err != nil {
		log.Error(err)
	}
	if match {
		log.Info("Matched SELECT")
		return FormatSelect(query)
	}

	match, err = regexp.Match(`(?i)^with`, []byte(query))
	if err != nil {
		log.Error(err)
	}
	if match {
		log.Info("Matched CTE")
		return FormatCTE(query)
	}

	return query
}

func FormatQuery(fileName string) string {

	queries := GetQuery(fileName)

	var finalQuery string
	for i := range queries {
		query := matcher(CleanQuery(queries[i]))

		if len(finalQuery) > 0 {
			finalQuery = finalQuery + ";\n\n" + strings.TrimSpace(query)
		} else {
			finalQuery = strings.TrimSpace(query)
		}
	}
	return finalQuery
}
