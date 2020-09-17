package main

import (
	"encoding/json"
	"regexp"
	"fmt"

	"github.com/Rhymen/go-whatsapp"
	"github.com/Rhymen/go-whatsapp/binary/proto"
)


type Context struct {
	RemoteJid, SenderName, GroupName, QuotedMessageSender, QuotedMessageText, MessageID, UserID string
	MessageProto *proto.Message
	MessageText string
	
	MediaCaption string
	MediaType string
	FileName string
}

func GetAllContext(info whatsapp.MessageInfo, context whatsapp.ContextInfo, userid string, wac *whatsapp.Conn) Context {
	var jid string
	if info.Source.Participant == nil {
		jid = info.RemoteJid
	}else{
		jid = *info.Source.Participant
	}
	senderName := wac.Store.Contacts[regexp.MustCompile(`(?m)(-.*@)`).ReplaceAllString(jid, "")].Name
	if senderName == "" {
		senderName = "+"+regexp.MustCompile(`(-.*|@.*)`).ReplaceAllString(jid, "")
	}
	var quotedMessageText, quotedMessageSender string
	if context.QuotedMessage != nil && context.QuotedMessage.Conversation != nil {
		quotedMessageText = *context.QuotedMessage.Conversation
		quotedMessageSender = wac.Store.Contacts[regexp.MustCompile(`(?m)(-.*@)`).ReplaceAllString(context.Participant, "")].Name
		if quotedMessageSender == "" {
			quotedMessageSender = "+"+regexp.MustCompile(`(-.*|@.*)`).ReplaceAllString(context.Participant, "")
		}
	}
	
	groupMeta, err := wac.GetGroupMetaData(info.RemoteJid)
	if err != nil {
		fmt.Println("ERROR! %v", err)
	}
	var groupMetaString string
	groupMetaString = <- groupMeta
	var meta map[string]string
	json.Unmarshal([]byte(groupMetaString), &meta)
	
	fmt.Println("\n", meta, "\n")
	//fmt.Println("Timestamp:", message.Info.Timestamp, "ID:", message.Info.Id, "JID remoto:", message.Info.RemoteJid, "JID sender:", message.Info.SenderJid,  "QuotedMessageID:", message.ContextInfo.QuotedMessageID, "Testo:", message.Text, "Sender:", senderName, "Group name", meta["subject"])
		
	return Context{RemoteJid: info.RemoteJid, SenderName: senderName, GroupName: meta["subject"], QuotedMessageSender: quotedMessageSender, QuotedMessageText: quotedMessageText, MessageID: info.Id, MessageProto: info.Source.GetMessage(), UserID: userid}
}
