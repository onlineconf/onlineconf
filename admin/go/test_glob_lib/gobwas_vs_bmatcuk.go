package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"
	_ "unsafe"

	_ "github.com/bmatcuk/doublestar/v4"
	"github.com/gobwas/glob"
)

//go:linkname matchWithSeparator github.com/bmatcuk/doublestar/v4.matchWithSeparator
func matchWithSeparator(pattern, name string, separator rune, validate bool) (matched bool, err error)

type serversGlobs struct {
	ServerGlobs []string `json:"server_globs"`
	GroupGlobs  []string `json:"group_globs"`
	Servers     []string `json:"servers"`
}

func testGlobs(globs, servers []string, separator rune) {
	for _, glb := range globs {
		g, err := glob.Compile(glb, separator)
		if err != nil {
			log.Printf("gobwas: %s: %v", glb, err)
			continue
		}

		for _, srv := range servers {
			gobwas := g.Match(srv)
			bmatcuk, err := matchWithSeparator(glb, srv, separator, true)

			if err != nil {
				log.Printf("bmatcuk: %s: %v", glb, err)
			}

			if gobwas != bmatcuk {
				log.Printf("glob=%s host=%s: gobwas=%t, bmatcuk=%t", glb, srv, gobwas, bmatcuk)
			}
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("usage: %s <fname.json|->", os.Args[0])
	}

	var (
		buf  []byte
		err  error
		data serversGlobs
	)

	if os.Args[1] == "-" {
		buf, err = io.ReadAll(os.Stdin)
	} else {
		buf, err = ioutil.ReadFile(os.Args[1])
	}

	if err != nil {
		log.Fatalf("%s: %v", os.Args[1], err)
	}

	if err = json.Unmarshal(buf, &data); err != nil {
		log.Fatal(err)
	}

	log.Printf("read %d servers, %d server globs, %d group globs\n", len(data.Servers), len(data.ServerGlobs), len(data.GroupGlobs))

	testGlobs(data.ServerGlobs, data.Servers, '/')
	testGlobs(data.GroupGlobs, data.Servers, '.')
}
