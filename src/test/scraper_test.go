package test

import (
	"strings"
	"testing"

	"github.com/BKrajancic/FLD-Bot/m/v2/src/bot"
	"github.com/BKrajancic/FLD-Bot/m/v2/src/command"
	"github.com/BKrajancic/FLD-Bot/m/v2/src/service"
	"github.com/BKrajancic/FLD-Bot/m/v2/src/service/demo_service"
)

func TestScraperWithCapture(t *testing.T) {
	bot := bot.Bot{}

	demoServiceSender := demo_service.DemoServiceSender{}
	bot.AddSender(&demoServiceSender)

	testConversation := service.Conversation{
		ServiceId:      demoServiceSender.Id(),
		ConversationId: "0",
	}

	testSender := service.User{Name: "Test_User", Id: demoServiceSender.Id()}
	testCmd := "!scrape"

	config := command.ScraperConfig{
		Command:       testCmd + " (.*)",
		URL:           "https://webscraper.io/test-sites/%s",
		Reply_capture: "<h1>([^<]*)</h1>",
	}

	scraper, err := command.GetScraper(config)
	if err != nil {
		t.Errorf("An error occured when making a reasonable scraper!")
	}

	bot.AddCommand(scraper)
	bot.OnMessage(testConversation, testSender, testCmd+" e-commerce/allinone")

	resultMessage, resultConversation := demoServiceSender.PopMessage()
	if !strings.HasPrefix(resultMessage.Description, "Test Sites E-commerce training site") {
		t.Errorf("Message was different!")
	}

	if resultConversation != testConversation {
		t.Errorf("Sender was different!")
	}

	bot.OnMessage(testConversation, testSender, testCmd+" tables")
	resultMessage, resultConversation = demoServiceSender.PopMessage()
	if !strings.HasPrefix(resultMessage.Description, "Table playground") {
		t.Errorf("Message was different!")
	}

	if resultConversation != testConversation {
		t.Errorf("Sender was different!")
	}
}

func TestScraperWithCaptureAndNoTitleCapture(t *testing.T) {
	bot := bot.Bot{}

	demoServiceSender := demo_service.DemoServiceSender{}
	bot.AddSender(&demoServiceSender)

	testConversation := service.Conversation{
		ServiceId:      demoServiceSender.Id(),
		ConversationId: "0",
	}

	testSender := service.User{Name: "Test_User", Id: demoServiceSender.Id()}
	testCmd := "!scrape"

	config := command.ScraperConfig{
		Command:        testCmd + " (.*)",
		URL:            "https://webscraper.io/test-sites/%s",
		Reply_capture:  "<h1>([^<]*)</h1>",
		Title_template: "Title",
	}

	scraper, err := command.GetScraper(config)
	if err != nil {
		t.Errorf("An error occured when making a reasonable scraper!")
	}

	bot.AddCommand(scraper)
	bot.OnMessage(testConversation, testSender, testCmd+" e-commerce/allinone")

	resultMessage, resultConversation := demoServiceSender.PopMessage()
	if resultMessage.Title != config.Title_template {
		t.Errorf("Title was different!")
	}

	if resultConversation != testConversation {
		t.Errorf("Sender was different!")
	}
}

func TestScraperWithTitleCapture(t *testing.T) {
	bot := bot.Bot{}

	demoServiceSender := demo_service.DemoServiceSender{}
	bot.AddSender(&demoServiceSender)

	testConversation := service.Conversation{
		ServiceId:      demoServiceSender.Id(),
		ConversationId: "0",
	}

	testSender := service.User{Name: "Test_User", Id: demoServiceSender.Id()}
	testCmd := "!scrape"

	config := command.ScraperConfig{
		Command:        testCmd + " (.*)",
		URL:            "https://webscraper.io/test-sites/%s",
		Reply_capture:  "<h1>([^<]*)</h1>",
		Title_template: "%s",
		Title_capture:  "<h2>([^<]*)</h2>",
	}

	scraper, err := command.GetScraper(config)
	if err != nil {
		t.Errorf("An error occured when making a reasonable scraper!")
	}

	bot.AddCommand(scraper)
	bot.OnMessage(testConversation, testSender, testCmd+" e-commerce/allinone")

	resultMessage, resultConversation := demoServiceSender.PopMessage()
	if resultMessage.Title != "Top items being scraped right now" {
		t.Errorf("Title was different!")
	}

	if resultConversation != testConversation {
		t.Errorf("Sender was different!")
	}
}

func TestScraperNoCapture(t *testing.T) {
	bot := bot.Bot{}

	demoServiceSender := demo_service.DemoServiceSender{}
	bot.AddSender(&demoServiceSender)

	testConversation := service.Conversation{
		ServiceId:      demoServiceSender.Id(),
		ConversationId: "0",
	}

	testSender := service.User{Name: "Test_User", Id: demoServiceSender.Id()}
	testCmd := "!scrape"

	config := command.ScraperConfig{
		Command:       testCmd,
		URL:           "https://webscraper.io/test-sites/e-commerce/allinone",
		Reply_capture: "<h1>([^<]*)</h1>",
	}

	scraper, err := command.GetScraper(config)
	if err != nil {
		t.Errorf("An error occured when making a reasonable scraper!")
	}

	bot.AddCommand(scraper)
	bot.OnMessage(testConversation, testSender, testCmd)

	resultMessage, resultConversation := demoServiceSender.PopMessage()
	if !strings.HasPrefix(resultMessage.Description, "Test Sites E-commerce training site") {
		t.Errorf("Message was different!")
	}

	if resultConversation != testConversation {
		t.Errorf("Sender was different!")
	}
}
