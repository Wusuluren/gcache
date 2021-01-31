package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
)

var (
	KeyType            = flag.String("KeyType", "int", "")
	ValueType          = flag.String("ValueType", "int", "")
	hashFn             = flag.String("hashFn", "hashInt", "")
	cleanupInterval    = flag.String("cleanupInterval", "10 * time.Minute", "")
	maxSlotSize        = flag.String("maxSlotSize", "1024*1024", "")
	reduceSlotSizeRate = flag.String("reduceSlotSizeRate", "0.75", "")
)

func main() {
	flag.Parse()

	srcBlob, err := ioutil.ReadFile("cache.g")
	ckerr(err)

	suffix := fmt.Sprintf("%s%s", trimLetterStr(upperFirstLetter(*KeyType)), trimLetterStr(upperFirstLetter(*ValueType)))

	dstBlob := srcBlob
	dstBlob = bytes.Replace(dstBlob, []byte("KeyType"), []byte(*KeyType), -1)
	dstBlob = bytes.Replace(dstBlob, []byte("ValueType"), []byte(*ValueType), -1)
	dstBlob = bytes.Replace(dstBlob, []byte("_hashFn"), []byte(*hashFn), -1)
	dstBlob = bytes.Replace(dstBlob, []byte("_cleanupInterval"), []byte(*cleanupInterval), -1)
	dstBlob = bytes.Replace(dstBlob, []byte("_maxSlotSize"), []byte(*maxSlotSize), -1)
	dstBlob = bytes.Replace(dstBlob, []byte("_reduceSlotSizeRate"), []byte(*reduceSlotSizeRate), -1)

	dstBlob = reRepl(dstBlob, `valueItem[\W]+`, "valueItem", "valueItem"+suffix)
	dstBlob = reRepl(dstBlob, `valueItemSort[\W]+`, "valueItemSort", "valueItemSort"+suffix)
	dstBlob = reRepl(dstBlob, `valueItemsSort[\W]+`, "valueItemsSort", "valueItemsSort"+suffix)
	dstBlob = reRepl(dstBlob, `cacheMap[\W]+`, "cacheMap", "cacheMap"+suffix)
	dstBlob = reRepl(dstBlob, `Cache[\W]+`, "Cache", "Cache"+suffix)
	dstBlob = reRepl(dstBlob, `NewCache[\W]+`, "NewCache", "NewCache"+suffix)

	fileName := fmt.Sprintf("cache_%s.go", suffix)
	err = ioutil.WriteFile(fileName, dstBlob, 0666)
	ckerr(err)
}

func reRepl(blob []byte, rePattern string, src, dst string) []byte {
	var reAll [][]byte
	reAll = regexp.MustCompile(rePattern).FindAll(blob, -1)
	for i := range reAll {
		reRepl := bytes.ReplaceAll(reAll[i], []byte(src), []byte(dst))
		blob = bytes.Replace(blob, reAll[i], reRepl, -1)
	}
	return blob
}

func trimLetterStr(s string) string {
	src := []byte(s)
	var dst []byte
	for i := range src {
		if (src[i] >= 'a' && src[i] <= 'z') ||
			(src[i] >= 'A' && src[i] <= 'Z') ||
			(src[i] >= '0' && src[i] <= '9') {
			dst = append(dst, src[i])
		} else if src[i] == '*' {
			dst = append(dst, 'P')
		}
	}
	return string(dst)
}

func upperFirstLetter(s string) string {
	if len(s) == 0 {
		return s
	}
	b := []byte(s)
	if b[0] >= 'a' && b[0] <= 'z' {
		b[0] -= 'a' - 'A'
	}
	return string(b)
}

func ckerr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
