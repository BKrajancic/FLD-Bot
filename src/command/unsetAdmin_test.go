package command

import (
	"testing"

	"github.com/BKrajancic/FLD-Bot/m/v2/src/service"
	"github.com/BKrajancic/FLD-Bot/m/v2/src/service/demoservice"
	"github.com/BKrajancic/FLD-Bot/m/v2/src/storage"
)

func TestUnsetAdmin(t *testing.T) {
	demoSender := demoservice.DemoSender{ServiceID: demoservice.ServiceID}
	tempStorage := storage.TempStorage{}
	var _storage storage.Storage = &tempStorage

	testSender := service.User{Name: "Test_User", ID: demoSender.ID()}
	guild := service.Guild{ServiceID: demoSender.ID(), GuildID: "0"}
	userID := "0"
	tempStorage.SetAdmin(guild, userID)
	testConversation := service.Conversation{
		ServiceID:      demoSender.ID(),
		ConversationID: "0",
		GuildID:        guild.GuildID,
		Admin:          true,
	}

	UnsetAdmin(testConversation, testSender, [][]string{{"", userID}}, &_storage, demoSender.SendMessage)

	if _storage.IsAdmin(guild, userID) {
		t.Errorf("Admin should be able to unset admins")
	}
}

func TestDontUnsetAdmin(t *testing.T) {
	demoSender := demoservice.DemoSender{ServiceID: demoservice.ServiceID}
	tempStorage := storage.TempStorage{}
	var _storage storage.Storage = &tempStorage

	testSender := service.User{Name: "Test_User", ID: demoSender.ID()}
	guild := service.Guild{ServiceID: demoSender.ID(), GuildID: "0"}
	userID := "0"
	tempStorage.SetAdmin(guild, userID)
	testConversation := service.Conversation{
		ServiceID:      demoSender.ID(),
		ConversationID: "0",
		GuildID:        guild.GuildID,
		Admin:          false,
	}

	UnsetAdmin(testConversation, testSender, [][]string{}, &_storage, demoSender.SendMessage)

	if _storage.IsAdmin(guild, userID) == false {
		t.Errorf("Non-Admin should not be able to unset")
	}
}