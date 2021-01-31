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

	suffix := trimLetterStr(fmt.Sprintf("%s%s", trimLetterStr(CamelString(*KeyType)), trimLetterStr(CamelString(*ValueType))))

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
		} else if src[i] == '[' {
			dst = append(dst, 'S')
		} else if src[i] == '{' {
			dst = append(dst, 'M')
		}
	}
	return string(dst)
}

// camel string, xx_yy to XxYy
func CamelString(s string) string {
	data := make([]byte, 0, len(s))
	j := false
	k := false
	num := len(s) - 1
	for i := 0; i <= num; i++ {
		d := s[i]
		if k == false && d >= 'A' && d <= 'Z' {
			k = true
		}
		if d >= 'a' && d <= 'z' && (j || k == false) {
			d = d - 32
			j = false
			k = true
		}
		if k && d == '_' && num > i && s[i+1] >= 'a' && s[i+1] <= 'z' {
			j = true
			continue
		}
		data = append(data, d)
	}
	return string(data[:])
}

func ckerr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
