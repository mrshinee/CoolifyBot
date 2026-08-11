package src

import (
	"coolifymanager/src/config"
	"coolifymanager/src/database"
	"coolifymanager/src/scheduler"
	"fmt"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
)

func unscheduleHandler(m *telegram.NewMessage) error {
	if !config.IsDev(m.Sender.ID) {
		_, err := m.Reply("🚫 ʏᴏᴜ ᴀʀᴇ ɴᴏᴛ ᴀᴜᴛʜᴏʀɪᴢᴇᴅ ᴛᴏ ᴜꜱᴇ ᴛʜɪꜱ ᴄᴏᴍᴍᴀɴᴅ.")
		return err
	}

	args := strings.Fields(m.Text())
	if len(args) < 2 {
		_, err := m.Reply("usage: /unschedule <task_id>")
		return err
	}

	taskID := args[1]

	if err := scheduler.RemoveTask(taskID); err != nil {
		_, err = m.Reply(fmt.Sprintf("⚠️ ᴡᴀʀɴɪɴɢ: ᴄᴏᴜʟᴅ ɴᴏᴛ ʀᴇᴍᴏᴠᴇ ᴛᴀꜱᴋ ꜰʀᴏᴍ ꜱᴄʜᴇᴅᴜʟᴇʀ (ᴍɪɢʜᴛ ɴᴏᴛ ʙᴇ ʀᴜɴɴɪɴɢ): %v", err))
	}
	
	if err := database.DeleteTask(taskID); err != nil {
		_, err = m.Reply(fmt.Sprintf("❌ ᴇʀʀᴏʀ ᴅᴇʟᴇᴛɪɴɢ ᴛᴀꜱᴋ ꜰʀᴏᴍ ᴅᴀᴛᴀʙᴀꜱᴇ: %v", err))
		return err
	}

	_, err := m.Reply(fmt.Sprintf("✅ ᴛᴀꜱᴋ <code>%s</code> ʀᴇᴍᴏᴠᴇᴅ ꜱᴜᴄᴄᴇꜱꜱꜰᴜʟʟʏ.", taskID))
	return err
}
