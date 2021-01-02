package command

import (
	"github.com/BKrajancic/FLD-Bot/m/v2/src/service"
	"github.com/BKrajancic/FLD-Bot/m/v2/src/storage"
)

// SetAdmin will set the value to be an admin. This is AS-IS, it is up to a service to handle it.
func UnsetAdmin(sender service.Conversation, user service.User, msg [][]string, storage *storage.Storage, sink func(service.Conversation, service.Message)) {
	if sender.Admin {
		guild := service.Guild{
			ServiceId: sender.ServiceId,
			GuildID:   sender.GuildID,
		}

		(*storage).UnsetAdmin(guild, msg[0][1])
		sink(sender, service.Message{Description: "Admin has been unset."})
	}
}