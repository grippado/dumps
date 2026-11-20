package main

import (
	"errors"
	"os"
	"path/filepath"
)

func resolveRoot(flagValue string) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if st, err := os.Stat(filepath.Join(wd, "dumps-meta")); err != nil || !st.IsDir() {
		return "", errors.New("sem --root e sem dumps-meta/ no diretório atual; rode dentro da raiz do dumps ou passe --root")
	}
	return wd, nil
}
