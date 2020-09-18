package main

import (
	"encoding/json"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gofiber/fiber"
	"github.com/siddontang/go-mysql/client"
)

func main() {

	app := fiber.New()

	db, _ := client.Connect("localhost:3306", "username", "password", "dbname")
	db.Ping()

	var startKeyb = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📦 Session", "/session"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔰 Pro", "/pro"),
			tgbotapi.NewInlineKeyboardButtonURL("🐱 GitHub", "https://github.com/MassiveBox/WaTgLink_Bot"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("🔐 Privacy", "https://telegra.ph/Privacy-Policy---Whatsapp-Telegram-Linker-09-16-2"),
			tgbotapi.NewInlineKeyboardButtonURL("🌡 Usage Conditions", "https://telegra.ph/Usage-Conditions---WhatsApp-Telegram-Linker-09-16"),
		),
	)
	const startMsg = "💻 <b>Welcome to Whatsapp-Telegram Linker.</b>\n\nWith this bot, you can connect Telegram and Whatsapp to receive WhatsApp messages on Telegram, and you can respond to WhatsApp messages directly via Telegram.\n\n🤖 <b>Proudly developed by @MassiveBox</b> - <a href='https://massivebox.eu.org/?page=4'>Donate</a>\n⚠️ The bot is still in beta! If you find something not working, have patience and report it."

	app.Post("/rp/12", func(c *fiber.Ctx) {

		var update tgbotapi.Update
		json.Unmarshal([]byte(c.Body()), &update)

		if update.Message != nil {

			if update.Message.Chat.ID < 0 {
				return
			}

			if update.Message.Text == "/start" {

				user, err := db.Execute("SELECT id,username FROM `bots`.`wtg` WHERE user_id = ?;", update.Message.Chat.ID)
				if err != nil {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Internal error establishing a connection to the database. Please try again"))
					return
				}
				id, _ := user.GetInt(0, 0)
				if id == 0 {
					db.Execute("INSERT INTO `wtg` (`id`, `username`, `user_id`, `autoreply`, `premium`, `session`) VALUES (NULL, ?, ?, '', '0', '');", update.Message.From.UserName, update.Message.Chat.ID)
				}
				username, _ := user.GetString(0, 1)
				if username == "ban" {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "You're banned from the bot."))
					return
				}

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, startMsg)
				msg.ReplyToMessageID = update.Message.MessageID
				msg.ParseMode = "HTML"
				msg.ReplyMarkup = startKeyb
				bot.Send(msg)

			}

			if update.Message.ReplyToMessage != nil {

				if update.Message.Text == "" {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "❌ <b>Format not allowed.</b> You can only send text messages at the moment.")
					msg.ReplyToMessageID = update.Message.MessageID
					msg.ParseMode = "HTML"
					bot.Send(msg)
					return
				}

				var entities []tgbotapi.MessageEntity
				if update.Message.ReplyToMessage.Entities != nil {
					entities = update.Message.ReplyToMessage.Entities
				}
				if update.Message.ReplyToMessage.CaptionEntities != nil {
					entities = update.Message.ReplyToMessage.CaptionEntities
				}
				log.Println(entities)
				if entities != nil {
					url := entities[0].URL
					matchs := regexp.MustCompile(`(?m)https://a.aa/(.*)/(.*)/(.*)`).FindStringSubmatch(url)
					log.Println(matchs, url)
					if len(matchs) == 4 {
						print("yaah\n")
						telegramToWhatsapp(update.Message.Text, matchs[1], matchs[2], matchs[3], db, update.Message.From.ID)
					}
				}

			}

			if update.Message.ReplyToMessage == nil && update.Message.Text != "/start" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "↪️ You must <b>reply to a message</b> in order to send a text to WhatsApp.")
				msg.ReplyToMessageID = update.Message.MessageID
				msg.ParseMode = "HTML"
				bot.Send(msg)
			}

		}

		if update.CallbackQuery != nil {

			if update.CallbackQuery.Data == "/start" {
				msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, startMsg)
				msg.ReplyMarkup = &startKeyb
				msg.ParseMode = "HTML"
				bot.Send(msg)
			}

			if update.CallbackQuery.Data == "/pro" {
				var startKeyb = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("🔙 Back", "/start"),
					),
				)
				msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "🔰 With the <b>Pro function</b> you will be able to extend the lenght of your sessions and get other features for free.\n\nDuring the beta test phase, the Pro option is not active.\nStar our GitHub repo to keep yourself updated.")
				msg.ReplyMarkup = &startKeyb
				msg.ParseMode = "HTML"
				bot.Send(msg)
			}

			if update.CallbackQuery.Data == "/session" || update.CallbackQuery.Data == "/session_r" {

				var startKeyb = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("▶️ Start", "/controller start"),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("⏸ Pause", "/controller pause"),
						tgbotapi.NewInlineKeyboardButtonData("🗑 Delete", "/controller delete"),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("🔙 Back", "/start"),
					),
				)

				var sessiondesc string
				idstring := strconv.Itoa(update.CallbackQuery.From.ID)
				i := informations{strconv.Itoa(update.CallbackQuery.From.ID), time.Now().Unix(), db}
				session, err := i.readSession()
				if session.ClientId == "" || err != nil {
					if update.CallbackQuery.Data == "/session" {
						bot.Send(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))
					} else {
						bot.Send(tgbotapi.NewCallback(update.CallbackQuery.ID, "Refreshed."))
					}
					sessiondesc = "⬛️ <b>Empty</b> - This means that you're not logged in, and when you start a session, you will be prompted to scan the QR code on your WhatsApp app."
				} else {
					bot.Send(tgbotapi.NewCallback(update.CallbackQuery.ID, "Loading..."))
					wac := connections[idstring]
					if wac == nil {
						sessiondesc = "💔 <b>Disconnected</b> - You're logged in, but your session is inactive."
					} else {
						pong, err := wac.AdminTest()
						if pong && err == nil {
							sessiondesc = "✅ <b>Connected</b> - Your session is active."
						} else {
							sessiondesc = "💔 <b>Disconnected</b> - You're logged in, but your session is inactive."
						}
					}
				}

				msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "🎛 Session is: "+sessiondesc)
				msg.ReplyMarkup = &startKeyb
				msg.ParseMode = "HTML"
				bot.Send(msg)

			}

			if strings.Contains(update.CallbackQuery.Data, "/controller ") {

				action := strings.Replace(update.CallbackQuery.Data, "/controller ", "", 1)

				var status string
				idstring := strconv.Itoa(update.CallbackQuery.From.ID)
				i := informations{strconv.Itoa(update.CallbackQuery.From.ID), time.Now().Unix(), db}
				session, err := i.readSession()
				if session.ClientId == "" || err != nil {
					status = "noauth"
				} else {
					wac := connections[idstring]
					if wac == nil {
						status = "inactive"
					} else {
						pong, err := wac.AdminTest()
						if pong && err == nil {
							status = "active"
						} else {
							status = "inactive"
						}
					}
				}

				if action == "start" {
					switch status {
					case "active":
						bot.Send(tgbotapi.NewCallbackWithAlert(update.CallbackQuery.ID, "Session is already active!"))
					case "noauth":
						var keyb = tgbotapi.NewInlineKeyboardMarkup(
							tgbotapi.NewInlineKeyboardRow(
								tgbotapi.NewInlineKeyboardButtonData("✅ Proceed", "/controller start-ok"),
								tgbotapi.NewInlineKeyboardButtonData("🔙 Back", "/session"),
							),
						)
						msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "🟢 When clicking Proceed, <b>you will be prompted to scan a QR code.</b>\n\nOpen the WhatsApp app on your smartphone, and if you're on Android, click the buttons in the upper-right corner, then Whatsapp Web; or if you have a iOS device, go into Settings, then click on WhatsApp web\n\nBy clicking Proceed, you confirm that you've accepted our Conditions of Usage and you're aware of our Privacy Policies, accessible from the main menu of this bot.")
						msg.ReplyMarkup = &keyb
						msg.ParseMode = "HTML"
						bot.Send(msg)
					case "inactive":
						bot.Send(tgbotapi.NewDeleteMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID))
						go startConnection(strconv.Itoa(update.CallbackQuery.From.ID), db)
					}
				}

				if action == "start-ok" && status == "noauth" {
					bot.Send(tgbotapi.NewDeleteMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID))
					go startConnection(strconv.Itoa(update.CallbackQuery.From.ID), db)
				}

				if action == "pause" {
					switch status {
					case "active":
						var keyb = tgbotapi.NewInlineKeyboardMarkup(
							tgbotapi.NewInlineKeyboardRow(
								tgbotapi.NewInlineKeyboardButtonData("🔙 Back", "/session"),
							),
						)
						msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "⏸ <b>Session paused.</b> You can restart it in the Session menu.")
						msg.ReplyMarkup = &keyb
						msg.ParseMode = "HTML"
						bot.Send(msg)
						log.Println("Sendo true")
						desist[strconv.Itoa(update.CallbackQuery.From.ID)] <- true
						log.Println("Fatto")
					default:
						bot.Send(tgbotapi.NewCallbackWithAlert(update.CallbackQuery.ID, "Session is not active!"))
					}
				}

				if action == "delete" {
					switch status {
					case "active":
						bot.Send(tgbotapi.NewCallbackWithAlert(update.CallbackQuery.ID, "Session is active, you must pause it first."))
					case "noauth":
						bot.Send(tgbotapi.NewCallbackWithAlert(update.CallbackQuery.ID, "No session to delete."))
					case "inactive":
						var keyb = tgbotapi.NewInlineKeyboardMarkup(
							tgbotapi.NewInlineKeyboardRow(
								tgbotapi.NewInlineKeyboardButtonData("🔙 Back", "/session"),
							),
						)
						msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "🗑 <b>Session deleted.</b> When you decide to start again your session, you will be asked to re-scan the QR code.")
						msg.ReplyMarkup = &keyb
						msg.ParseMode = "HTML"
						bot.Send(msg)
						db.Execute("UPDATE `wtg` SET `session` = '' WHERE `wtg`.`user_id` = ?;", update.CallbackQuery.Message.Chat.ID)
					}
				}

			}

		}

	})

	log.Fatal(app.Listen(12))

}
