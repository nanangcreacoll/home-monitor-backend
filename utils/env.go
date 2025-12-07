package utils

import (
	"os"
	"path/filepath"
)

func FindDotEnv(levels int) string {
	const envFileName = ".env"

	if infoPath, err := os.Stat(envFileName); err == nil && !infoPath.IsDir() {
		return envFileName
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return ""
	}

	for range levels {
		currentDir = filepath.Dir(currentDir)
		envPath := filepath.Join(currentDir, envFileName)
		if infoPath, err := os.Stat(envPath); err == nil && !infoPath.IsDir() {
			return envPath
		}

		if currentDir == "/" {
			break
		}
	}

	currentDir, err = os.Getwd()
	if err != nil {
		return ""
	}

	for range levels {
		entries, err := os.ReadDir(currentDir)
		if err != nil {
			return ""
		}

		found := false
		for _, entry := range entries {
			if entry.IsDir() {
				envPath := filepath.Join(currentDir, entry.Name(), envFileName)
				if infoPath, err := os.Stat(envPath); err == nil && !infoPath.IsDir() {
					return envPath
				}
			}
		}

		if currentDir == "/" {
			break
		}

		if !found {
			break
		}
		currentDir = filepath.Dir(currentDir)
	}

	return ""
}
