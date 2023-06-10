package search

import (
	"fmt"
	"path/filepath"
	"runtime"
)

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
)

/* Implements Trie */
type DataStore struct {
	isCompleteWord bool
	data           string
	childern       map[rune]*DataStore
}

type SearchResult struct {
	Found           bool
	Recommendations []string
}

func NewDataStore() *DataStore {
	return &DataStore{
		isCompleteWord: false,
		childern:       make(map[rune]*DataStore, 128),
	}
}

// Parses the dataset and initializes the datastore
func (node *DataStore) InitializeDataStore() {
	dataSet := parseDataset(filepath.Join(basepath, "/dataset.csv"))
	for _, data := range dataSet {
		node.Insert(data)
	}
}

// Inserts data into the tree
func (node *DataStore) Insert(data string) {

	var currentNode *DataStore = node

	for index, char := range data {
		if _, ok := currentNode.childern[char]; ok {
			currentNode = currentNode.childern[char]
			continue
		}

		if index+1 == len(data) {
			currentNode.childern[char] = &DataStore{
				isCompleteWord: true,
				data:           data,
				childern:       make(map[rune]*DataStore),
			}

			break
		}

		currentNode.childern[char] = &DataStore{
			isCompleteWord: false,
			childern:       make(map[rune]*DataStore),
		}

		currentNode = currentNode.childern[char]
	}
}

// Searches for the data if not found returns recommendations
func (node *DataStore) Search(data string) SearchResult {
	var currentNode *DataStore = node

	fmt.Println("Search string", data)

	if len(data) == 0 {
		return SearchResult{Found: false, Recommendations: make([]string, 0, 0)}
	}

	for index, char := range data {

		if _, ok := currentNode.childern[char]; ok {

			if currentNode.childern[char].isCompleteWord && index+1 == len(data) {

				return SearchResult{Found: true, Recommendations: node.generateRecommendation(currentNode)}
			}

			currentNode = currentNode.childern[char]

			continue
		} else {
			return SearchResult{Found: false, Recommendations: node.generateRecommendation(currentNode)}
		}

	}

	return SearchResult{Found: false, Recommendations: node.generateRecommendation(currentNode)}
}

// TODO: imporve this algo this is a very basic solution
func (node *DataStore) generateRecommendation(baseNode *DataStore) []string {
	var followChars []rune = make([]rune, 0, 3)

	var maxRecommendations = 5

	var followerNode *DataStore

	var recommendations []string = make([]string, 0, 3)

	for i := 0; i < 128; i++ {
		if _, ok := baseNode.childern[rune(i)]; ok {
			if len(followChars) == maxRecommendations {
				break
			}

			followChars = append(followChars, rune(i))
		}
	}

	for _, followChar := range followChars {

		followerNode = baseNode.childern[followChar]

		if followerNode.isCompleteWord {
			recommendations = append(recommendations, followerNode.data)
			continue
		}

		for i := 0; i < 128; {
			if _, ok := followerNode.childern[rune(i)]; ok {

				if followerNode.childern[rune(i)].isCompleteWord {
					recommendations = append(recommendations, followerNode.childern[rune(i)].data)

					break
				}

				followerNode = followerNode.childern[rune(i)]
				i = 0
			}

			i++
		}

	}

	return recommendations
}
