package main

import (
	"fmt"
	"log"
	"time"

	"github.com/rezaDastrs/internal/models"
	mail "github.com/xhit/go-simple-mail/v2"
)

func listenForMail() {
	//run in background
	go func() {
		for {
			msg := <-app.MailChan
			sendMesssage(msg)
		}
	}()
}

func sendMesssage(m models.MailData) {
	server := mail.NewSMTPClient()
	//localhost for sending dummy email
	server.Host = "localhost"
	//1025 is port for sending dummy email
	server.Port = 1025
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	client, err := server.Connect()
	if err != nil {
		log.Println(err)
	}

	//mail format to send
	email := mail.NewMSG()
	email.SetFrom(m.From).AddTo(m.To).SetSubject(m.Subject)
	email.SetBody(mail.TextHTML, m.Content)

	//send email
	err = email.Send(client)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("email sent !")
	}
}
