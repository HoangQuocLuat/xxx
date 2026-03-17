// parser/parser.go
package parser

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

type LogEntry struct {
	Datetime  string
	Level     string
	Message   string
	File      string
	Line      int
	Exception string
	Trace     []TraceFrame
	URL       string
	Body      map[string]interface{}
}

type TraceFrame struct {
	File string `json:"file"`
	Line int    `json:"line"`
	Func string `json:"func"`
}

var re = regexp.MustCompile(`\[(.+?)\] \w+\.(\w+): (.+?) (\{.+\})`)

func Parse(raw string) (*LogEntry, error) {
	m := re.FindStringSubmatch(raw)
	if m == nil {
		return nil, fmt.Errorf("no match")
	}

	var ctx struct {
		Exception string                 `json:"exception"`
		File      string                 `json:"file"`
		Line      int                    `json:"line"`
		Trace     []TraceFrame           `json:"trace"`
		URL       string                 `json:"url"`
		Body      map[string]interface{} `json:"body"`
	}
	json.Unmarshal([]byte(m[4]), &ctx)

	return &LogEntry{
		Datetime:  m[1],
		Level:     m[2],
		Message:   strings.TrimSpace(m[3]),
		File:      ctx.File,
		Line:      ctx.Line,
		Exception: ctx.Exception,
		Trace:     ctx.Trace,
		URL:       ctx.URL,
		Body:      ctx.Body,
	}, nil
}
