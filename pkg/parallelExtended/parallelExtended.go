package parallelExtended

import (
	"log"
	"sync"
	"github.com/aleedurrani/TimeComplexity/pkg/helperFunctions"
	"github.com/aleedurrani/TimeComplexity/pkg/fileHandling"
)


// This function uses goroutines to process the file in parallel with chunking based on defined number of go routines
// The number of go routines is defined by the user
func ParallelCountAll() (helperFunctions.Counts) {

	file := fileHandling.OpenFile()
	defer file.Close()


	fileSize := fileHandling.GetFileSize(file)

	// Define number of routines
	numRoutines := 8
	chunkSize := fileSize / int64(numRoutines)

	var wg sync.WaitGroup
	countChannels := helperFunctions.CreateCountChannels(numRoutines)

	// processChunk processes a chunk of the file and sends results through channels
	processChunk := func(start, end int64) {
		defer wg.Done()
		counts := helperFunctions.Counts{}
		inWord := false
		prevChar := ""

		chunk := make([]byte, end-start)

		if err := fileHandling.ReadChunk(file, chunk, start); err != nil {
			log.Fatal(err)
		}

		for _, char := range string(chunk) {
			helperFunctions.ProcessChar(byte(char), &inWord, &counts)
			prevChar = string(char)
		}

		// Only count the last word if it's the last chunk and we're in a word
		if end >= fileSize && inWord {
			counts.Word++
		}

		// Only count the last paragraph if it's the last chunk and the last character isn't a newline
		if end >= fileSize && prevChar != "\n" {
			counts.Paragraph++
		}

		// Send results through channels
		helperFunctions.SendCounts(counts, countChannels)
	}

	// Start goroutines
	for i := 0; i < numRoutines; i++ {
		start := int64(i) * chunkSize
		end := start + chunkSize
		if i == numRoutines-1 {
			end = fileSize
		}
		wg.Add(1)
		go processChunk(start, end)
	}

	// Close channels when all goroutines are done
	go func() {
		wg.Wait()
		helperFunctions.CloseChannels(countChannels)
	}()


	return helperFunctions.SumCounts(countChannels, numRoutines)
}