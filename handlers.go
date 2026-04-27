package main

import (
	"context"
	"fmt"
	"strings"

	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func messages(ctx context.Context, b *bot.Bot, update *models.Update) {
	users := getAllUsersInUpdate(update)
	if len(users) > 0 {
		newMember(ctx, b, update, users...)
		return
	}

	if update.Message == nil {
		return
	}

	ch, hasChallenge := getChallenge(update.Message.From.ID)

	if hasChallenge {
		if strings.EqualFold(strings.TrimSpace(getMessage(update.Message)), strings.TrimSpace(ch.Answer)) {
			approveUser(ctx, b, update.Message.Chat.ID, *update.Message.From)
		} else {
			banUser(ctx, b, update.Message.Chat.ID, *update.Message.From)
		}
	}
}

func newMember(ctx context.Context, b *bot.Bot, update *models.Update, users ...*models.User) {
	for _, member := range users {
		// Elimina cualquier Bot
		// TODO: construir lista blanca de Bots permitidos
		if member.IsBot {
			banUser(ctx, b, update.Message.Chat.ID, *member)
			continue
		}
		newChallenge(ctx, b, update.Message.Chat.ID, *member)
	}
}

func newChallenge(ctx context.Context, b *bot.Bot, chatID int64, user models.User) {
	ch := obtenerRetoBinding()

	msg, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      ch.Question + "\n\nResponde exactamente el nombre correcto **tiene 60 segundos**.\n\n" + fmt.Sprintf("Pista:\n```elisp\n(message \"%%s\" (lookup-key (current-global-map) (kbd \"%s\")))```", ch.Key),
		ParseMode: models.ParseModeMarkdown,
	})
	if err != nil {
		log.Printf("no se pudo enviar el reto al usuario, error: %v", err)
		return
	}

	ch = Challenge{
		Question:   ch.Question,
		Answer:     ch.Answer,
		MessageIDs: []int{msg.ID},
	}

	saveChallenge(user.ID, ch)
}

func banUser(ctx context.Context, b *bot.Bot, chatID int64, user models.User) {
	_, err := b.BanChatMember(ctx, &bot.BanChatMemberParams{
		ChatID:         chatID,
		UserID:         user.ID,
		RevokeMessages: true,
		UntilDate:      0,
	})
	if err != nil {
		log.Printf("Error al expulsar: %v\n", err)
		return
	} else {
		log.Printf("Usuario %d (%s) expulsado\n", user.ID, userModel2Name(user))
	}

	ch, hasChallenge := getChallenge(user.ID)
	if !hasChallenge {
		return
	}

	// delete the messages sent to the user
	for _, mid := range ch.MessageIDs {
		b.DeleteMessage(ctx, &bot.DeleteMessageParams{
			ChatID:    chatID,
			MessageID: mid,
		})
	}

	err = deleteChallenge(user.ID)
	if err != nil {
		log.Printf("no se pudo borrar reto, error: %v", err)
	}
}

func approveUser(ctx context.Context, b *bot.Bot, chatID int64, user models.User) {
	ch, hasChallenge := getChallenge(user.ID)
	if !hasChallenge {
		return
	}

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

func getAllUsersInUpdate(update *models.Update) []*models.User {
	var users []*models.User

	if update.Message != nil {
		for _, user := range update.Message.NewChatMembers {
			users = append(users, &user)
		}
	}

	return users
}
