package test

import (
	"time"
)

func testDurationMultiplication() {
	// This should trigger durationcheck linter
	badDuration := time.Second * time.Minute  // This should be flagged
	
	// This should be fine - multiplying a number by a duration
	goodDuration := 5 * time.Second
	
	time.Sleep(badDuration)
	time.Sleep(goodDuration)
}