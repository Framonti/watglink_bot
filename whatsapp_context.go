package main

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/Rhymen/go-whatsapp"
	"github.com/Rhymen/go-whatsapp/binary/proto"
)

type context struct {
	RemoteJid, SenderName, GroupName, QuotedMessageSender, QuotedMessageText, MessageID, userID string
	MessageProto                                                                                *proto.Message
	MessageText                                                                                 string

	MediaCaption string
	MediaType    string
	FileName     string
}

func getAllContext(info whatsapp.MessageInfo, wacontext whatsapp.ContextInfo, userid string, wac *whatsapp.Conn) context {
	var jid string
	if info.Source.Participant == nil {
		jid = info.RemoteJid
	} else {
		jid = *info.Source.Participant
	}
	senderName := wac.Store.Contacts[regexp.MustCompile(`(?m)(-.*@)`).ReplaceAllString(jid, "")].Name
	if senderName == "" {
		senderName = "+" + regexp.MustCompile(`(-.*|@.*)`).ReplaceAllString(jid, "")
	}
	var quotedMessageText, quotedMessageSender string
	if wacontext.QuotedMessage != nil && wacontext.QuotedMessage.Conversation != nil {
		quotedMessageText = *wacontext.QuotedMessage.Conversation
		quotedMessageSender = wac.Store.Contacts[regexp.MustCompile(`(?m)(-.*@)`).ReplaceAllString(wacontext.Participant, "")].Name
		if quotedMessageSender == "" {
			quotedMessageSender = "+" + regexp.MustCompile(`(-.*|@.*)`).ReplaceAllString(wacontext.Participant, "")
		}
	}

	groupMeta, err := wac.GetGroupMetaData(info.RemoteJid)
	if err != nil {
		fmt.Println("ERROR!", err.Error())
	}
	var groupMetaString string
	groupMetaString = <-groupMeta
	var meta map[string]string
	json.Unmarshal([]byte(groupMetaString), &meta)

	return context{RemoteJid: info.RemoteJid, SenderName: senderName, GroupName: meta["subject"], QuotedMessageSender: quotedMessageSender, QuotedMessageText: quotedMessageText, MessageID: info.Id, MessageProto: info.Source.GetMessage(), userID: userid}
}
