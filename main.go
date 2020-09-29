package main

import (
	"os"

	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/naoina/toml"
	"github.com/ziutek/mymysql/autorc"
	_ "github.com/ziutek/mymysql/thrsafe"
)

func main() {

	type Config struct {
		Telegram struct {
			Token string
		}
		Database struct {
			Server, Conn, Laddr, Username, Password, Dbname string
		}
	}

	f, err := os.Open("config.toml")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	var config Config
	toml.NewDecoder(f).Decode(&config)

	db := autorc.New(config.Database.Conn, config.Database.Laddr, config.Database.Server, config.Database.Username, config.Database.Password, config.Database.Dbname)
	db.Register("set names utf8")

	bot, _ = tgbotapi.NewBotAPI(config.Telegram.Token) // write bot to global var, in next version bot will instead be passed every time

	startBot(db)

}
