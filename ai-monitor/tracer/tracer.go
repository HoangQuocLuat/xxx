// tracer/tracer.go
package tracer

import (
    "bufio"
    "fmt"
    "os"
)

type CodeSnippet struct {
    File    string
    Lines   []string  // ±5 dòng xung quanh lỗi
    ErrLine int
}

// phpRoot là đường dẫn mount source PHP vào container Go
func Extract(phpRoot, filePath string, errLine int) CodeSnippet {
    fullPath := phpRoot + filePath
    f, err := os.Open(fullPath)
    if err != nil {
        return CodeSnippet{File: filePath, ErrLine: errLine}
    }
    defer f.Close()

    var all []string
    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        all = append(all, scanner.Text())
    }

    start := max(0, errLine-6)
    end   := min(len(all), errLine+5)

    var snippet []string
    for i := start; i < end; i++ {
        prefix := "   "
        if i+1 == errLine { prefix = ">>>" }
        snippet = append(snippet, fmt.Sprintf("%s %4d | %s", prefix, i+1, all[i]))
    }

    return CodeSnippet{File: filePath, Lines: snippet, ErrLine: errLine}
}