package main

import (
	"encoding/json"
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
	"github.com/siddontang/go-mysql/client"
	"github.com/skip2/go-qrcode"
)

var (
	desist      = make(map[string](chan bool))
	connections = make(map[string]*whatsapp.Conn)
)

func startConnection(userid string, db *client.Conn) {

	desist[userid] = make(chan bool)

	//create new WhatsApp connection
	wac, err := whatsapp.NewConn(5 * time.Second)
	wac.SetClientVersion(2, 2021, 4)
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
				print("Fixing 403...")
				i.Db.Execute("UPDATE `wtg` SET `session` = '' WHERE `wtg`.`user_id` = ?;", i.userID)
				startConnection(userid, db)
				return
			}
		}
		i.sendAlertToTelegram("❌ <b>Fatal error logging in:</b> " + err.Error() + "\nMake sure your phone is connected, and if you still get this error, contact @MassiveBox")
		return
	}
	i.sendAlertToTelegram("✅ <b>Session started!</b> Make sure to keep your phone connected to the internet to receive and send messages.")

	connections[userid] = wac
	var shutgo bool

	r, err := db.Execute("SELECT premium FROM `wtg` WHERE user_id = ?;", userid)
	if err != nil {

	}
	premium, _ := r.GetInt(0, 0)
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
			i.sendAlertToTelegram("❌ <b>Your session has expired!</b> Please open another session from the /start menu.\nClick on the \"Pro\" button in the start menu to get life-lasting sessions - For free!")
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
					if failedReconns == 0 {
						i.sendAlertToTelegram("⚠️ <b>Your device is disconnected!</b> Make sure that your phone is turned on, connected to the internet, and with WhatsApp Web active.")
					}
					if failedReconns == 1440 {
						i.sendAlertToTelegram("❌ <b>Your device has been disconnected for the past 12 hours.</b> Please open a new session.")
						desist[userid] <- true
					}
					if failedReconns%120 == 0 && failedReconns != 0 {
						i.sendAlertToTelegram("⚠️ <b>Your device is still disconnected!</b> You won't be able to receive or send messages.")
					}
					time.Sleep(30 * time.Second)
					failedReconns++
				} else {
					if failedReconns > 0 {
						i.sendAlertToTelegram("✅ <b>Device reconnected</b>.")
					}
					time.Sleep(50 * time.Second)
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
		desist[userid] <- false
		return
	}

}

type informations struct {
	userID    string
	StartTime int64
	Db        *client.Conn
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
			}
		}
	} else {
		log.Printf("error occoured: %v\n", err)
		if strings.Contains(err.Error(), "when not logged in") {
			desist[s.userID] <- true
		}
	}
}

func (s *waHandler) HandleTextMessage(message whatsapp.TextMessage) {
	if int64(message.Info.Timestamp) > s.StartTime && message.Info.FromMe == false {
		cont := getAllContext(message.Info, message.ContextInfo, s.userID, s.wac)
		cont.MessageText = message.Text
		whatsappToTelegram(cont)
	}
}

func (s *waHandler) HandleImageMessage(message whatsapp.ImageMessage) {
	if int64(message.Info.Timestamp) > s.StartTime && message.Info.FromMe == false {
		cont := getAllContext(message.Info, message.ContextInfo, s.userID, s.wac)
		media, err := message.Download()
		if err == nil {
			ioutil.WriteFile("./img"+s.userID+".png", media, 0644)
			cont.MediaType = "image"
			cont.MediaCaption = message.Caption
			whatsappToTelegram(cont)
		}
	}
}

func (s *waHandler) HandleStickerMessage(message whatsapp.StickerMessage) {
	if int64(message.Info.Timestamp) > s.StartTime && message.Info.FromMe == false {
		cont := getAllContext(message.Info, message.ContextInfo, s.userID, s.wac)
		media, err := message.Download()
		if err == nil {
			ioutil.WriteFile("./img"+s.userID+".png", media, 0644)
			cont.MediaType = "image"
			cont.MediaCaption = "_This is a sticker_"
			whatsappToTelegram(cont)
		}
	}
}

func (s *waHandler) HandleVideoMessage(message whatsapp.VideoMessage) {
	if int64(message.Info.Timestamp) > s.StartTime && message.Info.FromMe == false {
		cont := getAllContext(message.Info, message.ContextInfo, s.userID, s.wac)
		media, err := message.Download()
		if err == nil {
			random := strconv.Itoa(rand.New(rand.NewSource(time.Now().UnixNano())).Intn(9999))
			ioutil.WriteFile("./vid"+s.userID+"-"+random+".mp4", media, 0644)
			cont.FileName = random
			cont.MediaType = "video"
			cont.MediaCaption = message.Caption
			whatsappToTelegram(cont)
		}
	}
}

func (s *waHandler) HandleAudioMessage(message whatsapp.AudioMessage) {
	if int64(message.Info.Timestamp) > s.StartTime && message.Info.FromMe == false {
		cont := getAllContext(message.Info, message.ContextInfo, s.userID, s.wac)
		media, err := message.Download()
		if err == nil {
			ioutil.WriteFile("./aud"+s.userID+".mp3", media, 0644)
			cont.MediaType = "audio"
			whatsappToTelegram(cont)
		}
	}
}

func (s *waHandler) HandleDocumentMessage(message whatsapp.DocumentMessage) {
	if int64(message.Info.Timestamp) > s.StartTime && message.Info.FromMe == false {
		cont := getAllContext(message.Info, message.ContextInfo, s.userID, s.wac)
		media, err := message.Download()
		if err == nil {
			ioutil.WriteFile("./doc"+s.userID+"-"+message.FileName, media, 0644)
			cont.FileName = message.FileName
			cont.MediaType = "document"
			whatsappToTelegram(cont)
		}
	}
}

func (i informations) login(wac *whatsapp.Conn) error {
	//load saved session
	session, err := i.readSession()
	if err == nil {
		//restore session
		session, err = wac.RestoreWithSession(session)
		if err != nil {
			return fmt.Errorf("restoring failed: %v", err)
		}
	} else {
		//no saved session -> regular login
		qr := make(chan string)
		go func() {
			qrData := <-qr
			qrcode.WriteFile(qrData, qrcode.Medium, 256, "qr"+i.userID+".png")
			i.sendQr()
		}()
		session, err = wac.Login(qr)
		if err != nil {
			return fmt.Errorf("error during login: %v", err)
		}
	}

	//save session
	err = i.writeSession(session)
	if err != nil {
		return fmt.Errorf("error saving session: %v", err)
	}
	return nil
}

func (i informations) readSession() (whatsapp.Session, error) {
	r, err := i.Db.Execute("SELECT session FROM `wtg` WHERE `user_id` = ?;", i.userID)
	if err != nil {
		return whatsapp.Session{}, err
	}
	marshald, _ := r.GetString(0, 0)
	var session whatsapp.Session
	err = json.Unmarshal([]byte(marshald), &session)
	if err != nil {
		return whatsapp.Session{}, err
	}
	return session, nil
}

func (i informations) writeSession(session whatsapp.Session) error {
	marshald, err := json.Marshal(session)
	if err != nil {
		return err
	}
	_, err = i.Db.Execute("UPDATE `wtg` SET `session` = ? WHERE `wtg`.`user_id` = ?;", marshald, i.userID)
	return err
}
