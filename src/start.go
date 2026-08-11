package src

import (
	"fmt"
	"runtime"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
)

func startHandler(m *telegram.NewMessage) error {
	bot := m.Client.Me()
	response := fmt.Sprintf(`
🌟 <b>ʜᴇʏ, %s!</b>

ʏᴏᴜ'ᴠᴇ ʀᴇᴀᴄʜᴇᴅ <b>%s</b> — ʏᴏᴜʀ ꜱᴍᴀʀᴛ ᴄᴏɴᴛʀᴏʟ ᴘᴀɴᴇʟ ꜰᴏʀ ᴍᴀɴᴀɢɪɴɢ <b>ᴄᴏᴏʟɪꜰʏ</b> ᴘʀᴏᴊᴇᴄᴛꜱ ᴅɪʀᴇᴄᴛʟʏ ꜰʀᴏᴍ ᴛᴇʟᴇɢʀᴀᴍ.

<blockquote>⚡ ᴅᴇᴘʟᴏʏ · 🔄 ʀᴇꜱᴛᴀʀᴛ · 📋 ᴍᴀɴᴀɢᴇ — ᴀʟʟ ɪɴ ᴏɴᴇ ᴘʟᴀᴄᴇ.</blockquote>

ᴛᴀᴘ ᴛʜᴇ ʙᴜᴛᴛᴏɴ ʙᴇʟᴏᴡ ᴛᴏ ɢᴇᴛ ꜱᴛᴀʀᴛᴇᴅ 👇`, m.Sender.FirstName, bot.FirstName)

	keyboard := telegram.NewKeyboard().
		AddRow(
			telegram.Button.Data("📋 ʟɪꜱᴛ ᴘʀᴏᴊᴇᴄᴛꜱ", "list_projects"),
		).
		AddRow(
			telegram.Button.URL("💫 Uᴘᴅᴀᴛᴇs", "https://t.me/DynamicXNetwork"),
			telegram.Button.URL("🛠️ Mᴀɪɴᴛᴀɪɴᴇʀ", "https://t.me/NullXShadow"),
		)

	_, err := m.Reply(response, &telegram.SendOptions{
		ReplyMarkup: keyboard.Build(),
	})
	return err
}

func pingHandler(m *telegram.NewMessage) error {
	start := time.Now()
	updateLag := time.Since(time.Unix(int64(m.Date()), 0)).Milliseconds()

	msg, err := m.Reply("⏱️ ᴘɪɴɢɪɴɢ...")
	if err != nil {
		return err
	}

	latency := time.Since(start).Milliseconds()
	uptime := time.Since(startTime).Truncate(time.Second)

	response := fmt.Sprintf(
		"<b>📊 ꜱʏꜱᴛᴇᴍ ᴘᴇʀꜰᴏʀᴍᴀɴᴄᴇ ᴍᴇᴛʀɪᴄꜱ</b>\n\n"+
			"⏱️ <b>ʙᴏᴛ ʟᴀᴛᴇɴᴄʏ:</b> <code>%d ms</code>\n"+
			"🕒 <b>ᴜᴘᴛɪᴍᴇ:</b> <code>%s</code>\n"+
			"📩 <b>ᴜᴘᴅᴀᴛᴇ ʟᴀɢ:</b> <code>%d ms</code>\n"+
			"⚙️ <b>ɢᴏ ʀᴏᴜᴛɪɴᴇꜱ:</b> <code>%d</code>\n",
		latency, uptime, updateLag, runtime.NumGoroutine(),
	)

	_, err = msg.Edit(response)
	return err
}
