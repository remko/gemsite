package gemsite

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"mime"
	"net"
	"net/http"
	"net/textproto"
	"net/url"
	"path/filepath"
	"regexp"
	"runtime/pprof"
	"sort"
	"strings"
	"sync"
	"text/template"
	"time"
)

////////////////////////////////////////////////////////////////////////////////
// Server
////////////////////////////////////////////////////////////////////////////////

func Listen(laddr string) error {
	// Initialization
	mime.AddExtensionType(".gmi", "text/gemini")
	var err error
	content, err = fs.Sub(assets, "gemsite")
	if err != nil {
		return err
	}

	// Load search index
	if err := loadSearchIndex(); err != nil {
		return err
	}

	// TLS setup
	cert, err := tls.X509KeyPair(servercert, serverkey)
	if err != nil {
		return err
	}
	clientCertCAs := x509.NewCertPool()
	if ok := clientCertCAs.AppendCertsFromPEM(servercert); !ok {
		return fmt.Errorf("unable to add client ca cert")
	}

	listen, err := tls.Listen("tcp", laddr,
		&tls.Config{
			Certificates: []tls.Certificate{cert},
			ClientAuth:   tls.VerifyClientCertIfGiven,
			ClientCAs:    clientCertCAs,
		},
	)
	if err != nil {
		return err
	}
	defer listen.Close()
	log.Printf("listening on %s", laddr)

	// Accept connections
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Printf("error accepting connection: %v", err)
			continue
		}
		go handleRequest(conn)
	}
}

// Request handler
// Protocol: https://geminiprotocol.net/docs/specification.gmi
func handleRequest(conn net.Conn) {
	start := time.Now()
	defer conn.Close()

	tp := textproto.NewConn(conn)
	line, err := tp.ReadLine()
	defer func() {
		log.Printf("%s %s (%v)", conn.RemoteAddr(), line, time.Since(start))
	}()
	if err != nil {
		log.Printf("error reading request: %v", err)
		return
	}

	uri, err := url.Parse(line)
	if err != nil {
		log.Printf("error parsing url: %v", err)
		tp.PrintfLine("59")
		return
	}

	// Admin auth
	if strings.HasPrefix(uri.Path, "/_admin") {
		cstate := conn.(*tls.Conn).ConnectionState()
		if len(cstate.PeerCertificates) == 0 {
			tp.PrintfLine("60")
			return
		}
		cert := cstate.PeerCertificates[0]
		if !strings.HasPrefix(cert.Subject.CommonName, "admin@") {
			tp.PrintfLine("61")
			return
		}
	}

	// Handle dynamic paths
	path := uri.Path
	switch path {
	case "/search":
		search, err := url.QueryUnescape(uri.RawQuery)
		if err != nil {
			log.Printf("invalid search query: %v", err)
			tp.PrintfLine("59")
		} else if len(search) == 0 {
			tp.PrintfLine("10 Search:")
		} else {
			tp.PrintfLine("20 text/gemini")
			query := strings.Fields(search)
			pages := searchPages(query)
			if err := searchTemplate.Execute(conn, SearchTemplateContext{Query: strings.Join(query, " "), Pages: pages}); err != nil {
				log.Printf("error rendering: %v", err)
				return
			}
		}
		return

	case "/ublog":
		statuses, err := fetchStatuses()
		if err != nil {
			log.Printf("error fetching statuses: %v", err)
			tp.PrintfLine("42")
			return
		}

		tp.PrintfLine("20 text/gemini")
		if err := ublogTemplate.Execute(conn, UBlogTemplateContext{Statuses: statuses}); err != nil {
			log.Printf("error rendering: %v", err)
			return
		}
		return

	case "/_admin/pprof/profile":
		tp.PrintfLine("20 application/octet-stream")
		if err := pprof.StartCPUProfile(conn); err != nil {
			log.Printf("error collecting profile: %v", err)
			return
		}
		time.Sleep(20 * time.Second)
		pprof.StopCPUProfile()
		return
	}

	// Serve static file
	f, err := content.Open(pathToFile(path))
	if err != nil {
		log.Printf("error opening file: %v", err)
		tp.PrintfLine("51")
		return
	}
	tp.PrintfLine("20 %s", mime.TypeByExtension(filepath.Ext(path)))
	_, err = io.Copy(conn, f)
	if err != nil {
		log.Printf("error sending response: %v", err)
	}
}

