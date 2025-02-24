/*
@Author: David Majomi
@Class: Distributed Systems (Rutgers University)
@Date : 02/20/2025

--------------------------------------------------------------------------------
Objective [@Author Michael Palis]:												|
	Given 10 csv files, filter out cities that have a population of at least	|
	min_pop, which is inputted on the command line. 							|
	Along with min_pop, the command line also takes the directory of the csv 	|
	files as argument. 															|
																				|
	For example:																|
	> go run Multithreaded_Directory_Filter.go us-cities 1000000				|
																				|
	Requirements:																|
	1) Worker Pool																|
	2) Map function																|
	3) Reduce function															|
--------------------------------------------------------------------------------

*/

package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
)

const Nworkers = 3

// BufferSizeResults defines the buffer size for the results channel
const BufferSizeResults = 500

// BufferSizeWork defines the buffer size for individual work result channels
const BufferSizeWork = 50

// Work represents a task to be processed by a worker
type Work struct {
	file   string
	minPop int
	result chan Result
}

// Result represents a city that meets the population criteria
type Result struct {
	cityName   string
	state      string
	population int
}

// mapFunction processes files from the jobs channel, filtering cities by population
// Each worker reads a CSV file, extracts city data, and sends cities meeting
// the minimum population threshold through the result channel
func mapFunction(jobs chan Work) {

	for work := range jobs {
		f, err := os.Open(work.file)
		if err != nil {
			fmt.Printf("Error opening file %s: %v\n", work.file, err)
			continue
		}

		csvReader := csv.NewReader(f)

		for {
			var tempPopulation int
			rec, err := csvReader.Read()
			if err == io.EOF {
				// println("EOF")
				break
			}
			if err != nil {
				fmt.Printf("Error reading record from %s: %v\n", work.file, err)
				continue
			}

			tempPopulation, err = strconv.Atoi(rec[2])

			if err != nil {
				panic(err)
			}

			if tempPopulation >= work.minPop {
				work.result <- Result{
					cityName:   rec[0],
					state:      rec[1],
					population: tempPopulation,
				}
			}

		}

		close(work.result)
		f.Close()
	}
}

// maxPopulation finds the maximum population in a slice of Result
// Used for sorting states with equal numbers of qualifying cities
func maxPopulation(cities []Result) int {
	maxPop := 0
	for _, city := range cities {
		if city.population > maxPop {
			maxPop = city.population
		}
	}
	return maxPop
}

// reduce aggregates results from all workers, organizing cities by state
// It receives channels of Results from the workers, aggregates them by state,
// sorts states by number of cities (and max population as tiebreaker),
// and formats the output according to assignment requirements
func reduce(allResults chan chan Result) {
	var stateData map[string][]Result
	stateData = make(map[string][]Result)

	// Collect all results from worker channels
	for resultCh := range allResults {
		for result := range resultCh {

			_, isMapContainsKey := stateData[result.state]
			if isMapContainsKey {
				stateData[result.state] = append(stateData[result.state], result)

			} else {

				listOfCityData := []Result{result}
				stateData[result.state] = listOfCityData
			}
		}
	}

	// Create a slice to hold the keys for sorting
	var sortedStates []string
	for state := range stateData {
		sortedStates = append(sortedStates, state)
	}

	// Sort the keys based on the length of their corresponding slices in decreasing order
	// to display the states ordered by number of cities and max city population when they have the same number of citites
	sort.Slice(sortedStates, func(i, j int) bool {
		if len(stateData[sortedStates[i]]) != len(stateData[sortedStates[j]]) {
			return len(stateData[sortedStates[i]]) > len(stateData[sortedStates[j]])
		} else {
			maxPopI := maxPopulation(stateData[sortedStates[i]])
			maxPopJ := maxPopulation(stateData[sortedStates[j]])

			return maxPopI > maxPopJ
		}
	})

	// Print the state data
	for _, state := range sortedStates {
		fmt.Printf("%s: %d\n", state, len(stateData[state]))
		for _, cityData := range stateData[state] {
			fmt.Printf("- %s, %d\n", cityData.cityName, cityData.population)
		}
	}
}

func main() {

	// Parse command-line arguments
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run pop.go <directory> <min_population>")
		os.Exit(1)
	}

	var dir = os.Args[1]
	var minPop, err = strconv.Atoi(os.Args[2])

	if err != nil {
		fmt.Println("Bad command line arguments")
		panic(err)
	}

	// Create worker pool
	jobs := make(chan Work)
	wg := sync.WaitGroup{}
	for i := 0; i < Nworkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mapFunction(jobs)
		}()
	}

	// Channel for collecting results from all workers

	allResults := make(chan chan Result, BufferSizeResults) // Use buffered Channel

	// Go routine to read the directory
	go func() {
		defer close(jobs)
		defer close(allResults)

		filepath.Walk(dir, func(path string, d fs.FileInfo, err error) error {
			if err != nil {
				println(err)
				return err
			}
			if !d.IsDir() {
				ch := make(chan Result, BufferSizeWork)
				jobs <- Work{file: path, minPop: minPop, result: ch}
				allResults <- ch
			}
			return nil
		})
	}()

	// Start reduction process in separate goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		reduce(allResults)
	}()

	wg.Wait()
}
