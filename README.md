# Whatsapp-Telegram Linker Bot

[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=MassiveBox_WaTgLink_Bot&metric=alert_status)](https://sonarcloud.io/dashboard?id=MassiveBox_WaTgLink_Bot) [![TravisCI](https://travis-ci.com/MassiveBox/WaTgLink_Bot.svg?branch=master)](https://travis-ci.com/github/MassiveBox/WaTgLink_Bot/builds/) [<img class="badge" tag="github.com/MassiveBox/WaTgLink_Bot" src="https://goreportcard.com/badge/github.com/MassiveBox/WaTgLink_Bot">](https://goreportcard.com/report/github.com/MassiveBox/WaTgLink_Bot) [![DM me badge](https://img.shields.io/badge/contact-@MassiveBox-blue?logo=telegram)](https://t.me/MassiveBox) [![Donate](https://img.shields.io/badge/support-the%20project-yellow?logo=symantec)](https://massivebox.eu.org/?page=4)

This is repository containing the official code for the Telegram bot [@WaTgLink_Bot](https://t.me/WaTgLink_Bot), allowing users to send and receive WhatsApp messages on Telegram.  

Please note that this bot is still in beta stage. It has not been tested with a large amount of users connected at the same time, and it hasn't been tested for concurrency-related problems.  

It's also one of my first complex projects in Go, so feel free to open a issue/pr if you find unoptimized code or any kind of other issue.  

# Legal

Whatsapp-Telegram Linker is neither associated with nor sponsored by WhatsApp Inc., Facebook Inc. or Telegram FZ-LLC. We offer a service based on WhatsApp Web API and Telegram Bot API.   
Whatsapp-Telegram linker is fully reliant on WhatsApp, and it doesn't represent a financial threat to it because it is not a WhatsApp Business replacement and it does not prevent users from consuming advertising contents offered on any WhatsApp application or WhatsApp Web itself - As there are none. If this condition changes, the service will be shut down.

Every copyrighted name, logo or media used in the code, in other assets of this repository, in promotional media or in the bot itself, is property of it(s) holder(s) and no copyright infringement is intended. 

If you represent WhatsApp Inc., Telegram FZ-LLC or an affiliated company and you have a legal complaint feel free to contact me with the email address legal@massivebox.eu.org - I will comply with all legitimate requests. Thank you.

Please read the [Privacy Policy](https://github.com/MassiveBox/WaTgLink_Bot/blob/master/docs/PRIVACY.md) and the [Usage Conditions](https://github.com/MassiveBox/WaTgLink_Bot/blob/master/docs/USAGE_CONDITIONS.md) before using the service.  

All the open source libraries and repositories used are listed in the go.mod file.

# How to run locally

First, decide if you want to run using the executable or from source.

|                        | Pre-built  | From Source |
| ---------------------- | ---------- | ----------- |
| Complete support       | ✅          | ✅           |
| Instant deploy         | ✅          | ❌           |
| Ability to change code | ❌          | ✅           |
| Stability              | Guaranteed | Might vary  |



### Pre-built executable

- Download the files from the last release [here](https://github.com/MassiveBox/watglink_bot/releases/) and put them in a folder where sudo can write and read files.

- Set the appropriate credentials in the `config.toml` file

- Start your MySQL server and execute the following query inside the database you selected in the `config.toml` file:

  ```CREATE TABLE `wtg` ( `id` INT NOT NULL AUTO_INCREMENT , `username` VARCHAR(64) NOT NULL , `user_id` INT(10) NOT NULL , `autoreply` TEXT NOT NULL , `premium` TINYINT NOT NULL DEFAULT '0' , `session` TEXT NOT NULL , PRIMARY KEY (`id`)) ENGINE = InnoDB; ```

- Make sure localhost:12 is open, and set your server (ex: Nginx, or if you're just testing, use Ngrok) to route all traffic from https://your.domain.com/whateveryouwant to localhost:12/rp/12

- Set the Telegram webhook by opening this link and replacing the url and the bot token with the real one you have set: https://api.telegram.org/botYOURBOTTOKEN/setWebhook?url=https://https://your.domain.com/whateveryouwant

- Run the bot: `sudo ./start.sh` - Use nohup to keep the bot running even after you log off. If `start.sh` doesn't appear executable, use the command `chmod +x start.sh`

### Run from source

- Clone the repository in a local folder into your server
- (Optional) Remove useless files, including all `.MD`s, `.travis.yml`,  and the pre-built file `watg`.
- (Optional) Change the code as you please
- Re-build with `go build -race .` (The `-race` argument won't be necessary from next versions)
- Follow from step 2 of the explanation for the pre-built package.

If you manage to create your instance of Whatsapp-Telegram linker, make sure to DM me or open a issue. I will be happy to link your instance here!

# Contribute

Pull requests are always welcome! Feel free to for this repo, do your changes, and open a PR so I can check and approve it.
You will be credited in the contributors list, which is currently not present as it would be empty.  

If you want to donate, check out the [donations page.](https://massivebox.eu.org/?page=4 ) Thank you!