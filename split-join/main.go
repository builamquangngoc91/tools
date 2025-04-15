package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func parseSize(sizeStr string) (int64, error) {
	suffixes := map[string]int64{"KB": 1024, "MB": 1024 * 1024, "GB": 1024 * 1024 * 1024}
	for suffix, factor := range suffixes {
		if strings.HasSuffix(sizeStr, suffix) {
			num, err := strconv.ParseInt(strings.TrimSuffix(sizeStr, suffix), 10, 64)
			return num * factor, err
		}
	}
	return strconv.ParseInt(sizeStr, 10, 64) // bytes
}

func splitFile(inputPath string, maxSize int64) error {
	file, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	buf := make([]byte, maxSize)
	part := 0

	for {
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		partFile := fmt.Sprintf("%s.part%d", inputPath, part)
		err = os.WriteFile(partFile, buf[:n], 0644)
		if err != nil {
			return err
		}

		fmt.Println("Written", partFile)
		part++
	}

	return nil
}

func joinFiles(firstPart string) error {
	base := strings.TrimSuffix(firstPart, filepath.Ext(firstPart)) // remove `.partN`
	outFile, err := os.Create(base)
	if err != nil {
		return err
	}
	defer outFile.Close()

	part := 0
	for {
		partFile := fmt.Sprintf("%s.part%d", base, part)
		data, err := os.ReadFile(partFile)
		if err != nil {
			if os.IsNotExist(err) {
				break // No more parts
			}
			return err
		}

		_, err = outFile.Write(data)
		if err != nil {
			return err
		}
		fmt.Println("Appended", partFile)
		part++
	}

	fmt.Println("Joined to", base)
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  split --input <file> --size 5MB")
		fmt.Println("  join  --input <first-part>")
		return
	}

	action := os.Args[1]
	fs := flag.NewFlagSet(action, flag.ExitOnError)

	input := fs.String("input", "", "input file path")
	size := fs.String("size", "5MB", "max size per part (for split only)")
	fs.Parse(os.Args[2:])

	if *input == "" {
		fmt.Println("Error: input file is required")
		return
	}

	switch action {
	case "split":
		maxSize, err := parseSize(*size)
		if err != nil || maxSize <= 0 {
			fmt.Println("Invalid size:", *size)
			return
		}
		if err := splitFile(*input, maxSize); err != nil {
			fmt.Println("Split failed:", err)
		}
	case "join":
		if err := joinFiles(*input); err != nil {
			fmt.Println("Join failed:", err)
		}
	default:
		fmt.Println("Unknown command:", action)
	}
}
