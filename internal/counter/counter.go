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

	// Recolectamos todos los mapas locales primero
	var allMaps []map[string]int

	go func() {
		wg.Wait()
		close(wordFreqChan)
	}()

	// Recolectar todos los mapas en memoria
	for localFreq := range wordFreqChan {
		if len(localFreq) > 0 { // Solo agregamos mapas no vacíos
			allMaps = append(allMaps, localFreq)
		}
	}
	// Si no hay mapas o solo hay uno, no necesitamos paralelizar
	if len(allMaps) == 0 {
		return make(map[string]int)
	} else if len(allMaps) == 1 {
		return allMaps[0]
	}

	// Implementar una reducción paralela de tipo Map-Reduce
	return wc.parallelReduceMaps(allMaps)
}

func (wc *WordCounter) parallelReduceMaps(maps []map[string]int) map[string]int {
	numMaps := len(maps)
	if numMaps == 0 {
		return make(map[string]int)
	}

	// Si tenemos pocos mapas o la cantidad de CPUs es baja, usar reducción secuencial
	if numMaps <= 2 || wc.numCPUs <= 1 {
		return wc.sequentialReduceMaps(maps)
	}

	// Calcular el nivel óptimo de paralelización
	// Usamos un enfoque divide y vencerás
	numReducers := wc.numCPUs
	if numReducers > numMaps/2 {
		numReducers = numMaps / 2
		if numReducers < 1 {
			numReducers = 1
		}
	}

	var wg sync.WaitGroup
	resultChan := make(chan map[string]int, numReducers)

	// Dividir los mapas en chunks para cada reducer
	chunkSize := (numMaps + numReducers - 1) / numReducers

	for i := 0; i < numReducers; i++ {
		wg.Add(1)
		startIdx := i * chunkSize
		endIdx := (i + 1) * chunkSize
		if endIdx > numMaps {
			endIdx = numMaps
		}

		// No empezar una goroutine si solo hay un elemento
		if endIdx-startIdx <= 1 {
			if startIdx < numMaps {
				resultChan <- maps[startIdx]
			}
			wg.Done()
			continue
		}

		// Cada goroutine reduce un subconjunto de mapas
		go func(start, end int) {
			defer wg.Done()

			// Reducir el subconjunto de mapas asignados a esta goroutine
			subResult := wc.sequentialReduceMaps(maps[start:end])
			resultChan <- subResult
		}(startIdx, endIdx)
	}

	// Esperar a que todos los reductores terminen
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Combinar los resultados finales de forma secuencial
	// Esta fase debe ser más rápida porque estamos combinando
	// un número mucho menor de mapas (numReducers)
	var finalResults []map[string]int
	for result := range resultChan {
		finalResults = append(finalResults, result)
	}

	// Última reducción secuencial de los resultados intermedios
	return wc.sequentialReduceMaps(finalResults)
}

// sequentialReduceMaps combina varios mapas de forma secuencial
func (wc *WordCounter) sequentialReduceMaps(maps []map[string]int) map[string]int {
	if len(maps) == 0 {
		return make(map[string]int)
	}

	// Estimar el tamaño total del mapa
	totalSize := 0
	for _, m := range maps {
		totalSize += len(m)
	}

	// Inicializar con una capacidad aproximada
	result := make(map[string]int, totalSize/2)

	// Combinar todos los mapas
	for _, m := range maps {
		for word, count := range m {
			result[word] += count
		}
	}

	return result
}
