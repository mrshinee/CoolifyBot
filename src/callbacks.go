package src

import (
	"coolifymanager/src/config"
	"coolifymanager/src/database"
	"coolifymanager/src/scheduler"
	"fmt"
	"os"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func listProjectsHandler(cb *telegram.CallbackQuery) error {
	if !config.IsDev(cb.SenderID) {
		_, _ = cb.Answer("🚫 ʏᴏᴜ ᴀʀᴇ ɴᴏᴛ ᴀᴜᴛʜᴏʀɪᴢᴇᴅ.", &telegram.CallbackOptions{Alert: true})
		return nil
	}
	_, _ = cb.Answer("ᴘʀᴏᴄᴇꜱꜱɪɴɢ...")
	apps, err := config.Coolify.ListApplications()
	if err != nil {
		_, _ = cb.Edit("❌ ꜰᴀɪʟᴇᴅ ᴛᴏ ꜰᴇᴛᴄʜ ᴘʀᴏᴊᴇᴄᴛꜱ:" + err.Error())
		return nil
	}

	if len(apps) == 0 {
		_, _ = cb.Edit("😶 ɴᴏ ᴀᴘᴘʟɪᴄᴀᴛɪᴏɴꜱ ꜰᴏᴜɴᴅ.")
		return nil
	}

	page := 1
	data := cb.DataString()
	if strings.Contains(data, ":") {
		parts := strings.Split(data, ":")
		if len(parts) > 1 {
			fmt.Sscanf(parts[1], "%d", &page)
		}
	}
	if page < 1 {
		page = 1
	}

	start, end, paginationButtons := Paginate(len(apps), page, 7, "list_projects:")
	kb := telegram.NewKeyboard()
	for _, app := range apps[start:end] {
		text := fmt.Sprintf("📦 %s (%s)", app.Name, app.Status)
		data := "project_menu:" + app.UUID
		kb.AddRow(telegram.Button.Data(text, data))
	}

	if len(paginationButtons) > 0 {
		var row []telegram.KeyboardButton
		for _, btn := range paginationButtons {
			row = append(row, telegram.Button.Data(btn.Text, btn.Data))
		}
		kb.AddRow(row...)
	}

	kb.AddRow(telegram.Button.Data("🔄 ʀᴇꜰʀᴇꜱʜ", "refresh_projects:"))

	_, err = cb.Edit("<b>📋 ꜱᴇʟᴇᴄᴛ ᴀ ᴘʀᴏᴊᴇᴄᴛ:</b>", &telegram.SendOptions{ReplyMarkup: kb.Build()})
	return err
}

// refreshProjectsHandler clears the cached applications list and re-renders
// the project menu, so a deleted/recreated app doesn't keep showing up
// with a stale UUID for the rest of the 30-minute cache window.
func refreshProjectsHandler(cb *telegram.CallbackQuery) error {
	if !config.IsDev(cb.SenderID) {
		_, _ = cb.Answer("🚫 ʏᴏᴜ ᴀʀᴇ ɴᴏᴛ ᴀᴜᴛʜᴏʀɪᴢᴇᴅ.", &telegram.CallbackOptions{Alert: true})
		return nil
	}
	config.Coolify.InvalidateApplicationsCache()
	_, _ = cb.Answer("🔄 ʀᴇꜰʀᴇꜱʜᴇᴅ")
	return listProjectsHandler(cb)
}

func projectMenuHandler(cb *telegram.CallbackQuery) error {
	if !config.IsDev(cb.SenderID) {
		_, _ = cb.Answer("🚫 ʏᴏᴜ ᴀʀᴇ ɴᴏᴛ ᴀᴜᴛʜᴏʀɪᴢᴇᴅ.", &telegram.CallbackOptions{Alert: true})
		return nil
	}

	_, _ = cb.Answer("ᴘʀᴏᴄᴇꜱꜱɪɴɢ...")
	uuid := strings.TrimPrefix(cb.DataString(), "project_menu:")

	app, err := config.Coolify.GetApplicationByUUID(uuid)
	if err != nil {
		_, err = cb.Edit("❌ ꜰᴀɪʟᴇᴅ ᴛᴏ ʟᴏᴀᴅ ᴘʀᴏᴊᴇᴄᴛ: " + err.Error())
		return err
	}

	text := fmt.Sprintf("<b>📦 %s</b>\n🌐 %s\n📄 ꜱᴛᴀᴛᴜꜱ: <code>%s</code>", app.Name, app.FQDN, app.Status)
	keyboard := telegram.NewKeyboard().
		AddRow(telegram.Button.Data("🔄 ʀᴇꜱᴛᴀʀᴛ", "restart:"+uuid), telegram.Button.Data("🚀 ᴅᴇᴘʟᴏʏ", "deploy:"+uuid)).
		AddRow(telegram.Button.Data("📜 ʟᴏɢꜱ", "logs:"+uuid), telegram.Button.Data("ℹ️ ꜱᴛᴀᴛᴜꜱ", "status:"+uuid)).
		AddRow(telegram.Button.Data("📅 ꜱᴄʜᴇᴅᴜʟᴇ", "sch_m:"+uuid)).
		AddRow(telegram.Button.Data("🛑 ꜱᴛᴏᴘ", "stop:"+uuid), telegram.Button.Data("❌ ᴅᴇʟᴇᴛᴇ", "delete:"+uuid)).
		AddRow(telegram.Button.Data("🔙 ʙᴀᴄᴋ", "list_projects:"))

	_, err = cb.Edit(text, &telegram.SendOptions{
		ParseMode:   "HTML",
		ReplyMarkup: keyboard.Build(),
	})
	return err
}

func restartHandler(cb *telegram.CallbackQuery) error {
	if !config.IsDev(cb.SenderID) {
		_, _ = cb.Answer("🚫 ʏᴏᴜ ᴀʀᴇ ɴᴏᴛ ᴀᴜᴛʜᴏʀɪᴢᴇᴅ.", &telegram.CallbackOptions{Alert: true})
		return nil
	}
	_, _ = cb.Answer("ᴘʀᴏᴄᴇꜱꜱɪɴɢ...")
	uuid := strings.TrimPrefix(cb.DataString(), "restart:")

	keyboard := telegram.NewKeyboard().
		AddRow(telegram.Button.Data("🔙 ʙᴀᴄᴋ", "project_menu:"+uuid))

	res, err := config.Coolify.RestartApplicationByUUID(uuid)
	if err != nil {
		_, _ = cb.Edit("❌ ʀᴇꜱᴛᴀʀᴛ ꜰᴀɪʟᴇᴅ: "+err.Error(), &telegram.SendOptions{ReplyMarkup: keyboard.Build()})
		return nil
	}

	text := fmt.Sprintf("✅ ʀᴇꜱᴛᴀʀᴛ ǫᴜᴇᴜᴇᴅ!\nᴅᴇᴘʟᴏʏᴍᴇɴᴛ ᴜᴜɪᴅ: <code>%s</code>", res.DeploymentUUID)
	_, err = cb.Edit(text, &telegram.SendOptions{ParseMode: "HTML", ReplyMarkup: keyboard.Build()})
	return err
}

func deployHandler(cb *telegram.CallbackQuery) error {
	if !config.IsDev(cb.SenderID) {
		_, _ = cb.Answer("🚫 ʏᴏᴜ ᴀʀᴇ ɴᴏᴛ ᴀᴜᴛʜᴏʀɪᴢᴇᴅ.", &telegram.CallbackOptions{Alert: true})
		return nil
	}
	_, _ = cb.Answer("ᴘʀᴏᴄᴇꜱꜱɪɴɢ...")
	uuid := strings.TrimPrefix(cb.DataString(), "deploy:")
	keyboard := telegram.NewKeyboard().
		AddRow(telegram.Button.Data("🔙 ʙᴀᴄᴋ", "project_menu:"+uuid))
	res, err := config.Coolify.StartApplicationDeployment(uuid, false, false)
	if err != nil {
		_, _ = cb.Edit("❌ ᴅᴇᴘʟᴏʏ ꜰᴀɪʟᴇᴅ: "+err.Error(), &telegram.SendOptions{ReplyMarkup: keyboard.Build()})
		return err
	}
	text := fmt.Sprintf("✅ ᴅᴇᴘʟᴏʏᴍᴇɴᴛ ǫᴜᴇᴜᴇᴅ!\nᴅᴇᴘʟᴏʏᴍᴇɴᴛ ᴜᴜɪᴅ: <code>%s</code>", res.DeploymentUUID)
	_, err = cb.Edit(text, &telegram.SendOptions{ParseMode: "HTML", ReplyMarkup: keyboard.Build()})
	return err
}

func logsHandler(cb *telegram.CallbackQuery) error {
	if !config.IsDev(cb.SenderID) {
		_, _ = cb.Answer("🚫 ʏᴏᴜ ᴀʀᴇ ɴᴏᴛ ᴀᴜᴛʜᴏʀɪᴢᴇᴅ.", &telegram.CallbackOptions{Alert: true})
		return nil
	}
	_, _ = cb.Answer("ᴘʀᴏᴄᴇꜱꜱɪɴɢ...")
	uuid := strings.TrimPrefix(cb.DataString(), "logs:")
	keyboard := telegram.NewKeyboard().
		AddRow(telegram.Button.Data("🔙 ʙᴀᴄᴋ", "project_menu:"+uuid))

	msg, _ := cb.Edit("ᴘʀᴏᴄᴇꜱꜱɪɴɢ...")
	logsData, err := config.Coolify.GetApplicationLogsByUUID(uuid)
	if err != nil {
		_, _ = cb.Edit("❌ ʟᴏɢꜱ ᴇʀʀᴏʀ: "+err.Error(), &telegram.SendOptions{
			ReplyMarkup: keyboard.Build(),
		})
		return nil
	}

	tmpFile, err := os.CreateTemp("", "logs-*.txt")
	if err != nil {
		_, _ = cb.Edit("❌ ꜰᴀɪʟᴇᴅ ᴛᴏ ᴄʀᴇᴀᴛᴇ ᴛᴇᴍᴘ ꜰɪʟᴇ: " + err.Error())
		return err
	}

	defer os.Remove(tmpFile.Name())
	if _, err := tmpFile.Write([]byte(logsData)); err != nil {
		_, _ = cb.Edit("❌ ꜰᴀɪʟᴇᴅ ᴛᴏ ᴡʀɪᴛᴇ ʟᴏɢꜱ: " + err.Error())
		return err
	}
	tmpFile.Close()

	opts := telegram.SendOptions{
		Upload: &telegram.UploadOptions{
			ProgressCallback: func(pi *telegram.ProgressInfo) {
				msg.Edit(fmt.Sprintf("Uploading... %.2f%% complete (%.2f MB/s), ETA: %.2f seconds",
					pi.Percentage,
					pi.CurrentSpeed/1024/1024,
					pi.ETA,
				))
			},
			ProgressInterval: 5,
		},
		Media: tmpFile.Name(),
		Attributes: []telegram.DocumentAttribute{
			&telegram.DocumentAttributeFilename{
				FileName: tmpFile.Name(),
			},
		},
		Caption:     "LOGS",
		ReplyMarkup: keyboard.Build(),
	}
	_, err = msg.Edit("LOGS", &opts)
	if err != nil {
		_, _ = cb.Edit("❌ ꜰᴀɪʟᴇᴅ ᴛᴏ ꜱᴇɴᴅ ʟᴏɢꜱ: "+err.Error(), &telegram.SendOptions{ReplyMarkup: keyboard.Build()})
		return err
	}

	return nil
}

func statusHandler(cb *telegram.CallbackQuery) error {
	if !config.IsDev(cb.SenderID) {
		_, _ = cb.Answer("🚫 ʏᴏᴜ ᴀʀᴇ ɴᴏᴛ ᴀᴜᴛʜᴏʀɪᴢᴇᴅ.", &telegram.CallbackOptions{Alert: true})
		return nil
	}
	_, _ = cb.Answer("ᴘʀᴏᴄᴇꜱꜱɪɴɢ...")
	uuid := strings.TrimPrefix(cb.DataString(), "status:")
	keyboard := telegram.NewKeyboard().
		AddRow(telegram.Button.Data("🔙 ʙᴀᴄᴋ", "project_menu:"+uuid))
	app, err := config.Coolify.GetApplicationByUUID(uuid)
	if err != nil {
		_, _ = cb.Edit("❌ ꜱᴛᴀᴛᴜꜱ ᴇʀʀᴏʀ: "+err.Error(), &telegram.SendOptions{ReplyMarkup: keyboard.Build()})
		return nil
	}

	text := fmt.Sprintf("📦 <b>%s</b>\nᴄᴜʀʀᴇɴᴛ ꜱᴛᴀᴛᴜꜱ: <code>%s</code>", app.Name, app.Status)
	_, err = cb.Edit(text, &telegram.SendOptions{ParseMode: "HTML", ReplyMarkup: keyboard.Build()})
	return err
}

func stopHandler(cb *telegram.CallbackQuery) error {
	if !config.IsDev(cb.SenderID) {
		_, _ = cb.Answer("🚫 ʏᴏᴜ ᴀʀᴇ ɴᴏᴛ ᴀᴜᴛʜᴏʀɪᴢᴇᴅ.", &telegram.CallbackOptions{Alert: true})
		return nil
	}
	_, _ = cb.Answer("ᴘʀᴏᴄᴇꜱꜱɪɴɢ...")
	uuid := strings.TrimPrefix(cb.DataString(), "stop:")
	res, err := config.Coolify.StopApplicationByUUID(uuid)
	keyboard := telegram.NewKeyboard().
		AddRow(telegram.Button.Data("🔙 ʙᴀᴄᴋ", "project_menu:"+uuid))
	if err != nil {
		_, _ = cb.Edit("❌ ꜱᴛᴏᴘ ꜰᴀɪʟᴇᴅ: "+err.Error(), &telegram.SendOptions{ReplyMarkup: keyboard.Build()})
		return nil
	}

	_, err = cb.Edit("🛑 "+res.Message, &telegram.SendOptions{ReplyMarkup: keyboard.Build()})
	return err
}

func deleteHandler(cb *telegram.CallbackQuery) error {
	if !config.IsDev(cb.SenderID) {
		_, _ = cb.Answer("🚫 ʏᴏᴜ ᴀʀᴇ ɴᴏᴛ ᴀᴜᴛʜᴏʀɪᴢᴇᴅ.", &telegram.CallbackOptions{Alert: true})
		return nil
	}

	_, _ = cb.Answer("ᴘʀᴏᴄᴇꜱꜱɪɴɢ...")
	uuid := strings.TrimPrefix(cb.DataString(), "delete:")
	err := config.Coolify.DeleteApplicationByUUID(uuid)
	keyboard := telegram.NewKeyboard().
		AddRow(telegram.Button.Data("🔙 ʙᴀᴄᴋ", "project_menu:"+uuid))
	if err != nil {
		_, err = cb.Edit("❌ ᴅᴇʟᴇᴛᴇ ꜰᴀɪʟᴇᴅ: "+err.Error(), &telegram.SendOptions{ReplyMarkup: keyboard.Build()})
		return nil
	}

	_, err = cb.Edit("✅ ᴀᴘᴘʟɪᴄᴀᴛɪᴏɴ ᴅᴇʟᴇᴛᴇᴅ ꜱᴜᴄᴄᴇꜱꜱꜰᴜʟʟʏ.", &telegram.SendOptions{ReplyMarkup: keyboard.Build()})
	return err
}

func scheduleMenuHandler(cb *telegram.CallbackQuery) error {
	if !config.IsDev(cb.SenderID) {
		_, _ = cb.Answer("🚫 ʏᴏᴜ ᴀʀᴇ ɴᴏᴛ ᴀᴜᴛʜᴏʀɪᴢᴇᴅ.", &telegram.CallbackOptions{Alert: true})
		return nil
	}
	_, _ = cb.Answer("ᴘʀᴏᴄᴇꜱꜱɪɴɢ...")
	uuid := strings.TrimPrefix(cb.DataString(), "sch_m:")

	keyboard := telegram.NewKeyboard().
		AddRow(telegram.Button.Data("🔄 ʀᴇꜱᴛᴀʀᴛ", "sch_a:"+uuid+":restart")).
		AddRow(telegram.Button.Data("🔙 ʙᴀᴄᴋ", "project_menu:"+uuid))

	_, err := cb.Edit("<b>📅 ꜱᴇʟᴇᴄᴛ ᴀᴄᴛɪᴏɴ ᴛʏᴘᴇ:</b>", &telegram.SendOptions{
		ParseMode:   "HTML",
		ReplyMarkup: keyboard.Build(),
	})
	return err
}

func scheduleActionHandler(cb *telegram.CallbackQuery) error {
	if !config.IsDev(cb.SenderID) {
		_, _ = cb.Answer("🚫 ʏᴏᴜ ᴀʀᴇ ɴᴏᴛ ᴀᴜᴛʜᴏʀɪᴢᴇᴅ.", &telegram.CallbackOptions{Alert: true})
		return nil
	}
	_, _ = cb.Answer("ᴘʀᴏᴄᴇꜱꜱɪɴɢ...")
	// Format: sch_a:uuid:actionType
	data := strings.TrimPrefix(cb.DataString(), "sch_a:")
	parts := strings.Split(data, ":")
	if len(parts) < 2 {
		return nil
	}
	uuid := parts[0]
	actionType := parts[1]

	// Common intervals
	keyboard := telegram.NewKeyboard().
		AddRow(telegram.Button.Data("ʜᴏᴜʀʟʏ", fmt.Sprintf("sch_c:%s:%s:every_1h", uuid, actionType))).
		AddRow(telegram.Button.Data("ᴅᴀɪʟʏ", fmt.Sprintf("sch_c:%s:%s:every_1d", uuid, actionType))).
		AddRow(telegram.Button.Data("ᴇᴠᴇʀʏ 2 ᴅᴀʏꜱ", fmt.Sprintf("sch_c:%s:%s:every_2d", uuid, actionType))).
		AddRow(telegram.Button.Data("ᴇᴠᴇʀʏ 3 ᴅᴀʏꜱ", fmt.Sprintf("sch_c:%s:%s:every_3d", uuid, actionType))).
		AddRow(telegram.Button.Data("ᴡᴇᴇᴋʟʏ", fmt.Sprintf("sch_c:%s:%s:every_7d", uuid, actionType))).
		AddRow(telegram.Button.Data("🔙 ʙᴀᴄᴋ", "sch_m:"+uuid))

	_, err := cb.Edit("<b>⏰ ꜱᴇʟᴇᴄᴛ ꜱᴄʜᴇᴅᴜʟᴇ:</b>", &telegram.SendOptions{
		ParseMode:   "HTML",
		ReplyMarkup: keyboard.Build(),
	})
	return err
}

func scheduleCreateHandler(cb *telegram.CallbackQuery) error {
	if !config.IsDev(cb.SenderID) {
		_, _ = cb.Answer("🚫 ʏᴏᴜ ᴀʀᴇ ɴᴏᴛ ᴀᴜᴛʜᴏʀɪᴢᴇᴅ.", &telegram.CallbackOptions{Alert: true})
		return nil
	}
	_, _ = cb.Answer("ꜱᴄʜᴇᴅᴜʟɪɴɢ...")
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
		_, _ = cb.Edit("❌ ꜰᴀɪʟᴇᴅ ᴛᴏ ɢᴇᴛ ᴀᴘᴘʟɪᴄᴀᴛɪᴏɴ: " + err.Error())
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
		_, _ = cb.Edit("❌ ꜰᴀɪʟᴇᴅ ᴛᴏ ꜱᴀᴠᴇ ᴛᴀꜱᴋ: " + err.Error())
		return nil
	}

	if err := scheduler.ScheduleTask(task); err != nil {
		_ = database.DeleteTask(task.ID.Hex())
		_, _ = cb.Edit("❌ ꜰᴀɪʟᴇᴅ ᴛᴏ ꜱᴄʜᴇᴅᴜʟᴇ ᴛᴀꜱᴋ: " + err.Error())
		return nil
	}

	keyboard := telegram.NewKeyboard().
		AddRow(telegram.Button.Data("🔙 ʙᴀᴄᴋ", "project_menu:"+uuid))

	_, err = cb.Edit(fmt.Sprintf("✅ ᴛᴀꜱᴋ ꜱᴄʜᴇᴅᴜʟᴇᴅ ꜱᴜᴄᴄᴇꜱꜱꜰᴜʟʟʏ!\n\nɪᴅ: <code>%s</code>\nᴛʏᴘᴇ: %s\nꜱᴄʜᴇᴅᴜʟᴇ: %s", task.ID.Hex(), actionType, schedule), &telegram.SendOptions{
		ParseMode:   "HTML",
		ReplyMarkup: keyboard.Build(),
	})
	return err
}
