package bkrs

import (
	"bufio"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	yomi "github.com/etsune/bkrs2yomi/pkg/yomi"
	"github.com/teamlint/opencc"
)

type bkrsTerm struct {
	Expression string
	Pinyin     string
	Meaning    string
}

type bkrsCleaner struct {
	newlineRx *regexp.Regexp
	bbCodeRx  *regexp.Regexp
}

func makeCleaner() *bkrsCleaner {
	return &bkrsCleaner{
		newlineRx: regexp.MustCompile(`\[m\d\]`),
		bbCodeRx:  regexp.MustCompile(`\[\/?(m\d?|c|p|ref|b|i|ex|\*)\]`),
	}
}

func DownloadLatest() string {
	// https://bkrs.info/downloads/files2/dabkrs_v88_1.7z
	dlUrlrx := regexp.MustCompile(`downloads\/files2\/dabkrs_v\d+_1\.7z`)
	// тут нужна распаковка 7z
	return DownloadBkrs(dlUrlrx, "release")
}

func DownloadDaily() string {
	dlUrlrx := regexp.MustCompile(`downloads\/daily\/dabkrs_\d+\.gz`)
	return DownloadBkrs(dlUrlrx, "daily")
}

func DownloadDailyRu() string {
	dlUrlrx := regexp.MustCompile(`downloads\/daily\/dabruks_\d+\.gz`)
	return DownloadBkrs(dlUrlrx, "daily ru")
}

func DownloadBkrs(dlUrlrx *regexp.Regexp, version string) string {
	fmt.Printf("Downloading latest %s version...\n", version)
	resFileName := "dabkrs.gz"

	resp, err := http.Get("https://bkrs.info/p47")
	if err != nil {
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	sb := string(body)

	dlUrl := "https://bkrs.info/" + dlUrlrx.FindString(sb)
	resFileName = dlUrl[strings.LastIndex(dlUrl, "/")+1:]

	if _, err := os.Stat(resFileName); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			DownloadBkrsFile(dlUrl, resFileName)
			return resFileName
		}
	}

	fmt.Println("File already exists. Skipping download.")
	return resFileName
}

func DownloadBkrsFile(url, resFileName string) {
	out, err := os.Create(resFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Fatal(err)
	}
}

func ExportDict(inputFile, outputFile string, extended, ru bool, conversion int) error {

	yomi.CreateTempDir()

	fileExt := filepath.Ext(inputFile)

	var scanner *bufio.Scanner

	file, err := os.Open(inputFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	if fileExt == ".gz" {
		gr, err := gzip.NewReader(file)
		if err != nil {
			log.Fatal(err)
		}
		defer gr.Close()
		scanner = bufio.NewScanner(gr)

	} else {
		scanner = bufio.NewScanner(file)
	}

	ConvertDict(scanner, extended, ru, conversion)

	revision := filepath.Base(inputFile)

	title := "大БКРС"
	titleLat := "BKRS"
	url := "https://bkrs.info/"
	description := "Большой китайско-русский словарь, compiled with bkrs2yomi"

	if ru {
		title = "БРуКС"
		titleLat = "BRuKS"
		description = "Большой русско-китайский словарь, compiled with bkrs2yomi"

	} else {

		switch conversion {
		case 0:
			titleLat += "-Simpl"
		case 1:
			titleLat += "-Trad"
		case 2:
			titleLat += "-Trad-Addon"
			title += "-t"
		}

		if extended {
			titleLat += "-Extended"
		}
	}

	yomi.CreateIndexFile(revision, title, url, description)

	yomi.CreateZip(titleLat + "_yomichan.zip")

	yomi.RemoveTempDir()

	return nil
}

func ConvertDict(scanner *bufio.Scanner, extended, ru bool, conversion int) {

	var s2t *opencc.OpenCC

	if conversion == 1 || conversion == 2 {
		s2t, _ = opencc.New("s2t")
	}

	// яп версия
	// t2jp, _ := opencc.New("t2jp")

	cleaner := makeCleaner()

	i := 0
	count := 0
	globalCount := 0
	termFileIndex := 1

	var curBkrsTerm bkrsTerm
	var result yomi.YomiTermList

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if len(line) == 0 || line[0] == '#' {
			i = 0
			curBkrsTerm = bkrsTerm{}
			continue
		}

		switch i {
		case 0:
			curBkrsTerm.Expression = line
		case 1:
			if ru {
				curBkrsTerm.Meaning = CleanBkrsLine(line, cleaner)
				result = append(result, ConvertBkrsTermToYomiTerm(curBkrsTerm))
				count++
			}
			curBkrsTerm.Pinyin = line

		case 2:
			curBkrsTerm.Meaning = CleanBkrsLine(line, cleaner)
			if extended || curBkrsTerm.Pinyin != "_" {
				if conversion == 0 {
					result = append(result, ConvertBkrsTermToYomiTerm(curBkrsTerm))
					count++

				} else if conversion == 1 {
					curBkrsTerm.Expression, _ = s2t.Convert(curBkrsTerm.Expression)

					// яп версия
					// curBkrsTerm.Expression, _ = t2jp.Convert(curBkrsTerm.Expression)

					result = append(result, ConvertBkrsTermToYomiTerm(curBkrsTerm))
					count++

				} else if conversion == 2 {
					trad, _ := s2t.Convert(curBkrsTerm.Expression)

					// яп версия
					// trad, _ = t2jp.Convert(trad)

					if trad != curBkrsTerm.Expression {
						curBkrsTerm.Expression = trad
						result = append(result, ConvertBkrsTermToYomiTerm(curBkrsTerm))
						count++
					}
				}
			}
		}

		i++

		if count >= 10000 {
			yomi.WriteYomiFile(result, termFileIndex)
			result = yomi.YomiTermList{}
			globalCount += count
			termFileIndex++
			count = 0
		}
	}

	if count > 0 {
		yomi.WriteYomiFile(result, termFileIndex)
		globalCount += count
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Parsing complete — %d entries.\n", globalCount)
}

func CleanBkrsLine(meaning string, cleaner *bkrsCleaner) string {
	meaning = cleaner.newlineRx.ReplaceAllString(meaning, "\n")
	meaning = cleaner.bbCodeRx.ReplaceAllString(meaning, "")
	meaning = strings.Replace(meaning, "\\[", "[", -1)
	meaning = strings.Replace(meaning, "\\]", "]", -1)
	meaning = strings.TrimSpace(meaning)

	return meaning
}

func ConvertBkrsTermToYomiTerm(term bkrsTerm) yomi.YomiTerm {
	yomi := yomi.YomiTerm{
		Expression: term.Expression,
		Reading:    term.Pinyin,
		Glossary:   []string{term.Meaning},
		// Rules:          []string{},
		// DefinitionTags: []string{},
		// TermTags:       []string{},
	}

	return yomi
}
