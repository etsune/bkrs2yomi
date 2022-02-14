package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"

	bkrs2yomi "github.com/etsune/bkrs2yomi/pkg/bkrs"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [options] input-file\n", path.Base(os.Args[0]))
	fmt.Fprint(os.Stderr, "")
	fmt.Fprint(os.Stderr, "Parameters:\n")
	flag.PrintDefaults()
}

func main() {
	var (
		extended = flag.Bool("extended", false, "exports extended version of BKRS (includes entries without pinyin).")
		// latest     = flag.Bool("latest", false, "downloads latest release version and uses it for the conversion.")
		daily      = flag.Bool("daily", false, "downloads latest daily version and uses it for the conversion.")
		conversion = flag.Int("type", 0, "type of the conversion. 0 - simplified hanzi, 1 - traditional, 2 - traditional addon for type 0, excluding duplicates.")
	)

	var inputFile string

	flag.Usage = usage
	flag.Parse()

	if flag.NArg() == 0 {
		if *daily {
			inputFile = bkrs2yomi.DownloadDaily()
			// } else if *latest {
			// 	inputFile = bkrs2yomi.DownloadLatest()
		} else {
			usage()
			os.Exit(2)
		}
	} else {
		inputFile = flag.Arg(0)
	}

	if err := bkrs2yomi.ExportDict(inputFile, flag.Arg(1), *extended, *conversion); err != nil {
		log.Fatal(err)
	}
}
