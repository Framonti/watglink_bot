# Whatsapp-Telegram Linker Bot

This is repository containing the official code for the Telegram bot [@WaTgLink_Bot](https://t.me/WaTgLink_Bot), allowing users to send and receive WhatsApp messages on Telegram.  

Please note that this bot is still in beta stage. It has not been tested with a large amount of users connected at the same time, and it hasn't been tested for concurrency-related problems.  

It's also one of my first complex projects in Go, so feel free to open a issue/pr if you find unoptimized code or any kind of other issue.  

# How to run locally

- Download all the dependencies

  ​	github.com/Rhymen/go-whatsapp  
  ​	github.com/siddontang/go-mysql/client  
  ​	github.com/go-telegram-bot-api/telegram-bot-api/v5  
  ​	github.com/gofiber/fiber  

- Put all the files of this repository in a local folder

- Set the correct database credentials and bot token: line 20 in telegram.go, line 16 in bridge.go

- Start your MySql server and execute the following query inside the database you selected in line 20 of telegram.go:

  ```CREATE TABLE `wtg` ( `id` INT NOT NULL AUTO_INCREMENT , `username` VARCHAR(64) NOT NULL , `user_id` INT(10) NOT NULL , `autoreply` TEXT NOT NULL , `premium` TINYINT NOT NULL DEFAULT '0' , `session` TEXT NOT NULL , PRIMARY KEY (`id`)) ENGINE = InnoDB; ```

- Make sure localhost:12 is open, and set your server (ex: Nginx, or if you're just testing, use Ngrok) to route all traffic from https://your.domain.com/whateveryouwant to localhost:12/rp/12

- Set the Telegram webhook by opening this link and replacing the url and the bot token with the real one you have set: https://api.telegram.org/botYOURBOTTOKEN/setWebhook?url=https://https://your.domain.com/whateveryouwant

- Execute the command `go run .` to start the bot. Check the terminal to see if errors occur. You might need to use `sudo` to listen on port 12.

If you manage to create your instance of Whatsapp-Telegram linker, make sure to DM me or open a issue. I will be happy to link your instance here!

# Contribute

Pull requests are always welcome! Feel free to for this repo, do your changes, and open a PR so I can check and approve it.
You will be credited in the contributors list, which is currently not present as it would be empty.  

If you want to donate, check out the [donations page.](https://massivebox.eu.org/?page=4 ) Thank you!