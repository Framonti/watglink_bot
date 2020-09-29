package main

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"strconv"

	"github.com/Rhymen/go-whatsapp"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func whatsappToTelegram(cont context) {
	telegramIDi, err := strconv.Atoi(cont.userID)
	if err != nil {
		return
	}
	pmsg, err := json.Marshal(cont.MessageProto)
	if err != nil {
		return
	}
	protomessage := base64.StdEncoding.EncodeToString(pmsg)
	var lines string
	if cont.GroupName != "" {
		lines = "[💬](https://a.aa/" + cont.RemoteJid + "/" + cont.MessageID + "/" + protomessage + ") *Group:* " + cont.GroupName + "\n👤 *Sender:* " + cont.SenderName
	} else {
		lines = "[👤](https://a.aa/" + cont.RemoteJid + "/" + cont.MessageID + "/" + protomessage + ") *Sender:* " + cont.SenderName + ""
	}
	if cont.QuotedMessageText != "" {
		lines = lines + "\n\n↩️ *Reply to* " + cont.QuotedMessageSender + "\n\"" + cont.QuotedMessageText + "\""
	}
	switch cont.MediaType {
	case "image":
		pic := tgbotapi.NewPhotoUpload(int64(telegramIDi), "./img"+cont.userID+".png")
		pic.Caption = lines + "\n\n" + cont.MediaCaption
		pic.ParseMode = "Markdown"
		bot.Send(pic)
		os.Remove("./img" + cont.userID + ".png")
	case "video":
		vid := tgbotapi.NewVideoUpload(int64(telegramIDi), "./vid"+cont.userID+"-"+cont.FileName+".mp4")
		vid.Caption = lines + "\n\n" + cont.MediaCaption
		vid.ParseMode = "Markdown"
		bot.Send(vid)
		os.Remove("./vid" + cont.userID + "-" + cont.FileName + ".mp4")
	case "audio":
		aud := tgbotapi.NewAudioUpload(int64(telegramIDi), "./aud"+cont.userID+".mp3")
		aud.Caption = lines
		aud.ParseMode = "Markdown"
		bot.Send(aud)
		os.Remove("./aud" + cont.userID + ".mp3")
	case "document":
		doc := tgbotapi.NewDocumentUpload(int64(telegramIDi), "./doc"+cont.userID+"-"+cont.FileName)
		doc.Caption = lines
		doc.ParseMode = "Markdown"
		bot.Send(doc)
		os.Remove("./doc" + cont.userID + "-" + cont.FileName)
	default:
		msg := tgbotapi.NewMessage(int64(telegramIDi), lines+"\n\n"+cont.MessageText)
		msg.ParseMode = "Markdown"
		bot.Send(msg)
	}
}

func (i informations) sendAlertToTelegram(message string) {
	telegramIDi, err := strconv.Atoi(i.UserID)
	if err != nil {
		return
	}
	msg := tgbotapi.NewMessage(int64(telegramIDi), message)
	msg.ParseMode = "HTML"
	msg.DisableWebPagePreview = true
	bot.Send(msg)
}

func (i informations) sendQr() {
	telegramIDi, err := strconv.Atoi(i.UserID)
	if err != nil {
		return
	}
	pic := tgbotapi.NewPhotoUpload(int64(telegramIDi), "./qr"+i.UserID+".png")
	pic.Caption = "📷 <b>Scan this QR code</b> on your WhatsApp app to login.\n<a href='https://faq.whatsapp.com/general/download-and-installation/how-to-log-in-or-out'>Tutorial</a>"
	pic.ParseMode = "HTML"
	bot.Send(pic)
	os.Remove("./qr" + i.UserID + ".png")
}

func telegramToWhatsapp(message, toSendJid, messageID, messageProtos string, telegramID int) {

	wac := connections[strconv.Itoa(telegramID)]

	if wac == nil {
		msg := tgbotapi.NewMessage(int64(telegramID), "❌ <b>Error sending message to WhatsApp.</b>\nYou're disconnected! Start the session again to send messages.")
		msg.ParseMode = "HTML"
		bot.Send(msg)
		return
	}

	_, err := wac.Send(whatsapp.TextMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: toSendJid,
		},
		Text: message,
	})

	if err != nil {
		msg := tgbotapi.NewMessage(int64(telegramID), "❌ <b>Error sending message to WhatsApp.</b>\nFull traceback: "+err.Error())
		msg.ParseMode = "HTML"
		bot.Send(msg)
	}

}
