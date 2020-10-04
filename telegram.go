package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gofiber/fiber/v2"
	"github.com/ziutek/mymysql/autorc"
	_ "github.com/ziutek/mymysql/thrsafe"
)

var bot *tgbotapi.BotAPI
var language = "english" /* todo save&load this variable to match the user selected language (e.g. from db)
 	right now, it's a testing variable
	Also, decide which is the best file to declare this var */

func startBot(db *autorc.Conn) {

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	go reopenSessions(db)

	var startKeyb = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(readStringJSON(language, "session"), "/session"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(readStringJSON(language, "pro"), "/pro"),
			tgbotapi.NewInlineKeyboardButtonURL(readStringJSON(language, "github"),
				readStringJSON("url", "githubUrl")),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL(readStringJSON(language, "privacy"),
				readStringJSON("url", "privacyUrl")),
			tgbotapi.NewInlineKeyboardButtonURL(readStringJSON(language, "usageCondition"),
				readStringJSON("url", "usageConditionUrl")),
		),
	)
	var backKeyb = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Back", "/start"),
		),
	)
	const startMsg = "💻 <b>Welcome to Whatsapp-Telegram Linker.</b>\n\nWith this bot, you can connect Telegram and Whatsapp to receive WhatsApp messages on Telegram, and you can respond to WhatsApp messages directly via Telegram.\n\n🤖 <b>Proudly developed by @MassiveBox</b> - <a href='https://massivebox.eu.org/?page=4'>Donate</a>\n⚠️ The bot is still in beta! If you find something not working, have patience and report it."

	app.Post("/rp/12", func(c *fiber.Ctx) error {

		var update tgbotapi.Update
		json.Unmarshal([]byte(c.Body()), &update)

		if update.Message != nil {

			if update.Message.Chat.ID < 0 {
				return nil
			}

			if update.Message.Text == "/start" {

				user, _, err := db.Query("SELECT id,username FROM `wtg` WHERE user_id = %d;", update.Message.Chat.ID)
				if err != nil {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Internal error establishing a connection to the database. Please try again"))
					return err
				}
				var (
					username string
					id       int
				)
				if len(user) >= 1 {
					if len(user[0]) == 2 {
						id = user[0].Int(0)
						username = user[0].Str(1)
					}
				}
				if id == 0 {
					db.Query("INSERT INTO `wtg` (`id`, `username`, `user_id`, `autoreply`, `premium`, `session`) VALUES (NULL, '%s', '%d', '', '0', '');", update.Message.From.UserName, update.Message.Chat.ID)
				}
				if username == "ban" {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "You're banned from the bot."))
					return nil
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
					return nil
				}

				var entities []tgbotapi.MessageEntity
				if update.Message.ReplyToMessage.Entities != nil {
					entities = update.Message.ReplyToMessage.Entities
				}
				if update.Message.ReplyToMessage.CaptionEntities != nil {
					entities = update.Message.ReplyToMessage.CaptionEntities
				}

				if entities != nil {
					url := entities[0].URL
					matchs := regexp.MustCompile(`(?m)https://a.aa/(.*)/(.*)/(.*)`).FindStringSubmatch(url)
					if len(matchs) == 4 {
						telegramToWhatsapp(update.Message.Text, matchs[1], matchs[2], matchs[3], update.Message.From.ID)
					} else {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "🧐 <b>This doesn't look like a WhatsApp message...</b>")
						msg.ReplyToMessageID = update.Message.MessageID
						msg.ParseMode = "HTML"
						bot.Send(msg)
						return nil
					}
				}

			}

			if strings.Contains(update.Message.Text, "/post ") && update.Message.From.ID == 1334403986 {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Starting to send post..."))
				postTxt := strings.Replace(update.Message.Text, "/post ", "", 1)
				rows, _, _ := db.Query("SELECT user_id FROM `wtg`")
				for key, row := range rows {
					userID := row.Int(0)
					msg := tgbotapi.NewMessage(int64(userID), postTxt)
					msg.ParseMode = "HTML"
					bot.Send(msg)
					if key%10 == 0 && key > 0 {
						time.Sleep(5 * time.Second)
					}
				}
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Post sent."))
			}

			if update.Message.ReplyToMessage == nil && update.Message.Text != "/start" && strings.Contains(update.Message.Text, "/post ") == false {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "↪️ You must <b>reply to a message</b> in order to send a text to WhatsApp.")
				msg.ReplyToMessageID = update.Message.MessageID
				msg.ParseMode = "HTML"
				bot.Send(msg)
			}

		}

		if update.CallbackQuery != nil {

			if update.CallbackQuery.Data == "/start" {
				go bot.Send(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))
				msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, startMsg)
				msg.ReplyMarkup = &startKeyb
				msg.ParseMode = "HTML"
				bot.Send(msg)
			}

			if update.CallbackQuery.Data == "/pro" {

				resp, err := http.Get("https://api.botsarchive.com/getBotID.php?username=@" + bot.Self.UserName)
				if err != nil {
					return nil
				}
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return nil
				}

				type botsarchiveData struct {
					Ok      int    `json:"ok"`
					ID      int    `json:"id"`
					Message string `json:"message"`
					Result  struct {
						Msg string `json:"msg"`
					} `json:"result"`
				}

				var botdata botsarchiveData
				json.Unmarshal(body, &botdata)

				if botdata.Ok == 0 || botdata.ID == 0 {

					go bot.Send(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))
					msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "🔰 With the <b>Pro function</b> you will be able to extend the length of your sessions and get other features for free.\n\nDuring the beta test phase, the Pro option is not active.\nStar our GitHub repo to keep yourself updated.")
					msg.ReplyMarkup = &backKeyb
					msg.ParseMode = "HTML"
					bot.Send(msg)

				} else {

					go bot.Send(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))
					rows, _, err := db.Query("SELECT premium FROM `wtg` WHERE user_id = %d;", update.CallbackQuery.Message.Chat.ID)
					if err != nil {
						return nil
					}

					var premium int
					if len(rows) >= 1 {
						premium = rows[0].Int(0)
					}

					if premium == 1 {
						bot.Send(tgbotapi.NewCallbackWithAlert(update.CallbackQuery.ID, "You're already pro!"))
						return nil
					}
					go bot.Send(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))

					var keyb = tgbotapi.NewInlineKeyboardMarkup(
						tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonURL("⭐️ Rate ⭐️", botdata.Result.Msg),
						),
						tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData("🔙 Back", "/start"),
							tgbotapi.NewInlineKeyboardButtonData("✅ Check", "/check_pro "+strconv.Itoa(botdata.ID)),
						),
					)
					msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "😳 <b>Rate us 5 stars on BotsArchive to unlock Pro features!</b>\n\nClick on the link below, join the channel, click on the five stars button, then Confirm to unlock Pro.\n"+botdata.Result.Msg)
					msg.ReplyMarkup = &keyb
					msg.ParseMode = "HTML"
					bot.Send(msg)

				}

			}

			if strings.Contains(update.CallbackQuery.Data, "/check_pro ") {

				botid := strings.ReplaceAll(update.CallbackQuery.Data, "/check_pro ", "")

				resp, err := http.Get("https://api.botsarchive.com/getUserVote.php?bot_id=" + botid + "&user_id=" + strconv.Itoa(update.CallbackQuery.From.ID))
				if err != nil {
					return nil
				}
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return nil
				}

				type voteData struct {
					Ok     int    `json:"ok"`
					Result string `json:"result"`
				}

				var vote voteData
				json.Unmarshal(body, &vote)

				if vote.Result == "" {
					if vote.Ok == 1 {
						bot.Send(tgbotapi.NewCallbackWithAlert(update.CallbackQuery.ID, "You haven't voted!"))
					} else {
						bot.Send(tgbotapi.NewCallbackWithAlert(update.CallbackQuery.ID, "Internal error, please try again."))
					}
				} else {
					if vote.Result == "5" {
						go bot.Send(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))
						msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "🥳 <b>Thanks!</b> Pro is now enabled.\nIf you change your rating, it will get revoked.")
						msg.ReplyMarkup = &backKeyb
						msg.ParseMode = "HTML"
						bot.Send(msg)
						db.Query("UPDATE `wtg` SET `premium` = '1' WHERE `wtg`.`user_id` = %d;", update.CallbackQuery.Message.Chat.ID)
					} else {
						bot.Send(tgbotapi.NewCallbackWithAlert(update.CallbackQuery.ID, "Select five stars."))
					}
				}

			}

			if update.CallbackQuery.Data == "/session" || update.CallbackQuery.Data == "/session_r" {

				var keyb = tgbotapi.NewInlineKeyboardMarkup(
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
						sessiondesc = "⏸ <b>Paused</b> - You're logged in, but your session is inactive."
					} else {
						pong, err := wac.AdminTest()
						if pong && err == nil {
							sessiondesc = "✅ <b>Connected</b> - Your session is active."
						} else {
							sessiondesc = "💔 <b>Disconnected</b> - You're logged in, but your session is inactive because your phone isn't connected."
						}
					}
				}

				msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "🎛 Session is: "+sessiondesc)
				msg.ReplyMarkup = &keyb
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
						status = "active"
					}
				}

				if action == "start" {
					switch status {
					case "active":
						bot.Send(tgbotapi.NewCallbackWithAlert(update.CallbackQuery.ID, "Session is already active!"))
					case "noauth":
						go bot.Send(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))
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
						msg.ParseMode = "HTML"
						msg.ReplyMarkup = &keyb
						bot.Send(msg)
						go bot.Send(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))
						desist[strconv.Itoa(update.CallbackQuery.From.ID)] <- true
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
						go bot.Send(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))
						var keyb = tgbotapi.NewInlineKeyboardMarkup(
							tgbotapi.NewInlineKeyboardRow(
								tgbotapi.NewInlineKeyboardButtonData("🔙 Back", "/session"),
							),
						)
						msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "🗑 <b>Session deleted.</b> When you decide to start again your session, you will be asked to re-scan the QR code.")
						msg.ReplyMarkup = &keyb
						msg.ParseMode = "HTML"
						bot.Send(msg)
						db.Query("UPDATE `wtg` SET `session` = '' WHERE `wtg`.`user_id` = %d;", update.CallbackQuery.Message.Chat.ID)
					}
				}

			}

		}

		return nil

	})

	app.Get("/rp/12/ping", func(c *fiber.Ctx) error {
		// This is just a dummy route for making your uptime monitor's life  a little better.
		c.SendString("Hello!")
		return nil
	})

	log.Fatal(app.Listen(":12"))

}
