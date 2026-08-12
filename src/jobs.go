package src

import (
	"coolifymanager/src/config"
	"coolifymanager/src/database"
	"fmt"
	"strings"

	td "github.com/AshokShau/gotdbot"
)

const pageSize = 5

func jobsHandler(c *td.Client, msg *td.Message) error {
	if !config.IsDev(msg.SenderID()) {
		_, err := msg.ReplyText(c, "🚫 ʏᴏᴜ ᴀʀᴇ ɴᴏᴛ ᴀᴜᴛʜᴏʀɪᴢᴇᴅ ᴛᴏ ᴜsᴇ ᴛʜɪs ᴄᴏᴍᴍᴀɴᴅ.", nil)
		return err
	}

	text, kb, err := buildJobsMessage(1)
	if err != nil {
		_, err = msg.ReplyText(c, "❌ "+err.Error(), nil)
		return err
	}

	_, err = msg.ReplyText(c, text, &td.SendTextMessageOpts{ParseMode: "HTML", ReplyMarkup: kb})
	return err
}

func jobsPaginationHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	data := cb.DataString()
	if !config.IsDev(cb.SenderUserId) {
		_ = cb.Answer(c, 0, true, "🚫 ʏᴏᴜ ᴀʀᴇ ɴᴏᴛ ᴀᴜᴛʜᴏʀɪᴢᴇᴅ.", "")
		return nil
	}

	page := 1
	if parts := strings.Split(data, ":"); len(parts) > 1 {
		fmt.Sscanf(parts[1], "%d", &page)
	}

	text, kb, err := buildJobsMessage(page)
	if err != nil {
		_ = cb.Answer(c, 0, true, "❌ "+err.Error(), "")
		return nil
	}

	_, err = cb.EditMessageText(c, text, &td.EditTextMessageOpts{ParseMode: "HTML", ReplyMarkup: kb})
	return err
}

func buildJobsMessage(page int) (string, td.ReplyMarkup, error) {
	tasks, err := database.GetTasks()
	if err != nil {
		return "", nil, fmt.Errorf("error fetching tasks: %v", err)
	}

	if len(tasks) == 0 {
		return "📭 ɴᴏ sᴄʜᴇᴅᴜʟᴇᴅ ᴊᴏʙs ғᴏᴜɴᴅ.", nil, nil
	}

	start, end, buttons := Paginate(len(tasks), page, pageSize, "jobs:")

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<b>📅 sᴄʜᴇᴅᴜʟᴇᴅ ᴊᴏʙs (Page %d):</b>\n\n", page))

	for _, task := range tasks[start:end] {
		sb.WriteString(fmt.Sprintf("🆔 <code>%s</code>\n", task.ID.Hex()))
		sb.WriteString(fmt.Sprintf("🏷️ <b>ɴᴀᴍᴇ:</b> %s\n", task.Name))
		sb.WriteString(fmt.Sprintf("🔧 <b>ᴛʏᴘᴇ:</b> %s\n", task.Type))
		sb.WriteString(fmt.Sprintf("⏰ <b>sᴄʜᴇᴅᴜʟᴇ:</b> %s\n", task.Schedule))
		if task.OneTime {
			sb.WriteString(fmt.Sprintf("⏳ <b>ɴᴇxᴛ ʀᴜɴ:</b> %s\n", task.NextRun.Format("2006-01-02 15:04:05")))
		}
		sb.WriteString("➖➖➖➖➖➖➖➖➖➖\n")
	}

	kb := &td.ReplyMarkupInlineKeyboard{}
	if len(buttons) > 0 {
		row := make([]td.InlineKeyboardButton, 0, len(buttons))

		for _, btn := range buttons {
			row = append(row, td.InlineKeyboardButton{
				Text: btn.Text,
				Type: &td.InlineKeyboardButtonTypeCallback{
					Data: []byte(btn.Data),
				},
			})
		}

		kb.Rows = append(kb.Rows, row)
	}

	return sb.String(), kb, nil
}
