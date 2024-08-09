package parallel

import (
	"sync"
	"github.com/aleedurrani/TimeComplexity/pkg/helperFunctions"
	"github.com/aleedurrani/TimeComplexity/pkg/fileHandling"
)

// This function uses goroutines to process the file in parallel

// ParallelCountAll counts words, punctuation, vowels, sentences, paragraphs, and digits using goroutines
func ParallelCountAll() (helperFunctions.Counts) {

	file := fileHandling.OpenFile()
	defer file.Close()
	scanner := fileHandling.CreateRuneScanner(file)

	var wg sync.WaitGroup

	countChannels := helperFunctions.CountChannels{
		WordChan:      make(chan int),
		PunctChan:     make(chan int),
		VowelChan:     make(chan int),
		SentenceChan:  make(chan int),
		ParagraphChan: make(chan int),
		DigitChan:     make(chan int),
	}

	chunkSize := 1000000 
	chunk := make([]byte, 0, chunkSize)

	// processChunk processes a chunk of the file and sends results through channels
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

		// Send results through channels
		countChannels.WordChan <- counts.Word
		countChannels.PunctChan <- counts.Punct
		countChannels.VowelChan <- counts.Vowel
		countChannels.SentenceChan <- counts.Sentence
		countChannels.ParagraphChan <- counts.Paragraph
		countChannels.DigitChan <- counts.Digit
	}

	chunkCount := 0
	// Read the file in chunks and process each chunk in a separate goroutine
	for scanner.Scan() {
		chunk = append(chunk, scanner.Text()[0])
		if len(chunk) == chunkSize {
			wg.Add(1)
			go processChunk(chunk, false)
			chunk = make([]byte, 0, chunkSize)
			chunkCount++
		}
	}

	// Process the last chunk if it's not empty
	if len(chunk) > 0 {
		wg.Add(1)
		go processChunk(chunk, true)
		chunkCount++
	}

	// Close channels when all goroutines are done
	go func() {
		wg.Wait()
		close(countChannels.WordChan)
		close(countChannels.PunctChan)
		close(countChannels.VowelChan)
		close(countChannels.SentenceChan)
		close(countChannels.ParagraphChan)
		close(countChannels.DigitChan)
	}()

	// Sum up the results from all goroutines
	counts := helperFunctions.Counts{}

	for i := 0; i < chunkCount; i++ {
		counts.Word += <-countChannels.WordChan
		counts.Punct += <-countChannels.PunctChan
		counts.Vowel += <-countChannels.VowelChan
		counts.Sentence += <-countChannels.SentenceChan
		counts.Paragraph += <-countChannels.ParagraphChan
		counts.Digit += <-countChannels.DigitChan
	}

	return counts
}

