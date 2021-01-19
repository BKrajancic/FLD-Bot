package command

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"strings"
	"time"

	"math"
	"net/url"
	"regexp"

	"github.com/BKrajancic/FLD-Bot/m/v2/src/service"
	"github.com/BKrajancic/FLD-Bot/m/v2/src/storage"
	"github.com/BKrajancic/FLD-Bot/m/v2/src/utils"
	"github.com/PuerkitoBio/goquery"
)

// GoQueryScraperConfig can be turned into a scraper that uses GoQuery.
type GoQueryScraperConfig struct {
	Title         string          // When sending a post, what should the title be.
	Trigger       string          // Word which triggers this command to activate.
	Capture       string          // How to capture words.
	TitleSelector SelectorCapture // Regex captures for title replacement.
	URL           string          // A url to scrape from, can contain one "%s" which is replaced with the first capture group.
	ReplySelector SelectorCapture
	Help          string // Help message to display
}

// SelectorCapture is a method to capture from a webpage.
type SelectorCapture struct {
	Template       string              // Message template to be filled out.
	Selectors      []string            // What captures to use to fill out the template
	HandleMultiple string              // How to handle multiple captures. "Random" or "First."
	Replacements   []map[string]string // Replacements for each entry in selectors.
}

// A HTMLGetter returns a url and buffer based on a string when err is nil.
type HTMLGetter = func(string) (url string, out io.ReadCloser, err error)

// ToStringWithMap uses a map to fill out the template.
// HandleMultiple is completely ignored.
// If a key is missing, it is skipped.
func (s SelectorCapture) ToStringWithMap(dict map[string]string) (out string, err error) {
	out = s.Template
	for i, selector := range s.Selectors {
		val, ok := dict[selector]
		if ok {
			if s.Replacements != nil && i < len(s.Replacements) {
				for search, replace := range s.Replacements[i] {
					val = strings.ReplaceAll(val, search, replace)
				}
			}
			out = fmt.Sprintf(out, val)
		}
	}

	return out, err
}

// Match all selectors and fill out template. Then using HandleMultiple decide which to use.
func selectorCaptureToString(doc goquery.Document, selectorCapture SelectorCapture) string {
	var maxLength int64 = math.MaxInt64
	allCaptures := make([](*(goquery.Selection)), len(selectorCapture.Selectors))
	if len(selectorCapture.Selectors) == 0 {
		maxLength = 0
	} else {
		for i, selector := range selectorCapture.Selectors {
			capture := doc.Find(selector)
			allCaptures[i] = capture
			captureLength := int64(capture.Length())

			if captureLength < maxLength {
				maxLength = captureLength
			}
		}
	}

	reply := selectorCapture.Template

	if maxLength == 0 && strings.Contains(reply, "%s") {
		return "There was an error retrieving information from the webpage."
	} else if maxLength > 0 {
		maxLength--

		var index int = 0
		if maxLength > 0 {
			if selectorCapture.HandleMultiple == "Random" {
				rand.Seed(time.Now().UnixNano())
				index = int(rand.Int63n(maxLength))
			} else if selectorCapture.HandleMultiple == "Last" {
				index = int(maxLength)
			}
		}

		tmp := make([]interface{}, len(selectorCapture.Selectors))
		for i, selector := range allCaptures {
			selectorIndex := selector.Slice(int(index), int(index)+1)
			val := strings.TrimSpace(selectorIndex.Text())
			if i < len(selectorCapture.Replacements) {
				for search, replace := range selectorCapture.Replacements[i] {
					if strings.Contains(val, search) {
						val = strings.ReplaceAll(val, search, replace)
						break
					}
				}
			}
			tmp[i] = val
		}

		reply = fmt.Sprintf(reply, tmp...)
	}
	return reply
}

// GetWebScraper returns a webscraper command from a config.
func (g GoQueryScraperConfig) GetWebScraper() (Command, error) {
	return GetGoqueryScraperWithHTMLGetter(g, utils.HTMLGetWithHTTP)
}

// GetGoqueryScraperWithHTMLGetter makes a scraper from a config.
func GetGoqueryScraperWithHTMLGetter(config GoQueryScraperConfig, htmlGetter HTMLGetter) (Command, error) {
	curry := func(sender service.Conversation, user service.User, msg [][]string, storage *storage.Storage, sink func(service.Conversation, service.Message)) {
		goqueryScraper(
			config,
			sender,
			user,
			msg,
			storage,
			sink,
			htmlGetter,
		)
	}

	regex, err := regexp.Compile(config.Capture)
	if err != nil {
		return Command{}, err
	}

	return Command{
		Trigger: config.Trigger,
		Pattern: regex,
		Exec:    curry,
		Help:    config.Help,
	}, nil
}

// Return the received message
func goqueryScraper(goQueryScraperConfig GoQueryScraperConfig, sender service.Conversation, user service.User, msg [][]string, storage *storage.Storage, sink func(service.Conversation, service.Message), htmlGetter HTMLGetter) {
	substitutions := strings.Count(goQueryScraperConfig.URL, "%s")
	if (substitutions > 0) && (msg == nil || len(msg) == 0 || len(msg[0]) < substitutions) {
		sink(sender, service.Message{Description: "An error occurred when building the url."})
		return
	}

	fields := make([]service.MessageField, 0)
	for _, capture := range msg {
		msgURL := goQueryScraperConfig.URL

		for _, word := range capture[1:] {
			msgURL = fmt.Sprintf(msgURL, url.PathEscape(word))
		}

		redirect, htmlReader, err := htmlGetter(msgURL)
		if err == nil {
			defer htmlReader.Close()
			doc, err := goquery.NewDocumentFromReader(htmlReader)
			if err == nil {
				if doc.Text() == "" {
					fields = append(fields, service.MessageField{
						Field: msgURL,
						Value: fmt.Sprintf("Webpage not found at %s", redirect),
						URL:   "",
					})
				} else {
					msgCapture := selectorCaptureToString(*doc, goQueryScraperConfig.ReplySelector)
					fields = append(fields, service.MessageField{
						Field: selectorCaptureToString(*doc, goQueryScraperConfig.TitleSelector),
						Value: msgCapture,
						URL:   redirect,
					})
				}
			} else {
				fields = append(fields, service.MessageField{
					Field: msgURL,
					Value: "An error occurred when processing the webpage.",
				})
			}
		} else {
			fields = append(fields, service.MessageField{
				Field: "Error",
				Value: "An error occurred retrieving the webpage.",
				URL:   msgURL,
			})
		}
	}

	sink(sender, service.Message{
		Title:       goQueryScraperConfig.Title,
		Description: "",
		Fields:      fields,
	})
}

// GetGoqueryScraperConfigs retrieves an array of GoQueryScraperConfig from a json file.
// If a file doesn't exist, an example is made in its place, and an error is returned.
func GetGoqueryScraperConfigs(reader io.Reader) ([]GoQueryScraperConfig, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	var config []GoQueryScraperConfig
	return config, json.Unmarshal(bytes, &config)
}
