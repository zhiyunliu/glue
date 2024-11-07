package xdb

import (
	"fmt"
	"regexp"
	"testing"
	"time"
)

var patterns = []string{
	`^abc`,
	`def$`,
	`ghi`,
	`jkl`,
}

var combinedPattern = `^(abc|def$|ghi|jkl)`

func TestSinglePattern(t *testing.T) {
	re := regexp.MustCompile(combinedPattern)
	start := time.Now()
	for i := 0; i < 100000000; i++ {
		re.MatchString("abcdef")
	}
	elapsed := time.Since(start)
	fmt.Printf("Single pattern took %s\n", elapsed)
}

func TestMultiplePatterns(t *testing.T) {
	ress := make([]*regexp.Regexp, len(patterns))
	for i, pattern := range patterns {
		ress[i] = regexp.MustCompile(pattern)
	}

	start := time.Now()
	for i := 0; i < 100000000; i++ {
		for _, re := range ress {
			re.MatchString("abcdef")
		}
	}
	elapsed := time.Since(start)
	fmt.Printf("Multiple patterns took %s\n", elapsed)
}
