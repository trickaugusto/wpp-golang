package main

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"go.mau.fi/whatsmeow"

	waE2E "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
)

func eventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		//fmt.Println("Received a message!", v.Message.GetMessageContextInfo())
		fmt.Println("Received a message!", v.Message.GetConversation())
	}
}

func main() {
	dbLog := waLog.Stdout("Database", "DEBUG", true)
	container, err := sqlstore.New("sqlite3", "file:db/wppteste.db?_foreign_keys=on", dbLog)
	if err != nil {
		panic(err)
	}
	// If you want multiple sessions, remember their JIDs and use .GetDevice(jid) or .GetAllDevices() instead.
	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		panic(err)
	}
	clientLog := waLog.Stdout("Client", "DEBUG", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)
	defer client.Disconnect()

	client.AddEventHandler(eventHandler)

	fmt.Println("Starting...")
	fmt.Println("Client ID:", client.Store.ID)

	if client.Store.ID == nil {
		// No ID stored, new login
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			panic(err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				// Render the QR code here
				// e.g. qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				// or just manually `echo 2@... | qrencode -t ansiutf8` in a terminal
				fmt.Println("QR code:", evt.Code)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		// Already logged in, just connect
		err = client.Connect()
		if err != nil {
			panic(err)
		}

		sendMessage(client, formatNumberWpp("41998765432"), "teste")
	}
}

func formatNumberWpp(numero string) string {
	numero = strings.ReplaceAll(numero, " ", "")
	numero = strings.ReplaceAll(numero, "-", "")

	if len(numero) == 11 {
		numero = "55" + numero
	}

	re := regexp.MustCompile(`^55(\d{2})9(\d{8})$`)
	if re.MatchString(numero) {
		numeroCorrigido := re.ReplaceAllString(numero, "55$1$2")
		fmt.Println("Número corrigido:", numeroCorrigido)
		return numeroCorrigido
	}

	return numero
}

func sendMessage(client *whatsmeow.Client, phoneNumber string, message string) {
	toJID := types.NewJID(phoneNumber, types.DefaultUserServer)

	response, err := client.SendMessage(context.Background(), toJID, &waE2E.Message{
		Conversation: proto.String(message),
	})
	if err != nil {
		fmt.Println("Erro ao enviar a mensagem:", err)
	} else {
		fmt.Println("Mensagem enviada com sucesso! ID:", response.ID)
	}
}
