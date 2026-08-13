package src

import (
	"coolifymanager/src/config"
	"coolifymanager/src/database"
	"coolifymanager/src/scheduler"
	"fmt"
	"os"
	"strings"

	td "github.com/AshokShau/gotdbot"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func listProjectsHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	if !config.IsDev(cb.SenderUserId) {
		_ = cb.Answer(c, 0, true, "🚫 ʏᴏᴜ ᴀʀᴇ ɴᴏᴛ ᴀᴜᴛʜᴏʀɪᴢᴇᴅ.", "")
		return nil
	}

	_ = cb.Answer(c, 0, false, "ᴘʀᴏᴄᴇꜱꜱɪɴɢ...", "")
	apps, err := config.Coolify.ListApplications()
	if err != nil {
		_, _ = cb.EditMessageText(c, "ꜰᴀɪʟᴇᴅ ᴛᴏ ꜰᴇᴛᴄʜ ᴘʀᴏᴊᴇᴄᴛꜱ:"+err.Error(), nil)
		return nil
	}

	if len(apps) == 0 {
		_, _ = cb.EditMessageText(c, "😶 ɴᴏ ᴀᴘᴘʟɪᴄᴀᴛɪᴏɴꜱ ꜰᴏᴜɴᴅ.", nil)
		return nil
	}

	page := 1
	cbData := cb.DataString()
	if strings.Contains(cbData, ":") {
		parts := strings.Split(cbData, ":")
		if len(parts) > 1 {
			fmt.Sscanf(parts[1], "%d", &page)
		}
	}

	start, end, paginationButtons := Paginate(len(apps), page, 7, "list_projects:")

	kb := &td.ReplyMarkupInlineKeyboard{}
	for _, app := range apps[start:end] {
		text := fmt.Sprintf("📦 %s (%s)", app.Name, app.Status)
		data := "project_menu:" + app.UUID

		kb.Rows = append(kb.Rows, []td.InlineKeyboardButton{
			{
				Text: text,
				Type: &td.InlineKeyboardButtonTypeCallback{
					Data: []byte(data),
				},
			},
		})
	}

	if len(paginationButtons) > 0 {
		row := make([]td.InlineKeyboardButton, 0, len(paginationButtons))

		for _, btn := range paginationButtons {
			row = append(row, td.InlineKeyboardButton{
				Text: btn.Text,
				Type: &td.InlineKeyboardButtonTypeCallback{
					Data: []byte(btn.Data),
				},
			})
		}

		kb.Rows = append(kb.Rows, row)
	}

	_, err = cb.EditMessageText(c, "<b>📋 ꜱᴇʟᴇᴄᴛ ᴀ ᴘʀᴏᴊᴇᴄᴛ:</b>", &td.EditTextMessageOpts{ParseMode: "HTML", ReplyMarkup: kb})
	return err
}

func projectMenuHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	if !config.IsDev(cb.SenderUserId) {
		_ = cb.Answer(c, 0, true, "🚫 ʏᴏᴜ ᴀʀᴇ ɴᴏᴛ ᴀᴜᴛʜᴏʀɪᴢᴇᴅ.", "")
		return nil
	}

	_ = cb.Answer(c, 0, false, "ᴘʀᴏᴄᴇꜱꜱɪɴɢ...", "")

	cbData := cb.DataString()
	uuid := strings.TrimPrefix(cbData, "project_menu:")
	app, err := config.Coolify.GetApplicationByUUID(uuid)
	if err != nil {
		_, err = cb.EditMessageText(c, "❌ ꜰᴀɪʟᴇᴅ ᴛᴏ ʟᴏᴀᴅ ᴘʀᴏᴊᴇᴄᴛ: "+err.Error(), nil)
		return err
	}

	text := fmt.Sprintf("<b>📦 %s</b>\n🌐 %s\n📄 ꜱᴛᴀᴛᴜꜱ: <code>%s</code>", app.Name, app.FQDN, app.Status)
	kb := &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{
				{
					Text: "🔄 ʀᴇꜱᴛᴀʀᴛ",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte("restart:" + uuid),
					},
				},
				{
					Text: "🚀 ᴅᴇᴘʟᴏʏ",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte("deploy:" + uuid),
					},
				},
			},
			{
				{
					Text: "📜 ʟᴏɢꜱ",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte("logs:" + uuid),
					},
				},
				{
					Text: "ℹ️ ꜱᴛᴀᴛᴜꜱ",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte("status:" + uuid),
					},
				},
			},
			{
				{
					Text: "📅 ꜱᴄʜᴇᴅᴜʟᴇ",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte("sch_m:" + uuid),
					},
				},
			},
			{
				{
					Text: "🛑 ꜱᴛᴏᴘ",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte("stop:" + uuid),
					},
				},
				{
					Text: "❌ ᴅᴇʟᴇᴛᴇ",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte("delete:" + uuid),
					},
				},
			},
			{
				{
					Text: "🔙 ʙᴀᴄᴋ",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte("list_projects:"),
					},
				},
			},
		},
	}

	_, err = cb.EditMessageText(c, text, &td.EditTextMessageOpts{
		ParseMode:   "HTML",
		ReplyMarkup: kb,
	})

	return err
}

func restartHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	if !config.IsDev(cb.SenderUserId) {
		_ = cb.Answer(c, 0, true, "🚫 ʏᴏᴜ ᴀʀᴇ ɴᴏᴛ ᴀᴜᴛʜᴏʀɪᴢᴇᴅ.", "")
		return nil
	}
	_ = cb.Answer(c, 0, false, "ᴘʀᴏᴄᴇꜱꜱɪɴɢ...", "")

	cbData := cb.DataString()
	uuid := strings.TrimPrefix(cbData, "restart:")

	kb := &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{
				{
					Text: "🔙 ʙᴀᴄᴋ",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte("project_menu:" + uuid),
					},
				},
			},
		},
	}

	res, err := config.Coolify.RestartApplicationByUUID(uuid)
	if err != nil {
		_, _ = cb.EditMessageText(c, "❌ ʀᴇꜱᴛᴀʀᴛ ꜰᴀɪʟᴇᴅ: "+err.Error(), &td.EditTextMessageOpts{ReplyMarkup: kb})
		return nil
	}

	text := "✅ ʀᴇꜱᴛᴀʀᴛ ǫᴜᴇᴜᴇᴅ!"
	if res.DeploymentUUID != "" {
		text += fmt.Sprintf("\nᴅᴇᴘʟᴏʏᴍᴇɴᴛ ᴜᴜɪᴅ: <code>%s</code>", res.DeploymentUUID)
	} else if res.Message != "" {
		text += fmt.Sprintf("\n%s", res.Message)
	}
	_, err = cb.EditMessageText(c, text, &td.EditTextMessageOpts{ParseMode: "HTML", ReplyMarkup: kb})
	return err
}

func deployHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	if !config.IsDev(cb.SenderUserId) {
		_ = cb.Answer(c, 0, true, "🚫 ʏᴏᴜ ᴀʀᴇ ɴᴏᴛ ᴀᴜᴛʜᴏʀɪᴢᴇᴅ.", "")
		return nil
	}

	_ = cb.Answer(c, 0, false, "ᴘʀᴏᴄᴇꜱꜱɪɴɢ...", "")

	cbData := cb.DataString()
	uuid := strings.TrimPrefix(cbData, "deploy:")

	kb := &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{
				{
					Text: "🔙 ʙᴀᴄᴋ",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte("project_menu:" + uuid),
					},
				},
			},
		},
	}

	res, err := config.Coolify.StartApplicationDeployment(uuid, false, false)
	if err != nil {
		_, _ = cb.EditMessageText(c, "❌ ᴅᴇᴘʟᴏʏ ꜰᴀɪʟᴇᴅ: "+err.Error(), &td.EditTextMessageOpts{ReplyMarkup: kb})
		return err
	}

	text := fmt.Sprintf("✅ ᴅᴇᴘʟᴏʏᴍᴇɴᴛ ǫᴜᴇᴜᴇᴅ!\nᴅᴇᴘʟᴏʏᴍᴇɴᴛ ᴜᴜɪᴅ: <code>%s</code>", res.DeploymentUUID)
	_, err = cb.EditMessageText(c, text, &td.EditTextMessageOpts{ParseMode: "HTML", ReplyMarkup: kb})
	return err
}

func logsHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	if !config.IsDev(cb.SenderUserId) {
		_ = cb.Answer(c, 0, true, "🚫 ʏᴏᴜ ᴀʀᴇ ɴᴏᴛ ᴀᴜᴛʜᴏʀɪᴢᴇᴅ.", "")
		return nil
	}
	_ = cb.Answer(c, 0, false, "ᴘʀᴏᴄᴇꜱꜱɪɴɢ...", "")

	uuid := strings.TrimPrefix(cb.DataString(), "logs:")

	kb := &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{
				{
					Text: "🔙 ʙᴀᴄᴋ",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte("project_menu:" + uuid),
					},
				},
			},
		},
	}

	_, _ = cb.EditMessageText(c, "ᴘʀᴏᴄᴇꜱꜱɪɴɢ...", nil)
	logsData, err := config.Coolify.GetApplicationLogsByUUID(uuid)
	if err != nil {
		_, _ = cb.EditMessageText(c, "❌ ʟᴏɢꜱ ᴇʀʀᴏʀ: "+err.Error(), &td.EditTextMessageOpts{ReplyMarkup: kb})
		return nil
	}

	tmpFile, err := os.CreateTemp("", "logs-*.txt")
	if err != nil {
		_, _ = cb.EditMessageText(c, "❌ ꜰᴀɪʟᴇᴅ ᴛᴏ ᴄʀᴇᴀᴛᴇ ᴛᴇᴍᴘ ꜰɪʟᴇ: "+err.Error(), nil)
		return err
	}

	defer os.Remove(tmpFile.Name())
	if _, err := tmpFile.Write([]byte(logsData)); err != nil {
		_, _ = cb.EditMessageText(c, "❌ ꜰᴀɪʟᴇᴅ ᴛᴏ ᴡʀɪᴛᴇ ʟᴏɢꜱ: "+err.Error(), nil)
		return err
	}

	tmpFile.Close()

	file := tmpFile.Name()
	_, err = c.EditMessageMedia(cb.ChatId, &td.InputMessageDocument{Document: td.GetInputFile(file)}, cb.MessageId, &td.EditMessageMediaOpts{ReplyMarkup: kb})
	if err != nil {
		_, _ = cb.EditMessageText(c, "❌ ꜰᴀɪʟᴇᴅ ᴛᴏ ꜱᴇɴᴅ ʟᴏɢꜱ ꜰɪʟᴇ: "+err.Error(), &td.EditTextMessageOpts{ReplyMarkup: kb})
		return fmt.Errorf("edit message media error: %s", err.Error())
	}

	return nil
}

func statusHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	if !config.IsDev(cb.SenderUserId) {
		_ = cb.Answer(c, 0, true, "🚫 ʏᴏᴜ ᴀʀᴇ ɴᴏᴛ ᴀᴜᴛʜᴏʀɪᴢᴇᴅ.", "")
		return nil
	}
	_ = cb.Answer(c, 0, true, "ᴘʀᴏᴄᴇꜱꜱɪɴɢ...", "")

	cbData := cb.DataString()
	uuid := strings.TrimPrefix(cbData, "status:")

	kb := &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{
				{
					Text: "🔙 ʙᴀᴄᴋ",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte("project_menu:" + uuid),
					},
				},
			},
		},
	}

	app, err := config.Coolify.GetApplicationByUUID(uuid)
	if err != nil {
		_, _ = cb.EditMessageText(c, "❌ ꜱᴛᴀᴛᴜꜱ ᴇʀʀᴏʀ: "+err.Error(), &td.EditTextMessageOpts{ReplyMarkup: kb})
		return nil
	}

	text := fmt.Sprintf("📦 <b>%s</b>\nᴄᴜʀʀᴇɴᴛ ꜱᴛᴀᴛᴜꜱ: <code>%s</code>", app.Name, app.Status)
	_, err = cb.EditMessageText(c, text, &td.EditTextMessageOpts{ParseMode: "HTML", ReplyMarkup: kb})
	return err
}

func stopHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	if !config.IsDev(cb.SenderUserId) {
		_ = cb.Answer(c, 0, true, "🚫 ʏᴏᴜ ᴀʀᴇ ɴᴏᴛ ᴀᴜᴛʜᴏʀɪᴢᴇᴅ.", "")
		return nil
	}
	_ = cb.Answer(c, 0, false, "ᴘʀᴏᴄᴇꜱꜱɪɴɢ...", "")

	cbData := cb.DataString()
	uuid := strings.TrimPrefix(cbData, "stop:")

	res, err := config.Coolify.StopApplicationByUUID(uuid)
	kb := &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{
				{
					Text: "🔙 ʙᴀᴄᴋ",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte("project_menu:" + uuid),
					},
				},
			},
		},
	}

	if err != nil {
		_, _ = cb.EditMessageText(c, "❌ ꜱᴛᴏᴘ ꜰᴀɪʟᴇᴅ: "+err.Error(), &td.EditTextMessageOpts{ReplyMarkup: kb})
		return nil
	}

	_, err = cb.EditMessageText(c, "🛑 "+res.Message, &td.EditTextMessageOpts{ReplyMarkup: kb, ParseMode: "HTML"})
	return err
}

func deleteHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	if !config.IsDev(cb.SenderUserId) {
		_ = cb.Answer(c, 0, true, "🚫 ʏᴏᴜ ᴀʀᴇ ɴᴏᴛ ᴀᴜᴛʜᴏʀɪᴢᴇᴅ.", "")
		return nil
	}
	_ = cb.Answer(c, 0, false, "ᴘʀᴏᴄᴇꜱꜱɪɴɢ...", "")

	cbData := cb.DataString()
	uuid := strings.TrimPrefix(cbData, "delete:")

	err := config.Coolify.DeleteApplicationByUUID(uuid)
	kb := &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{
				{
					Text: "🔙 ʙᴀᴄᴋ",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte("project_menu:" + uuid),
					},
				},
			},
		},
	}

	if err != nil {
		_, err = cb.EditMessageText(c, "❌ ᴅᴇʟᴇᴛᴇ ꜰᴀɪʟᴇᴅ: "+err.Error(), &td.EditTextMessageOpts{ReplyMarkup: kb})
		return nil
	}

	_, err = cb.EditMessageText(c, "✅ ᴀᴘᴘʟɪᴄᴀᴛɪᴏɴ ᴅᴇʟᴇᴛᴇᴅ ꜱᴜᴄᴄᴇꜱꜱꜰᴜʟʟʏ.", &td.EditTextMessageOpts{ReplyMarkup: kb})
	return err
}

func scheduleMenuHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	if !config.IsDev(cb.SenderUserId) {
		_ = cb.Answer(c, 0, true, "🚫 ʏᴏᴜ ᴀʀᴇ ɴᴏᴛ ᴀᴜᴛʜᴏʀɪᴢᴇᴅ.", "")
		return nil
	}
	_ = cb.Answer(c, 0, false, "ᴘʀᴏᴄᴇꜱꜱɪɴɢ...", "")

	cbData := cb.DataString()
	uuid := strings.TrimPrefix(cbData, "sch_m:")

	kb := &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{
				{
					Text: "🔄 ʀᴇꜱᴛᴀʀᴛ",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte("sch_a:" + uuid + ":restart"),
					},
				},
			},
			{
				{
					Text: "🔙 ʙᴀᴄᴋ",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte("project_menu:" + uuid),
					},
				},
			},
		},
	}

	_, err := cb.EditMessageText(c, "<b>📅 ꜱᴇʟᴇᴄᴛ ᴀᴄᴛɪᴏɴ ᴛʏᴘᴇ:</b>", &td.EditTextMessageOpts{
		ParseMode:   "HTML",
		ReplyMarkup: kb,
	})
	return err
}

func scheduleActionHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	if !config.IsDev(cb.SenderUserId) {
		_ = cb.Answer(c, 0, true, "🚫 ʏᴏᴜ ᴀʀᴇ ɴᴏᴛ ᴀᴜᴛʜᴏʀɪᴢᴇᴅ.", "")
		return nil
	}
	_ = cb.Answer(c, 0, false, "ᴘʀᴏᴄᴇꜱꜱɪɴɢ...", "")

	// Format: sch_a:uuid:actionType
	cbData := cb.DataString()
	data := strings.TrimPrefix(cbData, "sch_a:")
	parts := strings.Split(data, ":")
	if len(parts) < 2 {
		return nil
	}
	uuid := parts[0]
	actionType := parts[1]

	// Common intervals
	kb := &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{
				{
					Text: "ʜᴏᴜʀʟʏ",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte(fmt.Sprintf("sch_c:%s:%s:every_1h", uuid, actionType)),
					},
				},
			},
			{
				{
					Text: "ᴅᴀɪʟʏ",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte(fmt.Sprintf("sch_c:%s:%s:every_1d", uuid, actionType)),
					},
				},
			},
			{
				{
					Text: "ᴇᴠᴇʀʏ 2 ᴅᴀʏꜱ",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte(fmt.Sprintf("sch_c:%s:%s:every_2d", uuid, actionType)),
					},
				},
			},
			{
				{
					Text: "ᴇᴠᴇʀʏ 3 ᴅᴀʏꜱ",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte(fmt.Sprintf("sch_c:%s:%s:every_3d", uuid, actionType)),
					},
				},
			},
			{
				{
					Text: "ᴡᴇᴇᴋʟʏ",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte(fmt.Sprintf("sch_c:%s:%s:every_7d", uuid, actionType)),
					},
				},
			},
			{
				{
					Text: "🔙 ʙᴀᴄᴋ",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte("sch_m:" + uuid),
					},
				},
			},
		},
	}

	_, err := cb.EditMessageText(c, "<b>⏰ ꜱᴇʟᴇᴄᴛ ꜱᴄʜᴇᴅᴜʟᴇ:</b>", &td.EditTextMessageOpts{
		ParseMode:   "HTML",
		ReplyMarkup: kb,
	})
	return err
}

func scheduleCreateHandler(c *td.Client, cb *td.UpdateNewCallbackQuery) error {
	if !config.IsDev(cb.SenderUserId) {
		_ = cb.Answer(c, 0, true, "🚫 ʏᴏᴜ ᴀʀᴇ ɴᴏᴛ ᴀᴜᴛʜᴏʀɪᴢᴇᴅ.", "")
		return nil
	}
	_ = cb.Answer(c, 0, false, "ᴘʀᴏᴄᴇꜱꜱɪɴɢ...", "")

	// Format: sch_c:uuid:actionType:schedule
	data := strings.TrimPrefix(cb.DataString(), "sch_c:")

	parts := strings.Split(data, ":")
	if len(parts) < 3 {
		return nil
	}
	uuid := parts[0]
	actionType := parts[1]
	schedule := parts[2]

	app, err := config.Coolify.GetApplicationByUUID(uuid)
	if err != nil {
		_, _ = cb.EditMessageText(c, "❌ ꜰᴀɪʟᴇᴅ ᴛᴏ ɢᴇᴛ ᴀᴘᴘʟɪᴄᴀᴛɪᴏɴ: "+err.Error(), nil)
		return nil
	}

	task := database.ScheduledTask{
		ID:          bson.NewObjectID(),
		Name:        app.Name,
		ProjectUUID: uuid,
		Type:        actionType,
		Schedule:    schedule,
	}

	if err := database.AddTask(task); err != nil {
		_, _ = cb.EditMessageText(c, "❌ ꜰᴀɪʟᴇᴅ ᴛᴏ ꜱᴀᴠᴇ ᴛᴀꜱᴋ: "+err.Error(), nil)
		return nil
	}

	if err := scheduler.ScheduleTask(task); err != nil {
		_ = database.DeleteTask(task.ID.Hex())
		_, _ = cb.EditMessageText(c, "❌ ꜰᴀɪʟᴇᴅ ᴛᴏ ꜱᴄʜᴇᴅᴜʟᴇ ᴛᴀꜱᴋ: "+err.Error(), nil)
		return nil
	}

	kb := &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{
				{
					Text: "🔙 ʙᴀᴄᴋ",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte("project_menu:" + uuid),
					},
				},
			},
		},
	}

	_, err = cb.EditMessageText(c, fmt.Sprintf("✅ ᴛᴀꜱᴋ ꜱᴄʜᴇᴅᴜʟᴇᴅ ꜱᴜᴄᴄᴇꜱꜱꜰᴜʟʟʏ!\n\nɪᴅ: <code>%s</code>\nᴛʏᴘᴇ: %s\nꜱᴄʜᴇᴅᴜʟᴇ: %s", task.ID.Hex(), actionType, schedule), &td.EditTextMessageOpts{
		ParseMode:   "HTML",
		ReplyMarkup: kb,
	})
	return err
}
