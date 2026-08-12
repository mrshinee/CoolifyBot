package src

import (
	"fmt"
	"runtime"
	"time"

	td "github.com/AshokShau/gotdbot"
)

func startHandler(c *td.Client, msg *td.Message) error {
	sender, err := msg.GetUser(c)
	if err != nil {
		return fmt.Errorf("failed to get sender: %w", err)
	}

	firstName := "there"
	if sender != nil {
		firstName = sender.FirstName
	}

	response := fmt.Sprintf(`🌟 <b>ʜᴇʏ, %s!</b>

ʏᴏᴜ'ᴠᴇ ʀᴇᴀᴄʜᴇᴅ <b>%s</b> — ʏᴏᴜʀ ꜱᴍᴀʀᴛ ᴄᴏɴᴛʀᴏʟ ᴘᴀɴᴇʟ ꜰᴏʀ ᴍᴀɴᴀɢɪɴɢ <b>ᴄᴏᴏʟɪꜰʏ</b> ᴘʀᴏᴊᴇᴄᴛꜱ ᴅɪʀᴇᴄᴛʟʏ ꜰʀᴏᴍ ᴛᴇʟᴇɢʀᴀᴍ.

<blockquote>⚡ ᴅᴇᴘʟᴏʏ · 🔄 ʀᴇꜱᴛᴀʀᴛ · 📋 ᴍᴀɴᴀɢᴇ — ᴀʟʟ ɪɴ ᴏɴᴇ ᴘʟᴀᴄᴇ.</blockquote>

ᴛᴀᴘ ᴛʜᴇ ʙᴜᴛᴛᴏɴ ʙᴇʟᴏᴡ ᴛᴏ ɢᴇᴛ ꜱᴛᴀʀᴛᴇᴅ 👇`, firstName, c.Me.FirstName)

	kb := &td.ReplyMarkupInlineKeyboard{
		Rows: [][]td.InlineKeyboardButton{
			{
				{
					Text: "📋 ʟɪꜱᴛ ᴘʀᴏᴊᴇᴄᴛꜱ",
					Type: &td.InlineKeyboardButtonTypeCallback{
						Data: []byte("list_projects"),
					},
				},
			},
			{
				{
					Text: "💫 Uᴘᴅᴀᴛᴇs",
					Type: &td.InlineKeyboardButtonTypeUrl{
						Url: "https://t.me/DynamicXNetwork",
					},
				},
				{
					Text: "🛠 Mᴀɪɴᴛᴀɪɴᴇʀ",
					Type: &td.InlineKeyboardButtonTypeUrl{
						Url: "https://t.me/NullXShadow",
					},
				},
			},
		},
	}

	_, err = msg.ReplyText(c, response, &td.SendTextMessageOpts{ParseMode: "HTML", ReplyMarkup: kb})
	if err != nil {
		return fmt.Errorf("failed to send start message: %w", err)
	}
	return nil
}

func pingHandler(c *td.Client, msg *td.Message) error {
	start := time.Now()
	sentMsg, err := msg.ReplyText(c, "⏱️ ᴘɪɴɢɪɴɢ...", nil)
	if err != nil {
		return fmt.Errorf("failed to send ping message: %w", err)
	}

	latency := time.Since(start).Milliseconds()
	uptime := time.Since(startTime).Truncate(time.Second)

	response := fmt.Sprintf(
		"<b>📊 ꜱʏꜱᴛᴇᴍ ᴘᴇʀꜰᴏʀᴍᴀɴᴄᴇ ᴍᴇᴛʀɪᴄꜱ</b>\n\n"+
			"⏱️ <b>ʙᴏᴛ ʟᴀᴛᴇɴᴄʏ:</b> <code>%d ms</code>\n"+
			"🕒 <b>ᴜᴘᴛɪᴍᴇ:</b> <code>%s</code>\n"+
			"📩 <b>ᴜᴘᴅᴀᴛᴇ ʟᴀɢ:</b> <code>%d ms</code>\n"+
			"⚙️ <b>ɢᴏ ʀᴏᴜᴛɪɴᴇꜱ:</b> <code>%d</code>\n",
		latency, uptime, runtime.NumGoroutine(),
	)

	_, err = sentMsg.EditText(c, response, &td.EditTextMessageOpts{ParseMode: "HTML"})
	if err != nil {
		return fmt.Errorf("failed to edit ping message: %w", err)
	}
	return nil
}
