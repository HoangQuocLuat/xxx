package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"ai-monitor/groq"
	"ai-monitor/parser"
	"ai-monitor/tracer"
	"ai-monitor/watcher"
)

func main() {
	logDir := getEnv("LOG_PATH", "/shared/logs")
	phpSrc := getEnv("PHP_SRC", "/shared/src")

	fmt.Println("AI Monitor started...")
	fmt.Printf("Watching : %s\n", logDir)
	fmt.Printf("PHP src  : %s\n", phpSrc)

	// Watch file log hôm nay, tự cập nhật ngày mới
	for {
		logFile := fmt.Sprintf("%s/app-%s.log", logDir, time.Now().Format("2006-01-02"))

		// Chờ file tồn tại
		for {
			if _, err := os.Stat(logFile); err == nil {
				break
			}
			fmt.Printf("Waiting for log file: %s\n", logFile)
			time.Sleep(3 * time.Second)
		}

		fmt.Printf("Watching file: %s\n", logFile)

		err := watcher.Watch(logFile, func(el watcher.ErrorLine) {
			entry, err := parser.Parse(el.Raw)
			if err != nil {
				fmt.Printf("[PARSE ERROR] %v\n", err)
				return
			}

			fmt.Printf("\n[DETECTED] %s\n", entry.Message)

			snippet := tracer.Extract(phpSrc, entry.File, entry.Line)

			var extraSnippets []tracer.CodeSnippet
			for _, frame := range entry.Trace {
				if frame.File == "" { continue }
				s := tracer.Extract(phpSrc, frame.File, frame.Line)
				extraSnippets = append(extraSnippets, s)
				if len(extraSnippets) >= 3 { break }
			}

			solution, err := groq.Analyze(groq.AnalyzeRequest{
				ErrorMessage:  entry.Message,
				Exception:     entry.Exception,
				MainSnippet:   snippet,
				TraceSnippets: extraSnippets,
				URL:           entry.URL,
			})
			if err != nil {
				fmt.Printf("[GROQ ERROR] %v\n", err)
				return
			}

			printSolution(entry, solution)
		})

		if err != nil {
			fmt.Printf("[WATCHER ERROR] %v — retrying in 5s\n", err)
			time.Sleep(5 * time.Second)
		}
	}
}

func printSolution(entry *parser.LogEntry, s *groq.Solution) {
	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf("ROOT CAUSE : %s\n", s.RootCause)
	fmt.Printf("FILE       : %s (line %d)\n", entry.File, entry.Line)
	fmt.Printf("SOLUTION   : %s\n", s.Solution)
	if s.CodeFix != "" {
		fmt.Printf("FIX        :\n%s\n", s.CodeFix)
	}
	fmt.Println(strings.Repeat("─", 60))
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}