////////////////////////////////////////////////////////////////////////////////
// Search
////////////////////////////////////////////////////////////////////////////////

const MinSearchWordLength = 3

type Page struct {
	Path  string
	Date  string
	Title string
}

var searchIndex map[string]map[*Page]struct{}

// Loads the search index from disk
// The search index consists of lines of null-separated strings
func loadSearchIndex() error {
	searchIndex = map[string]map[*Page]struct{}{}
	sis := bufio.NewScanner(strings.NewReader(searchidx))
	sis.Split(bufio.ScanLines)
	for sis.Scan() {
		entry := strings.Split(sis.Text(), "\x00")
		page := Page{
			Path:  entry[0],
			Title: entry[1],
			Date:  entry[2],
		}
		for _, word := range entry[3:] {
			ps, ok := searchIndex[word]
			if !ok {
				ps = map[*Page]struct{}{}
			}
			ps[&page] = struct{}{}
			searchIndex[word] = ps
		}
	}
	if err := sis.Err(); err != nil {
		return err
	}
	return nil
}

func searchPages(query []string) []*Page {
	var pages map[*Page]struct{}
	for _, q := range query {
		if len(q) <= MinSearchWordLength {
			continue
		}
		ps, ok := searchIndex[strings.ToLower(q)]
		if !ok {
			return []*Page{}
		}
		if pages == nil {
			pages = ps
		} else {
			npages := map[*Page]struct{}{}
			for page := range pages {
				if _, ok := ps[page]; ok {
					npages[page] = struct{}{}
				}
			}
			pages = npages
		}
	}
	result := make([]*Page, 0, len(pages))
	for page := range pages {
		result = append(result, page)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Title < result[j].Title
	})
	return result
}

////////////////////////////////////////////////////////////////////////////////
// Microblog
////////////////////////////////////////////////////////////////////////////////

var mastodonID = "109530760716287685"
var mastodonHost = "mas.to"
var mastodonFetchInterval = 1 * time.Hour

type Status struct {
	ID        string
	Content   string
	CreatedAt time.Time
	Links     []Link
	URL       string
}

type Link struct {
	URL   string
	Title string
}

var statuses []Status
var lastStatusesFetch time.Time
var statusesMu sync.Mutex

