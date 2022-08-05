package bot

import (
	"fmt"

	"github.com/dgraph-io/badger/v3"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/samarets/support-bot/internal/log"
)

func (b *bot) StartCommand(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		"🤖 Привіт, напиши своє питання - ми допоможемо",
	)

	_, err := b.bot.Send(msg)
	if err != nil {
		log.Error().Err(err).Send()
	}
}

func (b *bot) ConnectCommand(update tgbotapi.Update) {
	var chatID int64
	var userTg tgbotapi.User

	err := b.db.Get(mergePrefixDB(queue, update.Message.Chat.ID), &userTg)
	if err != nil {
		if err != badger.ErrKeyNotFound {
			log.Error().Err(err).Send()
			return
		}
	}

	err = b.db.Get(mergePrefixDB(rooms, update.Message.Chat.ID), &chatID)
	if err != nil {
		if err != badger.ErrKeyNotFound {
			log.Error().Err(err).Send()
			return
		}
	}

	var user tgbotapi.User
	err = b.db.GetFirstWherePrefix([]byte(queue), &user)
	if err != nil && err != badger.ErrKeyNotFound {
		log.Error().Err(err).Send()
		return
	}

	var defaultUser tgbotapi.User
	if user == defaultUser {
		return
	}

	err = b.db.Set(mergePrefixDB(rooms, user.ID), update.Message.Chat.ID)
	if err != nil {
		log.Error().Err(err).Send()
		return
	}

	err = b.db.Set(mergePrefixDB(rooms, update.Message.Chat.ID), user.ID)
	if err != nil {
		log.Error().Err(err).Send()
		return
	}

	err = b.db.Drop(mergePrefixDB(queue, user.ID))
	if err != nil {
		log.Error().Err(err).Send()
		return
	}

	msg := tgbotapi.NewMessage(user.ID, "🤖 До вас доєднався оператор")
	_, err = b.bot.Send(msg)
	if err != nil {
		log.Error().Err(err).Send()
		return
	}

	msg = tgbotapi.NewMessage(
		update.Message.Chat.ID,
		fmt.Sprintf(
			"🤖 Ви були доєднані до користувача [%s](tg://user?id=%d)\nID: %d\nПитання користувача:",
			user.FirstName+" "+user.LastName,
			user.ID,
			user.ID,
		),
	)
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	_, err = b.bot.Send(msg)
	if err != nil {
		log.Error().Err(err).Send()
		return
	}

	var bufferMessages []tgbotapi.Message
	err = b.db.Get(mergePrefixDB(buffer, user.ID), &bufferMessages)
	if err != nil && err != badger.ErrKeyNotFound {
		log.Error().Err(err).Send()
		return
	}

	for _, message := range bufferMessages {
		msg := tgbotapi.NewCopyMessage(
			update.Message.Chat.ID,
			message.Chat.ID,
			message.MessageID,
		)
		rMsg, err := b.bot.Send(msg)
		if err != nil {
			log.Error().Err(err).Send()
			return
		}

		err = b.db.Set(mergePrefixDB(messagesIDs, message.MessageID), rMsg.MessageID)
		if err != nil {
			log.Error().Err(err).Send()
			continue
		}

		err = b.db.Set(mergePrefixDB(messagesIDs, rMsg.MessageID), message.MessageID)
		if err != nil {
			log.Error().Err(err).Send()
			continue
		}
	}

	err = b.db.Drop(mergePrefixDB(buffer, user.ID))
	if err != nil {
		log.Error().Err(err).Send()
		return
	}
}

func (b *bot) BreakCommand(update tgbotapi.Update) {
	var whomBreak int64
	err := b.db.Get(mergePrefixDB(rooms, update.Message.Chat.ID), &whomBreak)
	if err != nil {
		if err != badger.ErrKeyNotFound {
			log.Error().Err(err).Send()
			return
		}

		return
	}

	err = b.db.Drop(mergePrefixDB(rooms, whomBreak))
	if err != nil {
		log.Error().Err(err).Send()
		return
	}

	err = b.db.Drop(mergePrefixDB(rooms, update.Message.Chat.ID))
	if err != nil {
		log.Error().Err(err).Send()
		return
	}

	msg := tgbotapi.NewMessage(whomBreak, "🤖 Розмову з оператором було завершено")
	_, err = b.bot.Send(msg)
	if err != nil {
		log.Error().Err(err).Send()
		return
	}

	msg = tgbotapi.NewMessage(update.Message.Chat.ID, "🤖 Ви завершили розмову з користувачем")
	_, err = b.bot.Send(msg)
	if err != nil {
		log.Error().Err(err).Send()
		return
	}
}
