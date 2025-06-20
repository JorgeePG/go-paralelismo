package counter

import (
	"regexp"
	"strings"
	"sync"
)

type WordCounter struct {
	wordRegex *regexp.Regexp
	numCPUs   int
}

func NewWordCounter(numCPUs int) *WordCounter {
	regex := regexp.MustCompile(`^[A-Za-zÁÉÍÓÚÑáéíóúñÀ-ÿ][A-Za-zÁÉÍÓÚÑáéíóúñÀ-ÿ0-9\-']*[A-Za-zÁÉÍÓÚÑáéíóúñÀ-ÿ0-9]$|^[A-Za-zÁÉÍÓÚÑáéíóúñÀ-ÿ]$`)
	return &WordCounter{
		wordRegex: regex,
		numCPUs:   numCPUs,
	}
}

func (wc *WordCounter) isValidWord(word string) bool {
	return wc.wordRegex.MatchString(word)
}

// CountWords cuenta el número total de palabras en un texto utilizando el mapa de frecuencias
func (wc *WordCounter) CountWords(input string) int {
	// Reutilizamos el mapa de frecuencias que ya calculamos para evitar procesar el texto dos veces
	wordFreq := wc.CountWordFrequency(input)

	// Contar el número total de palabras sumando todas las frecuencias
	totalCount := 0
	for _, freq := range wordFreq {
		totalCount += freq
	}

	return totalCount
}

// cleanWords limpia y preprocesa todas las palabras del texto en paralelo
func (wc *WordCounter) cleanWords(words []string) []string {
	punctRegex := regexp.MustCompile(`^[^A-Za-zÁÉÍÓÚÑáéíóúñÀ-ÿ0-9]+|[^A-Za-zÁÉÍÓÚÑáéíóúñÀ-ÿ0-9]+$`)

	// Si hay pocas palabras o solo un CPU disponible, usar el método secuencial
	if len(words) < 1000 || wc.numCPUs <= 1 {
		return wc.cleanWordsSequential(words, punctRegex)
	}

	var wg sync.WaitGroup
	numWorkers := wc.numCPUs
	chunkSize := (len(words) + numWorkers - 1) / numWorkers

	// Crear un slice de slices para almacenar los resultados de cada worker
	results := make([][]string, numWorkers)

	// Lanzar workers para procesar cada chunk
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		startIdx := i * chunkSize
		endIdx := (i + 1) * chunkSize
		if endIdx > len(words) {
			endIdx = len(words)
		}

		go func(start, end, workerID int) {
			defer wg.Done()

			// Preasignar capacidad para el resultado de este worker
			localCleanedWords := make([]string, 0, end-start)

			// Procesar palabras en el rango asignado
			for j := start; j < end; j++ {
				// Limpiar signos de puntuación
				cleanWord := punctRegex.ReplaceAllString(words[j], "")

				// Verificar si la palabra limpia es válida
				if cleanWord != "" && wc.isValidWord(cleanWord) {
					localCleanedWords = append(localCleanedWords, cleanWord)
				}
			}

			// Guardar el resultado de este worker
			results[workerID] = localCleanedWords
		}(startIdx, endIdx, i)
	}

	// Esperar a que todos los workers terminen
	wg.Wait()

	// Calcular el tamaño total del resultado final
	totalSize := 0
	for _, workerResult := range results {
		totalSize += len(workerResult)
	}

	// Consolidar los resultados en un único slice
	cleanedWords := make([]string, 0, totalSize)
	for _, workerResult := range results {
		cleanedWords = append(cleanedWords, workerResult...)
	}

	return cleanedWords
}

// Versión secuencial de cleanWords para usar cuando la paralelización no es conveniente
func (wc *WordCounter) cleanWordsSequential(words []string, punctRegex *regexp.Regexp) []string {
	cleanedWords := make([]string, 0, len(words))

	for _, word := range words {
		// Limpiar signos de puntuación al inicio y final de la palabra
		cleanWord := punctRegex.ReplaceAllString(word, "")

		// Verificar si la palabra limpia es válida
		if cleanWord != "" && wc.isValidWord(cleanWord) {
			cleanedWords = append(cleanedWords, cleanWord)
		}
	}

	return cleanedWords
}

func (wc *WordCounter) CountWordFrequency(input string) map[string]int {
	// Dividir el texto en palabras
	words := strings.Fields(input)

	// Limpiar las palabras
	words = wc.cleanWords(words)

	// Determinar el número de goroutines a utilizar
	numWorkers := wc.numCPUs

	// Optimizar el número de workers
	if numWorkers <= 0 {
		numWorkers = 1
	}

	var wg sync.WaitGroup
	// Buffer del canal para evitar bloqueos
	wordFreqChan := make(chan map[string]int, numWorkers)

	// Dividir las palabras en chunks para cada worker - asegurarnos de distribuir equitativamente
	chunkSize := (len(words) + numWorkers - 1) / numWorkers

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		startIndex := i * chunkSize
		endIndex := (i + 1) * chunkSize
		if endIndex > len(words) {
			endIndex = len(words)
		}

		// Cada goroutine procesa su propio chunk de palabras
		go func(start, end int) {
			defer wg.Done()

			// Preasignamos una capacidad estimada para mejorar rendimiento
			localFreq := make(map[string]int, (end-start)/2)
			for j := start; j < end; j++ {
				localFreq[words[j]]++
			}

			wordFreqChan <- localFreq
		}(startIndex, endIndex)
	}

	// Cerrar el canal cuando todas las goroutines terminen
	go func() {
		wg.Wait()
		close(wordFreqChan)
	}()

	// Consolidar todos los conteos en un solo mapa
	// Preasignamos una capacidad estimada para el mapa final
	finalWordFreq := make(map[string]int, len(words)/2)
	for localFreq := range wordFreqChan {
		for word, count := range localFreq {
			finalWordFreq[word] += count
		}
	}

	return finalWordFreq
}