func fetchStatuses() ([]Status, error) {
	if time.Since(lastStatusesFetch) < mastodonFetchInterval {
		return statuses, nil
	}
	statusesMu.Lock()
	defer statusesMu.Unlock()
	if time.Since(lastStatusesFetch) < mastodonFetchInterval {
		return statuses, nil
	}

	// https://docs.joinmastodon.org/methods/accounts/#statuses
	url := fmt.Sprintf("https://%s/api/v1/accounts/%s/statuses?exclude_replies=1&exclude_reblogs=1&limit=50", mastodonHost, mastodonID)
	if len(statuses) > 0 {
		url = url + "&min_id=" + statuses[0].ID
	}
	log.Printf("fetching statuses: %s", url)
	r, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	lastStatusesFetch = time.Now()

	var mstatuses []MStatus
	if err := json.NewDecoder(r.Body).Decode(&mstatuses); err != nil {
		return nil, err
	}
	log.Printf("fetched statuses: %d", len(mstatuses))
	var nstatuses []Status
	for _, s := range mstatuses {
		tcontent := htmltagRE.ReplaceAllString(s.Content, "")
		if len(tcontent) == 0 {
			continue
		}
		if tcontent[0] == '@' {
			continue
		}
		pcontent := strings.ReplaceAll(tcontent, "&#39;", "'")
		pcontent = strings.ReplaceAll(pcontent, "&quot;", "'")
		pcontent = strings.ReplaceAll(pcontent, "&amp;", "&")
		lms := urlRE.FindAllStringSubmatch(tcontent, -1)
		if len(lms) > 1 {
			pcontent = urlRE.ReplaceAllString(pcontent, "🌐")
		} else {
			pcontent = urlRE.ReplaceAllString(pcontent, "")
		}
		status := Status{ID: s.ID, Content: pcontent, CreatedAt: s.CreatedAt, URL: s.URL}
		for _, m := range lms {
			if s.Card.URL == m[0] {
				status.Links = append(status.Links, Link{URL: rewriteURL(s.Card.URL), Title: s.Card.Title})
			} else {
				url := rewriteURL(m[0])
				status.Links = append(status.Links, Link{URL: url, Title: stripURL(url)})
			}
		}
		for _, ma := range s.MediaAttachments {
			status.Links = append(status.Links, Link{URL: rewriteURL(ma.URL), Title: "🖼 " + ma.Description})
		}
		nstatuses = append(nstatuses, status)
	}
	statuses = append(nstatuses, statuses...)
	return statuses, nil
}

// https://docs.joinmastodon.org/entities/Status/
type MStatus struct {
	ID               string
	CreatedAt        time.Time `json:"created_at"`
	Content          string    `json:"content"`
	URL              string    `json:"url"`
	MediaAttachments []struct {
		URL         string `json:"url"`
		Description string `json:"description"`
	} `json:"media_attachments"`
	Card struct {
		URL   string `json:"url"`
		Title string `json:"title"`
	} `json:"card"`
}

// https://docs.joinmastodon.org/spec/activitypub/#sanitization
var htmltagRE = regexp.MustCompile(`</?(p|span|br|a|del|pre|code|em|strong|b|i|u|ul|ol|li|blockquote)[^>]*>`)

var urlRE = regexp.MustCompile(`https?://[^ ]*`)

////////////////////////////////////////////////////////////////////////////////
// Common
////////////////////////////////////////////////////////////////////////////////

var localRE = regexp.MustCompile(`^https?://mko.re`)

func rewriteURL(url string) string {
	if localRE.MatchString(url) {
		path := strings.TrimRight(localRE.ReplaceAllString(url, ""), "/")
		_, err := content.Open(pathToFile(path))
		if err == nil {
			return path
		}
	}
	return url
}

func stripURL(fullurl string) string {
	u, _ := url.Parse(fullurl)
	u.RawQuery = ""
	u.Fragment = ""
	return u.String()
}

func pathToFile(path string) string {
	if path == "/" || path == "" {
		return "index.gmi"
	} else if !strings.Contains(path, ".") {
		return path[1:] + ".gmi"
	} else {
		return path[1:]
	}
}

////////////////////////////////////////////////////////////////////////////////
// Resources
////////////////////////////////////////////////////////////////////////////////

//go:embed server.key
var serverkey []byte

//go:embed server.crt
var servercert []byte

//go:embed gemsite all:gemsite/_admin.gmi
var assets embed.FS
var content fs.FS

//go:embed search.idx
var searchidx string

//go:embed templates/search.gmi.tmpl
var searchTmpl embed.FS
var searchTemplate = template.Must(template.ParseFS(searchTmpl, "templates/search.gmi.tmpl"))

type SearchTemplateContext struct {
	Query string
	Pages []*Page
}

//go:embed templates/ublog.gmi.tmpl
var ublogTmpl embed.FS
var ublogTemplate = template.Must(template.ParseFS(ublogTmpl, "templates/ublog.gmi.tmpl"))

type UBlogTemplateContext struct {
	Statuses []Status
}
