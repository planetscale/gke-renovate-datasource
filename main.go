package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	// Regex to extract version numbers from hyperlinks
	versionRegex = regexp.MustCompile(`<a href="[^"]+">([0-9]+\.[0-9]+\.[0-9]+-gke\.[0-9]+)</a>`)

	// Markers to identify available and unavailable versions in the HTML blob of
	// each RSS feed <entry>
	availableMarker   = "The following versions are now available in the"
	unavailableMarker = "The following versions are no longer available in the"

	// Channel RSS feed URLs
	channelURLs = map[string]string{
		"stable":  "https://cloud.google.com/feeds/gke-stable-channel-release-notes.xml",
		"regular": "https://cloud.google.com/feeds/gke-regular-channel-release-notes.xml",
		"rapid":   "https://cloud.google.com/feeds/gke-rapid-channel-release-notes.xml",
	}
)

type RenovateCustomDatasource struct {
	Releases []RenovateRelease `json:"releases"`
}

type RenovateRelease struct {
	Version          string `json:"version"`
	ReleaseTimestamp string `json:"releaseTimestamp"`
}

type Feed struct {
	XMLName xml.Name `xml:"feed"`
	Entries []Entry  `xml:"entry"`
}

type Entry struct {
	Title   string    `xml:"title"`
	Updated time.Time `xml:"updated"`
	Content Content   `xml:"content"`
}

type Content struct {
	Data string `xml:",cdata"`
}

func main() {
	channel := flag.String("channel", "stable", "GKE release channel (stable, regular, rapid)")
	outFile := flag.String("out", "", "Output JSON file (required)")
	flag.Parse()

	if *outFile == "" {
		fmt.Fprintf(os.Stderr, "Error: -out flag is required\n")
		os.Exit(1)
	}

	url, ok := channelURLs[*channel]
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: invalid channel. Must be one of: stable, regular, rapid\n")
		os.Exit(1)
	}

	feed, err := fetchFeed(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching feed: %v\n", err)
		os.Exit(1)
	}

	output := processEntries(feed)

	file, err := os.Create(*outFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		os.Exit(1)
	}
}

func fetchFeed(url string) (*Feed, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var feed Feed
	if err := xml.Unmarshal(data, &feed); err != nil {
		return nil, err
	}

	return &feed, nil
}

func processEntries(feed *Feed) *RenovateCustomDatasource {
	output := &RenovateCustomDatasource{
		Releases: make([]RenovateRelease, 0),
	}

	for _, entry := range feed.Entries {
		versions := extractAvailableVersions(entry.Content.Data)
		timestamp := entry.Updated.Format(time.RFC3339)

		for _, version := range versions {
			release := RenovateRelease{
				Version:          version,
				ReleaseTimestamp: timestamp,
			}
			output.Releases = append(output.Releases, release)
		}
	}

	return output
}

func extractAvailableVersions(content string) []string {
	parts := strings.Split(content, unavailableMarker)
	if len(parts) < 1 {
		return nil
	}

	availableParts := strings.Split(parts[0], availableMarker)
	if len(availableParts) < 2 {
		return nil
	}

	availableSection := availableParts[1]
	matches := versionRegex.FindAllStringSubmatch(availableSection, -1)

	versions := make([]string, 0)
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) > 1 {
			version := match[1]
			if !seen[version] {
				versions = append(versions, version)
				seen[version] = true
			}
		}
	}

	return versions
}
