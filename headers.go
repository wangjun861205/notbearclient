package notbearclient

import (
	"io/ioutil"
	"os"
	"regexp"
)

type Header map[string]string
type HeaderMap map[string]Header

var AllHeaderRe = regexp.MustCompile(`(?ms)([\w_]+)\s*?=\s*?\{(.*?)\}`)
var HeaderRe = regexp.MustCompile(`([\w_-]*?)\s*?:\s*?(.*?)\n`)
var HeadersMap HeaderMap

func HandleErr(err error) {
	if err != nil {
		panic(err)
	}
}

func init() {
	file, err := os.Open("headers.settings")
	HandleErr(err)
	b, err := ioutil.ReadAll(file)
	HandleErr(err)
	s := string(b)
	hm := HeaderMap{}
	headers := AllHeaderRe.FindAllStringSubmatch(s, -1)
	for _, header := range headers {
		headerName := header[1]
		headerContent := header[2]
		hd := Header{}
		hs := HeaderRe.FindAllStringSubmatch(headerContent, -1)
		for _, h := range hs {
			hd[h[1]] = h[2]
		}
		hm[headerName] = hd
	}
	HeadersMap = hm
}
