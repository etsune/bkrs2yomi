package yomi

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	tempDir = "yomi_tmp"
)

type YomiTerm struct {
	Expression     string
	Reading        string
	DefinitionTags []string
	Rules          []string
	Score          int
	Glossary       []string
	Sequence       int
	TermTags       []string
}

type YomiTermRecord []interface{}
type YomiTermRecordList []YomiTermRecord

type YomiTermList []YomiTerm

type YomiIndex struct {
	Title       string `json:"title"`
	Format      int    `json:"format"`
	Revision    string `json:"revision"`
	Sequenced   bool   `json:"sequenced"`
	Url         string `json:"url"`
	Description string `json:"description"`
}

func WriteYomiFile(terms YomiTermList, fileIndex int) error {
	var results YomiTermRecordList
	for _, t := range terms {
		result := YomiTermRecord{
			t.Expression,
			t.Reading,
			strings.Join(t.DefinitionTags, " "),
			strings.Join(t.Rules, " "),
			t.Score,
			t.Glossary,
			t.Sequence,
			strings.Join(t.TermTags, " "),
		}
		results = append(results, result)
	}

	json, err := json.Marshal(results)
	if err != nil {
		return err
	}

	os.WriteFile(tempDir+"/"+fmt.Sprintf("term_bank_%d.json", fileIndex), json, os.ModePerm)

	return nil
}

func CreateZip(zipFileName string) {
	archive, err := os.Create(zipFileName)
	if err != nil {
		panic(err)
	}

	defer archive.Close()
	zipWriter := zip.NewWriter(archive)

	files, err := ioutil.ReadDir(tempDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		filename := filepath.Join(tempDir, f.Name())
		f1, err := os.Open(filename)
		if err != nil {
			log.Fatal(err)
		}
		defer f1.Close()
		w1, err := zipWriter.Create(f.Name())
		if err != nil {
			log.Fatal(err)
		}
		if _, err := io.Copy(w1, f1); err != nil {
			log.Fatal(err)
		}
	}

	zipWriter.Close()
}

func CreateIndexFile(revision, title, url, description string) {
	index := YomiIndex{
		Title:       title,
		Format:      3,
		Revision:    revision,
		Sequenced:   true,
		Url:         url,
		Description: description,
	}

	json, err := json.Marshal(index)
	if err != nil {
		log.Fatal(err)
	}

	os.WriteFile(tempDir+"/"+"index.json", json, os.ModePerm)
}

func RemoveTempDir() {
	newpath := filepath.Join(".", tempDir)
	os.RemoveAll(newpath)
}

func CreateTempDir() {
	newpath := filepath.Join(".", tempDir)
	os.RemoveAll(newpath)
	os.MkdirAll(newpath, os.ModePerm)
}
