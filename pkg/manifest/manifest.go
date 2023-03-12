// Copyright (c) 2019, Sylabs Inc. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package manifest

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/BTMichalowicz/go_util/pkg/util"
)

func getFileHash(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	hasher := sha256.New()
	_, err = io.Copy(hasher, f)
	if err != nil {
		return ""
	}

	return hex.EncodeToString(hasher.Sum(nil))
}

// Hash files returns the hash for a list of files (absolute path)
func HashFiles(files []string) []string {
	var hashData []string

	for _, file := range files {
		hash := getFileHash(file)
		hashData = append(hashData, file+": "+hash)
	}

	return hashData
}

// Create a new manifest
func Create(filepath string, entries []string) error {
	f, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create %s: %s", filepath, err)
	}

	_, err = f.WriteString(strings.Join(entries, "\n"))
	if err != nil {
		return fmt.Errorf("failed to write to %s: %s", filepath, err)
	}

	err = os.Chmod(filepath, 0444)
	if err != nil {
		return fmt.Errorf("failed to set manifest to ready only: %s", err)
	}

	return nil
}

// Load parses a manifest file and returns its content in a form of a map.
// The first returned map is indexed on the file name, while the second one is
// indexed on the hash
func Load(path string) (map[string]string, map[string]string, error) {
	if !util.FileExists(path) {
		// This is currently not an error, just log the fact there is no manifest
		return nil, nil, fmt.Errorf("%s does not exist", path)

	}

	h := make(map[string]string)
	h2 := make(map[string]string)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("failed to read manifest %s", path)
		return nil, nil, err
	}

	content := string(data)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		tokens := strings.Split(line, ": ")
		if len(tokens) == 2 {
			file := tokens[0]
			recordedHash := tokens[1]
			h[file] = recordedHash
			h2[recordedHash] = file
		}
	}
	return h, h2, nil
}

// Check parses a given manifest and check that all hash there are in the manifest are the same than current
// files
func Check(path string) error {
	data, _, err := Load(path)
	if err != nil {
		return err
	}

	for file, hash := range data {
		curFileHash := HashFiles([]string{file})
		if curFileHash[0] != hash {
			actualHash := strings.Split(curFileHash[0], ": ")[1]
			return fmt.Errorf("hashes differ (record: %s; actual: %s)", hash, actualHash)
		}
	}

	return nil
}
