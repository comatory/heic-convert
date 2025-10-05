package main

import (
	"flag"
	"fmt"
	"image/jpeg"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"

	"github.com/gen2brain/heic"
)

func printUsage() {
	fmt.Println("Usage: [options] <input files or directory>")
	fmt.Println("Options:")
	fmt.Println("  -h            Display help")
	fmt.Println("  -o <path>     Output directory (default: current directory)")
	fmt.Println("  -v            Enable verbose logging")
	fmt.Println("  -q <quality>  JPEG quality (1-100, default: 100)")
}

func normalizeFileName(fileName string) string {
	if strings.HasSuffix(fileName, ".HEIC") {
		return strings.TrimSuffix(path.Base(fileName), ".HEIC") + ".jpg"
	}

	return strings.TrimSuffix(path.Base(fileName), ".heic") + ".jpg"
}

func filterHeicFiles(inPath []string, verbose bool) ([]string, error) {
	var heicFilePaths []string

	for _, maybeHeic := range inPath {
		fileInfo, err := os.Stat(maybeHeic)

		if err != nil {
			if verbose {
				fmt.Fprintf(os.Stderr, "accessing path %s: %v\n", maybeHeic, err)
			}
			return nil, err
		}

		if fileInfo.IsDir() {
			if verbose {
				fmt.Printf("This is a directory. Skipping: %s\n", maybeHeic)
			}

			continue
		}

		if strings.HasSuffix(strings.ToLower(fileInfo.Name()), ".heic") {
			if verbose {
				fmt.Printf("Found .heic file: %s\n", fileInfo.Name())
			}

			heicFilePaths = append(heicFilePaths, maybeHeic)
		}
	}

	return heicFilePaths, nil
}

func convert(filePath string, outPath string, quality int, verbose bool, errChan chan error) {
	if verbose {
		fmt.Printf("Converting file: %s\n", filePath)
	}

	var outputFilePath string

	if outPath == "." {
		outputFilePath = normalizeFileName(filePath)
	} else {
		outputFilePath = path.Clean(outPath) + string(os.PathSeparator) + normalizeFileName(filePath)
	}

	reader, err := os.Open(filePath)
	if err != nil {
		errChan <- fmt.Errorf("opening file %s: %w", filePath, err)
		return
	}
	defer func() { _ = reader.Close() }()

	img, err := heic.Decode(reader)

	if err != nil {
		errChan <- fmt.Errorf("decoding HEIC image %s: %w", filePath, err)
		return
	}

	outFile, err := os.Create(outputFilePath)

	if err != nil {
		errChan <- fmt.Errorf("creating output file %s: %w", outputFilePath, err)
		return
	}

	defer func() { _ = outFile.Close() }()

	if verbose {
		fmt.Printf("Writing to output file: %s with quality %d\n", outputFilePath, quality)
	}

	err = jpeg.Encode(outFile, img, &jpeg.Options{Quality: quality})

	if err != nil {
		errChan <- fmt.Errorf("encoding JPEG image %s: %w", outputFilePath, err)
		return
	}

	if verbose {
		fmt.Printf("Successfully converted %s to %s\n", filePath, outputFilePath)
	}
}

func ensureOutputDir(path string, verbose bool) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if verbose {
			fmt.Printf("Creating output directory: %s\n", path)
		}

		return os.MkdirAll(path, 0755)
	}

	return nil
}

func main() {
	var (
		help    bool
		outPath string
		verbose bool
		quality int
	)

	flag.BoolVar(&help, "h", false, "Display help")
	flag.StringVar(&outPath, "o", ".", "Output directory")
	flag.BoolVar(&verbose, "v", false, "Enable verbose logging")
	flag.IntVar(&quality, "q", 100, "JPEG quality (1-100)")
	flag.Parse()

	if flag.NArg() == 0 && !help {
		fmt.Fprintln(os.Stderr, "No input files or directory specified.")
		printUsage()
		os.Exit(1)
	}

	if help {
		printUsage()
		return
	}

	if len(outPath) > 0 && outPath != "." {
		if err := ensureOutputDir(outPath, verbose); err != nil {
			fmt.Fprintf(os.Stderr, "creating output directory: %v\n", err)
			os.Exit(1)
		}

		if verbose {
			fmt.Printf("Output directory set to: %s\n", outPath)
		}
	}

	files, err := filterHeicFiles(flag.Args(), verbose)

	if err != nil {
		fmt.Fprintf(os.Stderr, "filtering .heic files: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		if verbose {
			fmt.Println("No .heic files found to convert.")
		}
		os.Exit(0)
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(files))
	var errorCount int
	semaphore := make(chan struct{}, runtime.NumCPU()*2)

	for _, file := range files {
		wg.Add(1)
		go func(fileToConvert string) {
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			defer wg.Done()
			convert(fileToConvert, outPath, quality, verbose, errChan)
		}(file)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		errorCount++
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}

	if errorCount > 0 {
		fmt.Fprintf(os.Stderr, "Completed with %d errors.\n", errorCount)
		os.Exit(1)
	}
}
