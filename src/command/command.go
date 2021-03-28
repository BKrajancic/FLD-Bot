// Package command includes actions that users can trigger by prefixing a message with a string.
package command

import (
	"fmt"
	"strings"
	"strconv"

	"github.com/BKrajancic/boby/m/v2/src/service"
	"github.com/BKrajancic/boby/m/v2/src/storage"
)

// A Command is how a User interacts with a bot.
type Command struct {
	Trigger       string             // Messages starting with Trigger are processed by this Command.
	Parameters    []CommandParameter // What text to capture following a trigger.
	Help          string             // What this command does.
	HelpInput     string             // Arguments following the trigger.
	Exec          func(service.Conversation, service.User, []interface{}, *storage.Storage, func(service.Conversation, service.Message))
	Storage       *storage.Storage
	observers     []service.Sender
}

type CommandParameter struct {
	Type         string
	Name         string
	Description  string
}

var INPUTS = map[string]func(string) (interface{}, error){
	"string": func (input string) (interface{}, error) {
		return input, nil
	},

	"int": func (input string) (interface{}, error) {
		val, err := strconv.Atoi(input)
		if err != nil {
			return nil, err
		}
		return val, nil
	},

	"bool": func (input string) (interface{}, error) {
		if strings.ToLower(input) == "true"{
			return true, nil
		}

		if strings.ToLower(input) == "false"{
			return false, nil
		}

		return false, fmt.Errorf("parsing error")
	},
}

// AddSender will append a sender that output messages are routed to.
func (c CommandParameter) ParseInput(input string) (interface{}, error) {
	return INPUTS[c.Type](input)
}

// AddSender will append a sender that output messages are routed to.
func (c *Command) AddSender(sender service.Sender) {
	c.observers = append(c.observers, sender)
}

// RouteByID routes a message to an observer of this Bot with the same ID() as
// conversation.ServiceID.
func (c *Command) RouteByID(conversation service.Conversation, msg service.Message) {
	for _, observer := range c.observers {
		if observer.ID() == conversation.ServiceID {
			observer.SendMessage(conversation, msg)
		}
	}
}

/*
This is a subject responsibility.
// OnMessage checks if a message begins with a prefix, and if so, calls Exec.
func (c *Command) OnMessage(conversation service.Conversation, source service.User, msg string) {
	prefix, ok := (*c.Storage).GetGuildValue(conversation.Guild(), "prefix")
	if ok != true {
		return
	}

	trigger := fmt.Sprintf("%s%s", prefix, c.Trigger)
	if strings.HasPrefix(msg, trigger) {
		content := strings.TrimSpace(msg[len(trigger):])
		newMatches := make([]interface{}, 0)

		for _, match := range c.Pattern.FindAllStringSubmatch(content, -1) {
			if len(match) > 1 {
				newMatches = append(newMatches, match[1:])
			}
		}

		c.Exec(
			conversation,
			source,
			newMatches,
			c.Storage,
			c.RouteByID,
		)
	}
}
*/
