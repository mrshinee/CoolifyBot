package src

import (
	"coolifymanager/src/config"
	"coolifymanager/src/database"
	"fmt"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
)

const pageSize = 5

func jobsHandler(m *telegram.NewMessage) error {
	if !config.IsDev(m.Sender.ID) {
		_, err := m.Reply("🚫 ʏᴏᴜ ᴀʀᴇ ɴᴏᴛ ᴀᴜᴛʜᴏʀɪᴢᴇᴅ ᴛᴏ ᴜꜱᴇ ᴛʜɪꜱ ᴄᴏᴍᴍᴀɴᴅ.")
		return err
	}

	text, kb, err := buildJobsMessage(1)
	if err != nil {
		_, err = m.Reply("❌ " + err.Error())
		return err
	}

	_, err = m.Reply(text, &telegram.SendOptions{ParseMode: "HTML", ReplyMarkup: kb})
	return err
}

func jobsPaginationHandler(cb *telegram.CallbackQuery) error {
	if !config.IsDev(cb.SenderID) {
		_, _ = cb.Answer("🚫 ʏᴏᴜ ᴀʀᴇ ɴᴏᴛ ᴀᴜᴛʜᴏʀɪᴢᴇᴅ.", &telegram.CallbackOptions{Alert: true})
		return nil
	}

	page := 1
	data := cb.DataString()
	if parts := strings.Split(data, ":"); len(parts) > 1 {
		fmt.Sscanf(parts[1], "%d", &page)
	}

	text, kb, err := buildJobsMessage(page)
	if err != nil {
		_, _ = cb.Answer("ᴇʀʀᴏʀ: "+err.Error(), &telegram.CallbackOptions{Alert: true})
		return nil
	}

	_, err = cb.Edit(text, &telegram.SendOptions{ParseMode: "HTML", ReplyMarkup: kb})
	return err
}

func buildJobsMessage(page int) (string, telegram.ReplyMarkup, error) {
	tasks, err := database.GetTasks()
	if err != nil {
		return "", nil, fmt.Errorf("error fetching tasks: %v", err)
	}

	if len(tasks) == 0 {
		return "📭 ɴᴏ ꜱᴄʜᴇᴅᴜʟᴇᴅ ᴊᴏʙꜱ ꜰᴏᴜɴᴅ.", nil, nil
	}

	start, end, buttons := Paginate(len(tasks), page, pageSize, "jobs:")

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<b>📅 ꜱᴄʜᴇᴅᴜʟᴇᴅ ᴊᴏʙꜱ (ᴘᴀɢᴇ %d):</b>\n\n", page))

	for _, task := range tasks[start:end] {
		sb.WriteString(fmt.Sprintf("🆔 <code>%s</code>\n", task.ID.Hex()))
		sb.WriteString(fmt.Sprintf("🏷️ <b>ɴᴀᴍᴇ:</b> %s\n", task.Name))
		sb.WriteString(fmt.Sprintf("🔧 <b>ᴛʏᴘᴇ:</b> %s\n", task.Type))
		sb.WriteString(fmt.Sprintf("⏰ <b>ꜱᴄʜᴇᴅᴜʟᴇ:</b> %s\n", task.Schedule))
		if task.OneTime {
			sb.WriteString(fmt.Sprintf("⏳ <b>ɴᴇxᴛ ʀᴜɴ:</b> %s\n", task.NextRun.Format("2006-01-02 15:04:05")))
		}
		sb.WriteString("➖➖➖➖➖➖➖➖➖➖\n")
	}

	kb := telegram.NewKeyboard()
	if len(buttons) > 0 {
		var row []telegram.KeyboardButton
		for _, btn := range buttons {
			row = append(row, telegram.Button.Data(btn.Text, btn.Data))
		}
		kb.AddRow(row...)
	}

	return sb.String(), kb.Build(), nil
}
