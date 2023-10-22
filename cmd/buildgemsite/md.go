package main

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
)

type State int

const (
	Initial State = iota
	InFrontMatter
	InBody
	InCode
)

type TextBlockType int

const (
	Normal = iota
	Bullet
	Quote
)

var imageRE = regexp.MustCompile(`^\!\[([^\]]*)\]\(([^\) ]*)( "([^\)]*)")?\)$`)
var linkRE = regexp.MustCompile(`\!?\[([^\]]*)\]\(([^\) ]*)( "([^\)]*)")?\)`)

func ConvertMarkdownToGemtext(in io.Reader, outw io.Writer) (Page, error) {
	out := bufio.NewWriter(outw)
	defer out.Flush()

	state := Initial
	links := [][]string{}

	var textblock *strings.Builder
	textblockType := Normal

	extractLinks := func(line string) string {
		lms := linkRE.FindAllStringSubmatch(line, -1)
		if len(lms) > 0 {
			links = append(links, lms...)
			line = linkRE.ReplaceAllString(line, "$1")
		}
		return line
	}

	addText := func(s string) {
		if textblock == nil {
			textblock = &strings.Builder{}
		} else {
			textblock.WriteByte(0x20)
		}
		textblock.WriteString(s)
	}

	flushTextBlock := func() {
		if textblock != nil {
			line := extractLinks(textblock.String())
			if textblockType == Bullet {
				out.WriteString("* ")
			} else if textblockType == Quote {
				out.WriteString("> ")
			}
			out.WriteString(line)
			out.WriteByte(0xa)
			textblock = nil
			textblockType = Normal
		}
	}

	flushLinks := func() bool {
		if len(links) > 0 {
			for _, link := range links {
				title := link[1]
				if len(link[4]) > 0 {
					title = link[4]
				}
				out.WriteString(fmt.Sprintf("=> %s %s\n", link[2], title))
			}
			links = [][]string{}
			return true
		}
		return false
	}

	scn := bufio.NewScanner(in)
	scn.Split(bufio.ScanLines)
	page := Page{}
	for scn.Scan() {
		line := scn.Text()
		if state == Initial {
			if strings.HasPrefix(line, "---") {
				state = InFrontMatter
			} else {
				return page, fmt.Errorf("missing front matter")
			}
		} else if state == InFrontMatter {
			if strings.HasPrefix(line, "---") {
				state = InBody
			} else if strings.HasPrefix(line, "title: ") {
				title := strings.TrimSpace(line[7:])
				title = title[1 : len(title)-1]
				page.Title = title
				out.WriteString(fmt.Sprintf("# %s\n\n", title))
			} else if strings.HasPrefix(line, "featured: ") {
				page.Featured = true
			} else if strings.HasPrefix(line, "date: ") {
				t, err := time.Parse("2006-01-02", line[6:])
				if err != nil {
					return page, err
				}
				page.Time = t
				out.WriteString(fmt.Sprintf("%s · %s\n\n", author, t.Format("January 2, 2006")))
			}
		} else if state == InBody {
			if line == "" {
				flushTextBlock()
				out.WriteString("\n")
				if flushLinks() {
					out.WriteString("\n")
				}
				continue
			}
			if strings.HasPrefix(line, "```") {
				flushTextBlock()
				out.WriteString(line)
				out.WriteByte(0xa)
				state = InCode
				continue
			}
			if strings.HasPrefix(line, "- ") {
				flushTextBlock()
				textblockType = Bullet
				addText(line[2:])
				continue
			}
			if line[0] == '>' {
				if textblockType != Quote {
					flushTextBlock()
				}
				if len(line) >= 3 {
					addText(line[2:])
					textblockType = Quote
				} else {
					flushTextBlock()
					out.WriteString(">\n")
				}
				continue
			}

			m := imageRE.FindStringSubmatch(line)
			if m != nil {
				flushTextBlock()
				out.WriteString(fmt.Sprintf("=> %s %s\n", m[2], m[1]))
				continue
			}

			if textblockType == Bullet {
				if !strings.HasPrefix(line, "  ") {
					return page, fmt.Errorf("expected indent: %s", line)
				}
				line = line[2:]
			}

			addText(line)
		} else if state == InCode {
			out.WriteString(line)
			out.WriteByte(0xa)
			if strings.HasPrefix(line, "```") {
				state = InBody
				flushLinks()
			}
		}
	}
	flushTextBlock()
	flushLinks()

	return page, nil
}
