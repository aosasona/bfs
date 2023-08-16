package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type (
	FileType string

	File struct {
		Name string   `json:"name"`
		Path string   `json:"path"`
		Type FileType `json:"type"`
	}
)

const (
	reset  = "\033[0m"
	red    = "\033[31m"
	cyan   = "\033[36m"
	yellow = "\033[33m"
	blue   = "\033[34m"
)

const (
	TypeDirectory FileType = "directory"
	TypeFile      FileType = "file"
)

var (
	root, query string
	asJson      bool

	results = []File{}
	stream  = make(chan File)
	done    = make(chan bool)
)

func main() {
	flag.StringVar(&root, "root", ".", "Root directory to search")
	flag.StringVar(&query, "query", "", "Query to search for")
	flag.BoolVar(&asJson, "json", false, "Stream results as JSON")
	flag.Parse()

	loadOpts()

	start := time.Now()
	info(fmt.Sprintf("-> Searching for `%s` in `%s`", query, root))

	// get all paths in that directory
	paths, err := getTargets(root)
	if err != nil {
		exit(fmt.Sprintf("Error getting paths: %s", err.Error()))
	}

	// split the paths into chunks and search each chunk in a goroutine
	var (
		chunks = chunk(*paths)
		wg     sync.WaitGroup
	)

	process(&chunks, &wg)
	go handleStream()
	go func() {
		wg.Wait()
		defer close(stream)
	}()

	<-done

	success(fmt.Sprintf("-> Found %d results in %d ms", len(results), int(time.Since(start).Milliseconds())))
}

func process(chunks *[][]File, wg *sync.WaitGroup) {
	for _, chunk := range *chunks {
		wg.Add(1)
		go func(chunk []File) {
			defer wg.Done()
			search(chunk)
		}(chunk)
	}
}

func search(paths []File) {
	for _, path := range paths {
		if contains(path.Path, query) {
			stream <- path
		}

		if path.Type == TypeDirectory {
			subPaths, err := getTargets(path.Path)
			if err != nil {
				warn(fmt.Sprintf("Error getting subpaths for %s: %s", path.Path, err.Error()))
				continue
			}
			search(*subPaths)
		}
	}
}

func handleStream() {
	// read until the stream is closed
	for {
		res, ok := <-stream
		if !ok {
			break
		}
		results = append(results, res)
		if asJson {
			json, err := json.Marshal(res)
			if err != nil {
				warn(fmt.Sprintf("Error marshalling %s: %s", res.Path, err.Error()))
				continue
			}

			print(string(json))
		} else {
			print(fmt.Sprintf("[+] %s", res.Path))
		}
	}
	done <- true
}

func loadOpts() {
	home, err := os.UserHomeDir()
	if err != nil {
		exit(fmt.Sprintf("Error getting user home directory: %s\n", err.Error()))
	}

	if root == "" {
		root = home
	} else if !filepath.IsAbs(root) {
		root, err = filepath.Abs(root)
		if err != nil {
			exit(fmt.Sprintf("Error getting absolute path for %s: %s\n", root, err.Error()))
		}
	}

	if query == "" {
		query = flag.Arg(0)
	}

	if query == "" {
		exit("No query provided")
	}
}

func getTargets(root string) (*[]File, error) {
	if isFile(root) {
		exit("Root is a file, not a directory")
	}

	targets := make([]File, 0)

	files, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		t := TypeFile
		if file.IsDir() {
			t = TypeDirectory
		}

		targets = append(targets, File{Name: file.Name(), Path: filepath.Join(root, file.Name()), Type: t})
	}

	return &targets, nil
}

// Did I need a generic chunk function? No. Did I write one anyway? Yes.
func chunk[T any](arr []T) [][]T {
	chunks := make([][]T, 0)
	if len(arr) == 0 {
		return chunks
	}

	noOfChunks := randInt(int(math.Ceil(float64(len(arr) / 2))))
	size := int(math.Ceil(float64(len(arr) / noOfChunks)))

	for i := 0; i < len(arr); i += size {
		end := i + size

		if end > len(arr) {
			end = len(arr)
		}

		chunks = append(chunks, arr[i:end])
	}

	return chunks
}

func randInt(max int) int {
	var n int
	for {
		n = rand.Intn(max)
		if n != 0 {
			break
		}
	}

	return n
}

func contains(path string, query string) bool {
	return strings.Contains(path, query)
}

func isFile(path string) bool {
	f, err := os.Stat(path)
	return err == nil && !f.IsDir()
}

func exit(msg string) {
	fmt.Printf("%s%s%s\n", red, msg, reset)
	os.Exit(1)
}

func print(msg string) {
	fmt.Printf("%s\n", msg)
}

func success(msg string) {
	fmt.Printf("%s%s%s\n", cyan, msg, reset)
}

func warn(msg string) {
	fmt.Printf("%s%s%s\n", yellow, msg, reset)
}

func info(msg string) {
	fmt.Printf("%s%s%s\n", blue, msg, reset)
}
