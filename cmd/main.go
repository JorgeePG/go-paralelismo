package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"time"

	"word-counter/internal/counter"
	"word-counter/internal/file"
)

func main() {
	filePath := "example.txt"
	fileHandler := file.FileHandler{}

	// Crear archivo CSV para guardar los resultados
	csvFile, err := os.Create("resultados_rendimiento.csv")
	if err != nil {
		log.Fatalf("Error al crear archivo CSV: %v", err)
	}
	defer csvFile.Close()

	// Crear escritor CSV
	csvWriter := csv.NewWriter(csvFile)
	defer csvWriter.Flush()

	// Escribir encabezados
	if err := csvWriter.Write([]string{"NumCPUs", "TiempoNanosegundos"}); err != nil {
		log.Fatalf("Error al escribir encabezados CSV: %v", err)
	}

	content, err := fileHandler.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	// Variables para guardar el mejor tiempo
	mejorTiempo := time.Duration(1<<63 - 1) // Valor máximo para time.Duration
	mejorCPUs := 0
	var mejorWordFreq map[string]int
	var mejorTotal int
	var mejorPalabra string
	var mejorCountMax int

	for numCPUsToUse := 1; numCPUsToUse <= runtime.NumCPU(); numCPUsToUse++ {
		inicio := time.Now()
		wordCounter := counter.NewWordCounter(numCPUsToUse)
		wordFreq := wordCounter.CountWordFrequency(content)
		tiempoTotal := time.Since(inicio)
		total := wordCounter.CountWords(content)

		// Imprimir solo número de CPUs y tiempo para todas las pruebas
		fmt.Printf("Utilizando %d CPUs - Tiempo: %v\n", numCPUsToUse, tiempoTotal)

		// Guardar resultado en el CSV con tiempo en nanosegundos (valor numérico puro)
		tiempoNanos := tiempoTotal.Nanoseconds()
		if err := csvWriter.Write([]string{strconv.Itoa(numCPUsToUse), strconv.FormatInt(tiempoNanos, 10)}); err != nil {
			log.Printf("Error al escribir en CSV: %v", err)
		}

		// Guardar los resultados si este tiempo es mejor que el anterior
		if tiempoTotal < mejorTiempo {
			mejorTiempo = tiempoTotal
			mejorCPUs = numCPUsToUse
			mejorWordFreq = wordFreq
			mejorTotal = total

			// Encontrar la palabra que más aparece
			mejorPalabra, mejorCountMax = "", 0
			for word, count := range wordFreq {
				if mejorCountMax < count {
					mejorPalabra, mejorCountMax = word, count
				}
			}
		}
	}

	// Mostrar los resultados detallados solo para la ejecución con mejor tiempo
	fmt.Println()
	fmt.Println("==================================")
	fmt.Println("RESULTADOS DEL MEJOR TIEMPO")
	fmt.Println("==================================")
	fmt.Printf("Mejor tiempo conseguido utilizando %d CPUs: %v\n", mejorCPUs, mejorTiempo)
	fmt.Println("==================================")

	if len(mejorWordFreq) == 0 {
		fmt.Println("No se encontraron palabras válidas en el texto.")
	} else {
		fmt.Printf("Palabra que más aparece: %-20s Frecuencia: %d\n", mejorPalabra, mejorCountMax)
		fmt.Println("==================================")
		fmt.Printf("Total de palabras: %d\n", mejorTotal)
	}
	fmt.Println("==================================")
	fmt.Printf("Los resultados de rendimiento han sido guardados en 'resultados_rendimiento.csv'\n")
}
