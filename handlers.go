package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func messages(ctx context.Context, b *bot.Bot, update *models.Update) {
	events := getAllUsersInUpdate(update)
	if len(events) > 0 {
		newMember(ctx, b, events...)
		return
	}

	if update.Message == nil {
		return
	}

	ch, hasChallenge := getChallenge(update.Message.From.ID)

	if hasChallenge {
		if strings.EqualFold(strings.TrimSpace(getMessage(update.Message)), ch.Answer) {
			approveUser(ctx, b, update.Message.Chat.ID, *update.Message.From, update.Message.ID)
		} else {
			banUser(ctx, b, update.Message.From.ID, update.Message.ID)
		}
	}
}

func newMember(ctx context.Context, b *bot.Bot, events ...NewUserEvent) {
	for _, event := range events {
		// Elimina cualquier Bot
		// TODO: construir lista blanca de Bots permitidos
		if event.User.IsBot {
			banUser(ctx, b, event.User.ID, -1)
			continue
		}
		newChallenge(ctx, b, event.Chat, *event.User)
	}
}

func newChallenge(ctx context.Context, b *bot.Bot, chatID int64, user models.User) {
	ch := obtenerRetoBinding()

	msgA, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   fmt.Sprintf("Hola, %s", userModel2Name(user)),
	})
	if err != nil {
		log.Printf("no se pudo enviar saludo al usuario, error: %v", err)
		return
	}
	msg, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text: ch.Question + "\n\nResponde exactamente el nombre correcto **tiene 60 segundos**\\.\n\n" +
			fmt.Sprintf("Pista:\n```shell\nemacs -q --batch --eval '(princ (format \"%%s\" (lookup-key (current-global-map) (kbd \"%s\"))))'```", ch.Key),
		ParseMode: models.ParseModeMarkdown,
	})
	if err != nil {
		log.Printf("no se pudo enviar el reto al usuario, error: %v", err)
		return
	}

	ch = Challenge{
		Question:   ch.Question,
		Answer:     ch.Answer,
		MessageIDs: []int{msgA.ID, msg.ID},
		UserID:     user.ID,
		UserName:   userModel2Name(user),
		ChatID:     chatID,
		CreatedAt:  time.Now().UTC(),
	}

	saveChallenge(user.ID, ch)
}

func banUser(ctx context.Context, b *bot.Bot, userID int64, answerMsg int) {
	ch, ok := getChallenge(userID)
	if !ok {
		log.Printf("el usuario %s no tenia un reto, igual será expulsado", ch.UserName)
	}

	_, err := b.BanChatMember(ctx, &bot.BanChatMemberParams{
		ChatID:    ch.ChatID,
		UserID:    ch.UserID,
		UntilDate: 0,
	})
	if err != nil {
		log.Printf("Error al expulsar: %v\n", err)
		return
	} else {
		log.Printf("Usuario %d (%s) expulsado\n", ch.UserID, ch.UserName)
	}

	if answerMsg > 0 {
		ch.MessageIDs = append(ch.MessageIDs, answerMsg)
	}

	// delete the messages sent to the user
	for _, mid := range ch.MessageIDs {
		b.DeleteMessage(ctx, &bot.DeleteMessageParams{
			ChatID:    ch.ChatID,
			MessageID: mid,
		})
	}

	if ok {
		err = deleteChallenge(userID)
		if err != nil {
			log.Printf("no se pudo borrar reto, error: %v", err)
		}
	}
}

func approveUser(ctx context.Context, b *bot.Bot, chatID int64, user models.User, answerMsg int) {
	ch, hasChallenge := getChallenge(user.ID)
	if !hasChallenge {
		return
	}

	ch.MessageIDs = append(ch.MessageIDs, answerMsg)

	// delete the messages sent to the user
	for _, mid := range ch.MessageIDs {
		b.DeleteMessage(ctx, &bot.DeleteMessageParams{
			ChatID:    chatID,
			MessageID: mid,
		})
	}

	err := deleteChallenge(user.ID)
	if err != nil {
		log.Printf("no se pudo borrar reto, error: %v", err)
	}
}

func userModel2Name(user models.User) string {
	return fmt.Sprintf("%s %s [@%s]", user.FirstName, user.LastName, user.Username)
}

func getMessage(msg *models.Message) string {
	if msg == nil {
		return ""
	}

	return msg.Text
}

type NewUserEvent struct {
	User *models.User
	Chat int64
}

func getAllUsersInUpdate(update *models.Update) []NewUserEvent {
	var users []NewUserEvent

	if update.Message != nil {
		for _, user := range update.Message.NewChatMembers {
			users = append(users, NewUserEvent{
				User: &user,
				Chat: update.Message.Chat.ID,
			})
		}
	}

	if update.ChatMember != nil {
		member := update.ChatMember
		if member.OldChatMember.Type == "left" && member.NewChatMember.Type == "member" {
			users = append(users, NewUserEvent{
				User: &member.From,
				Chat: member.Chat.ID,
			})
		}
	}

	return users
}
