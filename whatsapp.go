package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Rhymen/go-whatsapp"
	"github.com/ziutek/mymysql/autorc"
	_ "github.com/ziutek/mymysql/thrsafe"
)

var (
	desist      = make(map[string](chan bool))
	connections = make(map[string]*whatsapp.Conn)
)

func startConnection(userid string, db *autorc.Conn) {

	desist[userid] = make(chan bool)

	//create new WhatsApp connection
	wac, err := whatsapp.NewConn(5 * time.Second)
	wac.SetClientVersion(2, 2039, 9)
	wac.SetClientName("WaTgLink Bot official", "watglink")

	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating connection: %v\n", err)
		return
	}

	i := informations{userid, time.Now().Unix(), db}
	wac.AddHandler(&waHandler{wac, i})

	err = i.login(wac)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error logging in: %v\n", err)
		if strings.Contains(err.Error(), "admin login responded with") {
			if strings.Contains(err.Error(), "403") || strings.Contains(err.Error(), "401") {
				i.sendAlertToTelegram("🔧 <b>Recovering from Unauthorized error...</b> We will send you a new QR code to scan shortly, please prepare your smartphone.")
				i.Db.Query("UPDATE `wtg` SET `session` = '' WHERE `wtg`.`user_id` = %s;", i.UserID)
				startConnection(userid, db)
				return
			}
		}
		if err.Error() == "error during login: qr code scan timed out" {
			i.sendAlertToTelegram("🕔 <b>QR timed out!</b> Please open a new session, and prepare your smartphone before clicking on the Proceed button in order to be able to scan the image in time.")
			return
		}
		i.sendAlertToTelegram("❌ <b>Fatal error logging in:</b> " + err.Error() + "\nMake sure your phone is connected, and if you still get this error, contact @MassiveBox")
		return
	}
	i.sendAlertToTelegram("✅ <b>Session started!</b> Make sure to keep your phone connected to the internet to receive and send messages.")

	connections[userid] = wac
	var shutgo bool

	rows, _, err := db.Query("SELECT premium FROM `wtg` WHERE user_id = %s;", userid)
	if err != nil {
		i.sendAlertToTelegram("Error connecting to the database, please try again by opening a new session.")
		return
	}
	var premium int
	if len(rows) >= 1 {
		premium = rows[0].Int(0)
	}
	if premium == 0 {
		i.sendAlertToTelegram("⏰ Your session will be terminated in <b>three hours!</b> Please click the \"Pro\" button in the /start menu to know how to remove all time limits for free.")
		go func() {
			for x := 0; x < 18; x++ {
				if shutgo == false {
					time.Sleep(10 * time.Minute)
				} else {
					return
				}
			}
			if shutgo == false {
				i.sendAlertToTelegram("❌ <b>Your session has expired!</b> Please open another session from the /start menu.\nClick on the \"Pro\" button in the start menu to get life-lasting sessions - For free!")
			}
			desist[userid] <- true
		}()
	}

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		if shutgo == false {
			i.sendAlertToTelegram("❌ <b>Admin has closed your session.</b> This could be due to a upgrade of the bot, so please open a new session after waiting a couple of minutes.")
			print("Disconnecting...\n")
			session, _ := wac.Disconnect()
			i.writeSession(session)
		}
		os.Exit(1)
	}()

	go func() {
		var failedReconns int
		for true {
			if shutgo == false {
				pong, err := wac.AdminTest()
				if !pong || err != nil {
					time.Sleep(50 * time.Second)
					if failedReconns == 0 {
						i.sendAlertToTelegram("⚠️ <b>Your device is disconnected!</b> Check your phone signal. <a href='https://github.com/MassiveBox/watglink_bot/blob/master/docs/DISCONNECTED.md'>Help</a>")
					}
					if failedReconns == 1440 {
						i.sendAlertToTelegram("❌ <b>Session closed due to excessive inactivity.</b> Please open a new session.")
						desist[userid] <- true
					}
					if failedReconns%120 == 0 && failedReconns != 0 && failedReconns != 1440 {
						i.sendAlertToTelegram("⚠️ <b>Your device is still disconnected!</b> You won't be able to receive or send messages.")
					}
					time.Sleep(30 * time.Second)
					failedReconns++
				} else {
					if failedReconns > 0 {
						i.sendAlertToTelegram("✅ <b>Device reconnected</b>.")
					}
					failedReconns = 0
				}
			} else {
				return
			}
		}
	}()

	select {
	case <-desist[userid]:
		session, _ := wac.Disconnect()
		i.writeSession(session)
		shutgo = true
		connections[userid] = nil
		desist[userid] <- false
		return
	}

}

