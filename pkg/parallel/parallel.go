package parallel

import (
	"sync"
	"github.com/aleedurrani/TimeComplexity/pkg/utils/helperFunctions"
	"github.com/aleedurrani/TimeComplexity/pkg/utils/fileHandling"
)

// This function uses goroutines to process the file in parallel
// ParallelCountAll counts words, punctuation, vowels, sentences, paragraphs, and digits using goroutines
func ParallelCountAll() (helperFunctions.Counts) {

	file := fileHandling.OpenFile()
	scanner := fileHandling.CreateRuneScanner(file)

	var wg sync.WaitGroup

	countChannel := make(chan helperFunctions.Counts)

	chunkSize := 1000000 
	chunk := make([]byte, chunkSize)

	// processChunk processes a chunk of the file and sends results through the channel
	processChunk := func(chunk []byte, isLastChunk bool) {
		defer wg.Done()
		counts := helperFunctions.Counts{}
		inWord := false
		prevChar := ""

		for _, char := range string(chunk) {
			helperFunctions.ProcessChar(byte(char), &inWord, &counts)
			prevChar = string(char)
		}
		// Only count the last word if it's the last chunk and we're in a word
		if  isLastChunk && inWord {
			counts.Word++
		}

		// Only count the last paragraph if it's the last chunk and the last character isn't a newline
		if isLastChunk && prevChar != "\n" {
			counts.Paragraph++
		}

		// Send results through the channel
		countChannel <- counts
	}

	chunkCount := 0
	// Read the file in chunks and process each chunk in a separate goroutine
	for scanner.Scan() {
		chunk = append(chunk, scanner.Text()[0])
		if len(chunk) == chunkSize {
			wg.Add(1)
			go processChunk(chunk, false)
			chunk = make([]byte, chunkSize)
			chunkCount++
		}
	}

	// Process the last chunk if it's not empty
	if len(chunk) > 0 {
		wg.Add(1)
		go processChunk(chunk, true)
		chunkCount++
	}

	// Close channel when all goroutines are done
	go func() {
		wg.Wait()
		close(countChannel)
	}()

	// Sum up the results from all goroutines
	totalCounts := helperFunctions.SumCounts(countChannel)

	return totalCounts
}

