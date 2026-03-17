// watcher/watcher.go
package watcher

import (
	"bufio"
	"encoding/json"
	"os"
	"regexp"
	"strings"

	"github.com/fsnotify/fsnotify"
)

type ErrorLine struct {
	Raw     string
	Message string
	Context map[string]interface{}
}

var logRegex = regexp.MustCompile(`\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\] \w+\.ERROR: (.+?) (\{.*\})$`)

func parseLine(raw string) ErrorLine {
	el := ErrorLine{Raw: raw}

	matches := logRegex.FindStringSubmatch(raw)
	if matches == nil {
		el.Message = raw
		return el
	}

	el.Message = strings.TrimSpace(matches[1])

	var ctx map[string]interface{}
	if err := json.Unmarshal([]byte(matches[2]), &ctx); err == nil {
		el.Context = ctx
	}

	return el
}

func Watch(logPath string, onError func(ErrorLine)) error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer w.Close()

	w.Add(logPath)

	f, err := os.Open(logPath)
	if err != nil {
		return err
	}
	defer f.Close()

	f.Seek(0, os.SEEK_END)
	scanner := bufio.NewScanner(f)

	for range w.Events {
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, ".ERROR:") {
				onError(parseLine(line))
			}
		}
	}

	return nil
}
