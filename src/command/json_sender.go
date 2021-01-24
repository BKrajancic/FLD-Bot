package command

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/BKrajancic/boby/m/v2/src/service"
	"github.com/BKrajancic/boby/m/v2/src/storage"
)

// JSONGetterConfig can be used to extract from JSON into a message.
type JSONGetterConfig struct {
	Trigger     string
	Capture     string
	Title       FieldCapture
	Captures    []JSONCapture
	URL         string
	Help        string
	Description string
	Grouped     bool
	Delay       int
	Token       TokenMaker
	RateLimit   RateLimitConfig
}

// A TokenMaker is useful for creating a token that may be part of an API request.
type TokenMaker struct {
	Type    string // Can be MD5
	Prefix  string // When calculating a token, what should be prepended
	Postfix string // When calculating a token, what should be appended
	Size    int    // Take the first 'Size' characters from the result.
}

// MakeToken will make a token from this token maker.
// If t.type is "MD5"
func (t TokenMaker) MakeToken(input string) (out string) {
	if t.Type == "MD5" {
		fullString := []byte(t.Prefix + input + t.Postfix)
		return fmt.Sprintf("%x", md5.Sum(fullString))[:t.Size]
	}
	return ""
}

// A FieldCapture represents a template to be filled out by selectors.
// TODO Improve documentation here.
type FieldCapture struct {
	Template  string   // Message template to be filled out.
	Selectors []string // What captures to use to fill out the template
}

// ToStringWithMap uses a map to fill out the template.
// HandleMultiple is completely ignored.
// If a key is missing, it is skipped.
func (f FieldCapture) ToStringWithMap(dict map[string]interface{}) (out string, err error) {
	out = f.Template
	for _, selector := range f.Selectors {
		val, ok := dict[selector]
		if ok {
			out = fmt.Sprintf(out, val.(string))
		}
	}
	return out, err
}

// JSONCapture is a pair of FieldCapture to represent a title, body pair in a message.
type JSONCapture struct {
	Title FieldCapture
	Body  FieldCapture
}

// JSONGetter will accept a string and provide a reader. This could be a file, a webpage, who cares!
type JSONGetter = func(string) (out io.ReadCloser, err error)

// GetWebScraper returns a webscraper command from a config.
func (j JSONGetterConfig) GetWebScraper(jsonGetter JSONGetter) (Command, error) {
	curry := func(sender service.Conversation, user service.User, msg [][]string, storage *storage.Storage, sink func(service.Conversation, service.Message)) {
		jsonGetterFunc(
			j,
			sender,
			user,
			msg,
			storage,
			sink,
			jsonGetter,
		)
	}

	regex, err := regexp.Compile(j.Capture)
	if err != nil {
		return Command{}, err
	}

	return Command{
		Trigger: j.Trigger,
		Pattern: regex,
		Exec:    curry,
		Help:    j.Help,
	}, nil
}

// Return the received message
func jsonGetterFunc(config JSONGetterConfig, sender service.Conversation, user service.User, msg [][]string, storage *storage.Storage, sink func(service.Conversation, service.Message), jsonGetter JSONGetter) {
	substitutions := strings.Count(config.URL, "%s")
	if (substitutions > 0) && (msg == nil || len(msg) == 0 || len(msg[0]) < substitutions) {
		sink(sender, service.Message{Description: "An error occurred when building the url."})
		return
	}

	fields := make([]service.MessageField, 0)
	for _, capture := range msg {
		msgURL := config.URL

		replacements := strings.Count(msgURL, "%s")

		for i, word := range capture[1:] {
			if i == replacements {
				break
			}
			msgURL = strings.Replace(msgURL, "%s", url.PathEscape(word), 1)
		}

		msgURL += config.Token.MakeToken(strings.Join(capture[1:], ""))
		fmt.Print(msgURL)
		jsonReader, err := jsonGetter(msgURL)
		if err == nil {
			defer jsonReader.Close()
			buf, err := ioutil.ReadAll(jsonReader)
			if err == nil {
				dict := make(map[string]interface{})
				err := json.Unmarshal(buf, &dict)
				if err == nil {
					for _, capture := range config.Captures {
						body, err := capture.Body.ToStringWithMap(dict)
						if err == nil && strings.Contains(body, "%s") == false {
							title, err := capture.Title.ToStringWithMap(dict)
							if err == nil {
								fields = append(fields, service.MessageField{Field: title, Value: body})
							}
						}
					}
					if config.Grouped {
						title, err := config.Title.ToStringWithMap(dict)
						if err == nil {
							sink(sender, service.Message{
								Title:       title,
								Fields:      fields,
								Description: config.Description,
							})
						}
					} else {
						for _, field := range fields {
							sink(sender, service.Message{
								Title:       field.Field,
								Description: field.Value,
							})

							time.Sleep(time.Duration(config.Delay) * time.Second)
						}

					}
				}
			}
		}
	}

}