type informations struct {
	UserID    string
	StartTime int64
	Db        *autorc.Conn
}

type waHandler struct {
	wac *whatsapp.Conn
	informations
}

//HandleError needs to be implemented to be a valid WhatsApp handler
func (s *waHandler) HandleError(err error) {
	if e, ok := err.(*whatsapp.ErrConnectionFailed); ok {
		for x := 0; x < 10; x++ {
			log.Printf("Connection failed, underlying error: %v", e.Err)
			log.Println("Waiting 30sec...")
			<-time.After(30 * time.Second)
			log.Println("Reconnecting...")
			err := s.wac.Restore()
			if err != nil {
				fmt.Printf("Restore failed: %v", err)
				if err.Error() == "invalid session" {
					s.sendAlertToTelegram("❌ <b>Fatal error:</b> Runtime error, session is invalid. Please open a new session, you will be prompted to scan a new QR code.")
					s.Db.Query("UPDATE `wtg` SET `session` = '' WHERE `wtg`.`user_id` = %s;", s.UserID)
					desist[s.UserID] <- true
				}
			}
		}
	} else {
		log.Printf("error occoured: %v\n", err)
		if strings.Contains(err.Error(), "when not logged in") {
			desist[s.UserID] <- true
		}
	}
}

func (s *waHandler) HandleTextMessage(message whatsapp.TextMessage) {
	if int64(message.Info.Timestamp) > s.StartTime && message.Info.FromMe == false {
		cont := getAllContext(message.Info, message.ContextInfo, s.UserID, s.wac)
		cont.MessageText = message.Text
		whatsappToTelegram(cont)
	}
}

func (s *waHandler) HandleImageMessage(message whatsapp.ImageMessage) {
	if int64(message.Info.Timestamp) > s.StartTime && message.Info.FromMe == false {
		cont := getAllContext(message.Info, message.ContextInfo, s.UserID, s.wac)
		media, err := message.Download()
		if err == nil {
			ioutil.WriteFile("./img"+s.UserID+".png", media, 0644)
			cont.MediaType = "image"
			cont.MediaCaption = message.Caption
			whatsappToTelegram(cont)
		}
	}
}

func (s *waHandler) HandleStickerMessage(message whatsapp.StickerMessage) {
	if int64(message.Info.Timestamp) > s.StartTime && message.Info.FromMe == false {
		cont := getAllContext(message.Info, message.ContextInfo, s.UserID, s.wac)
		media, err := message.Download()
		if err == nil {
			ioutil.WriteFile("./img"+s.UserID+".png", media, 0644)
			cont.MediaType = "image"
			cont.MediaCaption = "_This is a sticker_"
			whatsappToTelegram(cont)
		}
	}
}

func (s *waHandler) HandleVideoMessage(message whatsapp.VideoMessage) {
	if int64(message.Info.Timestamp) > s.StartTime && message.Info.FromMe == false {
		cont := getAllContext(message.Info, message.ContextInfo, s.UserID, s.wac)
		media, err := message.Download()
		if err == nil {
			random := strconv.Itoa(rand.New(rand.NewSource(time.Now().UnixNano())).Intn(9999))
			ioutil.WriteFile("./vid"+s.UserID+"-"+random+".mp4", media, 0644)
			cont.FileName = random
			cont.MediaType = "video"
			cont.MediaCaption = message.Caption
			whatsappToTelegram(cont)
		}
	}
}

func (s *waHandler) HandleAudioMessage(message whatsapp.AudioMessage) {
	if int64(message.Info.Timestamp) > s.StartTime && message.Info.FromMe == false {
		cont := getAllContext(message.Info, message.ContextInfo, s.UserID, s.wac)
		media, err := message.Download()
		if err == nil {
			ioutil.WriteFile("./aud"+s.UserID+".mp3", media, 0644)
			cont.MediaType = "audio"
			whatsappToTelegram(cont)
		}
	}
}

func (s *waHandler) HandleDocumentMessage(message whatsapp.DocumentMessage) {
	if int64(message.Info.Timestamp) > s.StartTime && message.Info.FromMe == false {
		cont := getAllContext(message.Info, message.ContextInfo, s.UserID, s.wac)
		media, err := message.Download()
		if err == nil {
			ioutil.WriteFile("./doc"+s.UserID+"-"+message.FileName, media, 0644)
			cont.FileName = message.FileName
			cont.MediaType = "document"
			whatsappToTelegram(cont)
		}
	}
}
