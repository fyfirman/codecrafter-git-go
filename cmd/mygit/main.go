package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: mygit <command> [<args>...]\n")
		os.Exit(1)
	}
	//
	switch command := os.Args[1]; command {
	case "init":
		for _, dir := range []string{".git", ".git/objects", ".git/refs"} {
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
			}
		}

		headFileContents := []byte("ref: refs/heads/main\n")
		if err := os.WriteFile(".git/HEAD", headFileContents, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %s\n", err)
		}

		fmt.Println("Initialized git directory")

	case "cat-file":
		if len(os.Args) != 4 || os.Args[2] != "-p" {
			fmt.Fprintf(os.Stderr, "usage: mygit cat-file -p [path-file]\n")
			os.Exit(1)
		}

		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot get current work directory %s\n", err)
			os.Exit(1)
		}

		path := cwd + "/.git/objects/" + os.Args[3][0:2] + "/" + os.Args[3][2:]

		b, err := ioutil.ReadFile(path)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file %s\n", err)
			os.Exit(1)
		}

		r, err := zlib.NewReader(bytes.NewReader(b))

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading uncompressed data %s\n", err)
			os.Exit(1)
		}
		defer r.Close()

		decompressedData, err := ioutil.ReadAll(r)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading decompressed data %s\n", err)
		}

		content := decompressedData[strings.IndexByte(string(decompressedData), 0)+1:]
		fmt.Printf("%s", content)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}
