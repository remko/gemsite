package main

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"
	"unicode"
)

const MinSearchWordLength = 3

var author = "Remko Tronçon"
var contentDir = "gemsite"
var contentSrcDir = "content"
var generated = []string{"blog.gmi", "index.gmi"}

func build() error {
	site := Site{Posts: []Page{}}

	// Generate pages
	pages, err := buildContent(contentSrcDir, contentDir)
	if err != nil {
		return err
	}

	// Collect posts
	for _, page := range pages {
		if strings.HasPrefix(page.Path, "blog/") && !page.Time.IsZero() && page.Title != "" {
			site.Posts = append(site.Posts, page)
		}
	}
	sort.Slice(site.Posts, func(i, j int) bool {
		return site.Posts[i].Time.After(site.Posts[j].Time)
	})

	// Generate collection pages
	for _, p := range generated {
		tmpl := template.Must(template.ParseFiles(fmt.Sprintf("templates/%s.tmpl", p)))
		f := CreateIfChangedFile(path.Join(contentDir, p))
		defer f.Close()
		if err := tmpl.Execute(f, site); err != nil {
			return err
		}
		if err = f.Close(); err != nil {
			return err
		}
	}

	// Index pages
	f, err := os.Create("search.idx")
	if err != nil {
		return err
	}
	defer f.Close()
	if err = writeSearchIndex(os.DirFS(contentDir), pages, f); err != nil {
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}

	return nil
}

func buildContent(srcdir string, destdir string) ([]Page, error) {
	srcfs := os.DirFS(srcdir)
	pages := []Page{}
	err := fs.WalkDir(srcfs, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".md") {
			inf, err := srcfs.Open(path)
			if err != nil {
				return err
			}
			defer inf.Close()
			outp := path[:len(path)-3] + ".gmi"
			outf := CreateIfChangedFile(filepath.Join(destdir, outp))
			defer outf.Close()
			page, err := ConvertMarkdownToGemtext(inf, outf)
			if err != nil {
				return fmt.Errorf("error converting %s: %w", path, err)
			}
			if err = outf.Close(); err != nil {
				return err
			}
			page.Path = outp
			page.URL = "/" + path[:len(path)-3]
			pages = append(pages, page)
		} else {
			if err := copyFile(filepath.Join(srcdir, path), filepath.Join(destdir, path)); err != nil {
				return err
			}
			if strings.HasSuffix(path, ".gmi") && !contains(generated, path) && !strings.HasPrefix("_", path) {
				page, err := parsePage(srcfs, path)
				if err != nil {
					return fmt.Errorf("%s: %w", path, err)
				}
				page.Path = path
				page.URL = "/" + path[:len(path)-4]
				pages = append(pages, page)
			}
		}
		return nil
	})
	return pages, err
}

func copyFile(from, to string) error {
	inf, err := os.Open(from)
	if err != nil {
		return err
	}
	defer inf.Close()
	outf := CreateIfChangedFile(to)
	defer outf.Close()
	_, err = io.Copy(outf, inf)
	if err != nil {
		return err
	}
	return outf.Close()
}

func parsePage(content fs.FS, path string) (Page, error) {
	page := Page{}
	f, err := content.Open(path)
	if err != nil {
		return page, err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Text()
		if line == "" {
			continue
		}
		if page.Title == "" {
			if !strings.HasPrefix(line, "# ") {
				return page, fmt.Errorf("missing title")
			}
			page.Title = strings.Trim(line[1:], " ")
		} else {
			_, rawdate, found := strings.Cut(line, "·")
			if found {
				t, err := time.Parse("January _2, 2006", strings.Trim(rawdate, " "))
				if err == nil {
					page.Time = t
				}
			}
			break
		}
	}
	return page, s.Err()
}

func writeSearchIndex(content fs.FS, pages []Page, w io.Writer) error {
	bw := bufio.NewWriter(w)
	defer bw.Flush()
	for _, page := range pages {
		f, err := content.Open(page.Path)
		if err != nil {
			return err
		}
		defer f.Close()
		tokens := map[string]struct{}{}
		ls := bufio.NewScanner(f)
		for ls.Scan() {
			line := ls.Text()
			if strings.HasPrefix(line, "=>") {
				continue
			}
			words := strings.FieldsFunc(line, func(r rune) bool { return !unicode.IsLetter(r) })
			for _, word := range words {
				word = strings.ToLower(word)
				if len(word) < MinSearchWordLength || word == "remko" || word == "tron\xc3\xa7on" {
					continue
				}
				tokens[word] = struct{}{}
			}
		}

		// Write entry
		bw.WriteString(page.URL)
		bw.WriteByte(0)
		bw.WriteString(page.Title)
		bw.WriteByte(0)
		bw.WriteString(page.Date())
		for t := range tokens {
			bw.WriteByte(0)
			bw.WriteString(t)
		}
		bw.WriteByte(0xa)
	}
	return nil
}

type Site struct {
	Posts []Page
}

type Page struct {
	URL      string
	Path     string
	Time     time.Time
	Title    string
	Featured bool
}

func (p Page) Date() string {
	return p.Time.Format(time.DateOnly)
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func main() {
	if err := build(); err != nil {
		panic(err)
	}
}
