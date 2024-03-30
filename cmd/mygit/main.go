package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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

		gitObjPath, err := GetGitObjectPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot get git object path %s\n", err)
			os.Exit(1)
		}

		path := gitObjPath + os.Args[3][0:2] + "/" + os.Args[3][2:]

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

	case "hash-object":
		if len(os.Args) != 4 || os.Args[2] != "-w" {
			fmt.Fprintf(os.Stderr, "usage: mygit hash-object -w [path-file]\n")
			os.Exit(1)
		}

		inputPath := os.Args[3]

		filename := filepath.Base(inputPath)

		hasher := sha1.New()
		hasher.Write([]byte(filename))
		sha := hex.EncodeToString(hasher.Sum((nil)))

		gitObjPath, err := GetGitObjectPath()

		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot get git object path %s\n", err)
			os.Exit(1)
		}

		targetPath := gitObjPath + sha[:2] + "/" + sha[2:]
		targetFolderPath := ".git/objects/" + sha[:2]

		if !DirExists(targetFolderPath) {
			if err := os.Mkdir(targetFolderPath, os.ModePerm); err != nil {
				fmt.Fprintf(os.Stderr, "Unable to create folder with filepath: %s \n", err)
				os.Exit(1)
			}
		}

		data, err := os.ReadFile(inputPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot read file: %s\n", err)
			os.Exit(1)
		}
		dataToHash := fmt.Sprintf("blob %d\x00%s", len(data), data)

		outputFile, err := os.Create(targetPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot create a file: %s\n", err)
			os.Exit(1)
		}
		defer outputFile.Close()

		zlibWriter := zlib.NewWriter(outputFile)
		defer zlibWriter.Close()

		_, err = zlibWriter.Write([]byte(dataToHash))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot write to file: %s \n", err)
			os.Exit(1)
		}

		fmt.Printf("%s", sha)

	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}

func GetGitObjectPath() (string, error) {
	cwd, err := os.Getwd()

	return cwd + "/.git/objects/", err
}

func DirExists(dirPath string) bool {
	info, err := os.Stat(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false // Directory does not exist
		}
		return false // Some other error, treat as not existing for simplicity
	}
	return info.IsDir() // Check if the path is indeed a directory
}
