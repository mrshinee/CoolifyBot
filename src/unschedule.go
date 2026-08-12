package src

import (
	"coolifymanager/src/config"
	"coolifymanager/src/database"
	"coolifymanager/src/scheduler"
	"fmt"
	"strings"

	td "github.com/AshokShau/gotdbot"
)

func unscheduleHandler(c *td.Client, msg *td.Message) error {
	if !config.IsDev(msg.SenderID()) {
		_, err := msg.ReplyText(c, "🚫 ʏᴏᴜ ᴀʀᴇ ɴᴏᴛ ᴀᴜᴛʜᴏʀɪᴢᴇᴅ ᴛᴏ ᴜꜱᴇ ᴛʜɪꜱ ᴄᴏᴍᴍᴀɴᴅ.", nil)
		return err
	}

	args := strings.Fields(msg.Text())
	if len(args) < 2 {
		_, err := msg.ReplyText(c, "usage: /unschedule <task_id>", nil)
		return err
	}
	taskID := args[1]

	if err := scheduler.RemoveTask(taskID); err != nil {
		_, err = msg.ReplyText(c, fmt.Sprintf("⚠️ ᴡᴀʀɴɪɴɢ: ᴄᴏᴜʟᴅ ɴᴏᴛ ʀᴇᴍᴏᴠᴇ ᴛᴀꜱᴋ ꜰʀᴏᴍ ꜱᴄʜᴇᴅᴜʟᴇʀ (ᴍɪɢʜᴛ ɴᴏᴛ ʙᴇ ʀᴜɴɴɪɴɢ): %v", err), nil)
	}

	if err := database.DeleteTask(taskID); err != nil {
		_, err = msg.ReplyText(c, fmt.Sprintf("❌ ᴇʀʀᴏʀ ᴅᴇʟᴇᴛɪɴɢ ᴛᴀꜱᴋ ꜰʀᴏᴍ ᴅᴀᴛᴀʙᴀꜱᴇ: %v", err), nil)
		return err
	}

	_, err := msg.ReplyText(c, fmt.Sprintf("✅ ᴛᴀꜱᴋ <code>%s</code> ʀᴇᴍᴏᴠᴇᴅ ꜱᴜᴄᴄᴇꜱꜱꜰᴜʟʟʏ.", taskID), &td.SendTextMessageOpts{ParseMode: "HTML"})
	return err
}
