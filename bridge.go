package main

import(
	"encoding/base64"
	"encoding/json"
	"strconv"
	"time"
	"os"

	"github.com/Rhymen/go-whatsapp"
	"github.com/siddontang/go-mysql/client"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
)


var bot, _ = tgbotapi.NewBotAPI("bot token")

func WhatsappToTelegram(cont Context) {
	telegramIDi, err := strconv.Atoi(cont.UserID)
	if err != nil {
		return
	}
	pmsg, err := json.Marshal(cont.MessageProto)
	if err != nil {
		return
	}
	protomessage := base64.StdEncoding.EncodeToString(pmsg)
	print(string(protomessage)+"\n")
	var lines string
	if cont.GroupName != "" {
		lines = "[💬](https://a.aa/"+cont.RemoteJid+"/"+cont.MessageID+"/"+protomessage+") *Group:* "+cont.GroupName+"\n👤 *Sender:* "+cont.SenderName
	}else{
		lines = "[👤](https://a.aa/"+cont.RemoteJid+"/"+cont.MessageID+"/"+protomessage+") *Sender:* "+cont.SenderName+""
	}
	if cont.QuotedMessageText != "" {
		lines = lines+"\n\n↩️ *Reply to* "+cont.QuotedMessageSender+"\n\""+cont.QuotedMessageText+"\""
	}
	switch cont.MediaType {
		case "image":
			pic := tgbotapi.NewPhotoUpload(int64(telegramIDi), "./img"+cont.UserID+".png")
			pic.Caption = lines+"\n\n"+cont.MediaCaption
			pic.ParseMode = "Markdown"
			bot.Send(pic)
			os.Remove("./img"+cont.UserID+".png")
		case "video":
			vid := tgbotapi.NewVideoUpload(int64(telegramIDi), "./vid"+cont.UserID+"-"+cont.FileName+".mp4")
			vid.Caption = lines+"\n\n"+cont.MediaCaption
			vid.ParseMode = "Markdown"
			bot.Send(vid)
			os.Remove("./vid"+cont.UserID+"-"+cont.FileName+".mp4")
		case "audio":
			aud := tgbotapi.NewAudioUpload(int64(telegramIDi), "./aud"+cont.UserID+".mp3")
			aud.Caption = lines
			aud.ParseMode = "Markdown"
			bot.Send(aud)
			os.Remove("./aud"+cont.UserID+".mp3")
		case "document":
			doc := tgbotapi.NewDocumentUpload(int64(telegramIDi), "./doc"+cont.UserID+"-"+cont.FileName)
			doc.Caption = lines
			doc.ParseMode = "Markdown"
			bot.Send(doc)
			os.Remove("./doc"+cont.UserID+"-"+cont.FileName)
		default:
			msg := tgbotapi.NewMessage(int64(telegramIDi), lines+"\n\n"+cont.MessageText)
			msg.ParseMode = "Markdown"
			bot.Send(msg)
	}
}

func (i Informations) SendAlertToTelegram(message string) {
	telegramIDi, err := strconv.Atoi(i.UserID)
	if err != nil {
		return
	}
	msg := tgbotapi.NewMessage(int64(telegramIDi), message)
	msg.ParseMode = "HTML"
	bot.Send(msg)
}

func (i Informations) SendQr() {
	telegramIDi, err := strconv.Atoi(i.UserID)
	if err != nil {
		return
	}
	pic := tgbotapi.NewPhotoUpload(int64(telegramIDi), "./qr"+i.UserID+".png")
	pic.Caption = "📷 <b>Scan this QR code</b> on your WhatsApp app to login.\n<a href='https://web.whatsapp.com/whatsapp-webclient-login_a0f99e8cbba9eaa747ec23ffb30d63fe.mp4'>Tutorial</a>"
	pic.ParseMode = "HTML"
	bot.Send(pic)
	os.Remove("./qr"+i.UserID+".png")
}


func TelegramToWhatsapp(message, toSendJid, messageID, messageProtos string, db *client.Conn, telegramID int) {
	
	wac := connections[strconv.Itoa(telegramID)]	
	i := Informations{strconv.Itoa(telegramID), time.Now().Unix(), db}
	
	if wac == nil {
		return
	}
	
	_, err := wac.Send(whatsapp.TextMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: toSendJid,
		},
		Text: message,
	})
	
	if err != nil {
		i.SendAlertToTelegram("❌ <b>Error sending message to WhatsApp.</b>\nFull traceback: "+err.Error())
	}
	
}
