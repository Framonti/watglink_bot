package main

import (
	"encoding/json"
	"fmt"

	"github.com/Rhymen/go-whatsapp"
	"github.com/skip2/go-qrcode"
	"github.com/ziutek/mymysql/autorc"
	_ "github.com/ziutek/mymysql/thrsafe"
)

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
			qrcode.WriteFile(qrData, qrcode.Medium, 256, "qr"+i.UserID+".png")
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
	r, _, err := i.Db.Query("SELECT session FROM `wtg` WHERE `user_id` = %s;", i.UserID)
	if err != nil {
		return whatsapp.Session{}, err
	}
	var marshald string
	if len(r) >= 1 {
		marshald = r[0].Str(0)
	}
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
	_, _, err = i.Db.Query("UPDATE `wtg` SET `session` = '"+string(marshald)+"' WHERE `wtg`.`user_id` = %s;", i.UserID)
	return err
}

func reopenSessions(db *autorc.Conn) {
	rows, _, _ := db.Query("SELECT user_id FROM `wtg` WHERE session != ''")
	for _, row := range rows {
		userID := row.Str(0)
		informations{userID, 0, db}.sendAlertToTelegram("➰ <b>Bot was restarted</b> - Re-opening your session... Sorry for the inconvenience.")
		startConnection(userID, db)
	}
}
