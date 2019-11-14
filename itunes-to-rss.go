package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"text/template"

	"github.com/markbates/pkger"
)

func main() {
	if len(os.Args) > 1 {
		processCommandLine(os.Args[1:])
	} else {
		httpServer()
	}
}

func httpServer() {
	tmpl := loadTemplates()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Got request: URL=%s method=%s", r.URL, r.Method)

		if r.Method != http.MethodPost {
			if err := tmpl.Execute(w, nil); err != nil {
				http.Error(w, "internal processing error", http.StatusInternalServerError)
			}
			return
		}

		link := r.FormValue("link")
		log.Printf("Got link %s", link)
		feedURLs, err := processLink(link)
		if err != nil {
			log.Printf("W: Cannot extract feed URLs from '%s': %s", link, err)
			return
		}
		log.Printf("Got feed URLs %#v", feedURLs)

		data := struct{ FeedURLs []PodcastInfo }{feedURLs}
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "internal processing error", http.StatusInternalServerError)
		}
	})

	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Printf("E: Web server error: %s", err)
	}
}

func loadTemplates() *template.Template {
	file, err := pkger.Open("/public/index.html.tmpl")
	if err != nil {
		panic("E: Cannot open index.html template. Stop.")
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	if err != nil {
		panic("E: Cannot read index.html template. Stop.")
	}

	s := string(b)

	return template.Must(template.New("index.html").Parse(s))
}

func processCommandLine(args []string) {
	for _, link := range args {
		feedURLs, err := processLink(link)
		if err != nil {
			fmt.Printf("W: Cannot extract feed URLs from '%s': %s", link, err)
			continue
		}

		for _, feedURL := range feedURLs {
			fmt.Println(feedURL.Name, feedURL.Feed)
		}
	}
}

func processLink(link string) ([]PodcastInfo, error) {
	lookupURL, err := linkToLookupURL(link)
	if err != nil {
		err = fmt.Errorf(
			"cannot obtain lookup URL from '%s': %w",
			link, err)
		return nil, err
	}

	resp, err := http.Get(lookupURL)
	if err != nil {
		err = fmt.Errorf(
			"cannot retrieve podcast information for '%s': %w",
			link, err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf(
			"unexpected response from server when retrieving podcast information for '%s': %s",
			link, resp.Status)
		return nil, err
	}

	return extractFeedURLs(resp.Body)
}

func linkToLookupURL(link string) (string, error) {
	const lookupTemplate = `https://itunes.apple.com/lookup?id=%s&entity=podcast`

	// validate the target URL
	u, err := url.Parse(link)
	if err != nil {
		err = fmt.Errorf("cannot parse '%s': %w", link, err)
		return "", err
	}

	// the expected form for the URL is:
	// https://podcasts.apple.com/us/podcast/{name}/id{id}

	// extract the last component from the URL path
	ss := strings.Split(u.Path, "/")
	podcastID := ss[len(ss)-1]

	// and check that it has the expected format
	if !strings.HasPrefix(podcastID, "id") {
		err = fmt.Errorf("unexpected URL format for '%s'", link)
		return "", err
	}

	id := strings.TrimPrefix(podcastID, "id")

	return fmt.Sprintf(lookupTemplate, id), nil
}

type PodcastInfo struct {
	Name string
	Feed string
}

func extractFeedURLs(r io.Reader) ([]PodcastInfo, error) {
	body, err := ioutil.ReadAll(r)
	if err != nil {
		err = fmt.Errorf("cannot read podcast information: %w", err)
		return nil, err
	}

	var info struct {
		Results []struct {
			TrackName string `json:"trackName"`
			FeedURL   string `json:"feedURL"`
		} `json:"results"`
	}

	if err := json.Unmarshal(body, &info); err != nil {
		err = fmt.Errorf("cannot extract podcast information: %w", err)
		return nil, err
	}

	feedURLs := make([]PodcastInfo, 0, len(info.Results))
	for _, result := range info.Results {
		feedURLs = append(feedURLs, PodcastInfo{result.TrackName, result.FeedURL})
	}

	return feedURLs, nil
}
