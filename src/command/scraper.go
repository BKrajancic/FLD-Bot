package command

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"net/http"
	"regexp"

	"github.com/BKrajancic/FLD-Bot/m/v2/src/service"
)

type ScraperConfig struct {
	Command string // Regular expression which triggers this scraper. Can contain capture groups.
	Url     string // A url to scrape from, can contain one "%s" which is replaced with the first capture group.
	Re      string // Regular expression used to parse a webpage.
}

// Get the scraper for this module
func GetScraper(filepath string) (func(service.Conversation, service.User, [][]string, func(service.Conversation, string)), string, error) {

	var config ScraperConfig

	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		bytes, err := json.Marshal(config)

		if err != nil {
			return nil, "", errors.New("Unable to create example JSON")
		}

		file, err := os.Create(filepath)
		if err != nil {
			return nil, "", errors.New(fmt.Sprintf("Unable to create file: %s", filepath))
		}
		defer file.Close()

		_, err = file.Write(bytes)
		if err != nil {
			return nil, "", errors.New(fmt.Sprintf("Unable to write to file: %s", filepath))
		}
		return nil, "", errors.New(fmt.Sprintf("File %s did not exist, an example has been writen.", filepath))
	}
	bytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		fmt.Printf("Unable to read file: %s", filepath)
		return nil, "", err
	}

	json.Unmarshal(bytes, &config)

	curry := func(sender service.Conversation, user service.User, msg [][]string, sink func(service.Conversation, string)) {
		scraper(config.Url, config.Re, sender, user, msg, sink)
	}

	return curry, config.Command, nil
}

// Return the received message
func scraper(url_template string, re_s string, sender service.Conversation, user service.User, msg [][]string, sink func(service.Conversation, string)) {
	url := fmt.Sprintf(url_template, msg[0][1])
	re := regexp.MustCompile(re_s)
	msg_template := "%s.\nRead more at: %s"

	response, err := http.Get(url)
	if err == nil {
		// Read response data in to memory
		body, err := ioutil.ReadAll(response.Body)

		if err == nil {
			// Create a regular expression to find comments
			body_s := string(body)

			matches := re.FindAllStringSubmatch(body_s, -1)

			if matches != nil {
				all_captures := make([]string, len(matches))
				for i, captures := range matches {
					all_captures[i] = strings.Join(captures[1:], " ")
				}

				reply := fmt.Sprintf(msg_template, strings.Join(all_captures, " "), url)
				sink(sender, reply)
			} else {
				sink(sender, "The webpage was not found.")
			}
		} else {
			sink(sender, "An error occurred when processing the webpage.")
		}
	} else {
		sink(sender, "An error occurred retrieving the webpage.")
	}
}
