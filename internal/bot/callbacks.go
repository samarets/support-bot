package bot

import (
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/samarets/support-bot/internal/log"
)

const (
	acceptCallback  = "accept"
	declineCallback = "decline"
)

func (b *bot) AcceptCallback(update tgbotapi.Update, supportID, userID int64) {
	st, err := b.db.getUserState(supportID)
	if err != nil {
		log.Error().Err(err).Send()
		return
	}
	if st != defaultState {
		callback := tgbotapi.NewCallback(
			update.CallbackQuery.ID,
			b.tl.GetMessageWithoutPrefix(b.db.languageDB().get(supportID), "complete_last_appeal"),
		)
		_, err := b.bot.Request(callback)
		if err != nil {
			log.Error().Err(err).Send()
			return
		}

		return
	}

	st, err = b.db.getUserState(userID)
	if err != nil {
		log.Error().Err(err).Send()
		return
	}
	if st != queueState {
		callback := tgbotapi.NewCallback(
			update.CallbackQuery.ID,
			b.tl.GetMessageWithoutPrefix(b.db.languageDB().get(supportID), "user_no_expect"),
		)
		_, err := b.bot.Request(callback)
		if err != nil {
			log.Error().Err(err).Send()
			return
		}

		return
	}

	user, err := b.db.queueDB().get(userID)
	if err != nil {
		log.Error().Err(err).Send()
		return
	}
	if user == nil {
		return
	}

	msg := tgbotapi.NewMessage(
		userID,
		b.tl.GetMessage(b.db.languageDB().get(userID), "operator_connected"),
	)
	msgUser, err := b.bot.Send(msg)
	if err != nil && err.Error() == botWasBlockedError {
		msg = tgbotapi.NewMessage(
			supportID,
			b.tl.GetMessage(
				b.db.languageDB().get(supportID), "user_connected", map[string]interface{}{
					"Name": user.FirstName + " " + user.LastName,
					"ID":   user.ID,
				},
			),
		)
		msg.ParseMode = tgbotapi.ModeMarkdownV2
		_, err = b.bot.Send(msg)
		if err != nil {
			log.Error().Err(err).Send()
			callback := tgbotapi.NewCallback(
				update.CallbackQuery.ID,
				b.tl.GetMessageWithoutPrefix(b.db.languageDB().get(supportID), "server_error"),
			)
			_, err := b.bot.Request(callback)
			if err != nil {
				log.Error().Err(err).Send()
				return
			}

			return
		}

		updateMsg := tgbotapi.NewEditMessageText(
			update.FromChat().ID,
			update.CallbackQuery.Message.MessageID,
			b.tl.GetMessageWithoutPrefix(
				"", "user_block_bot_update", map[string]interface{}{
					"LastMessage": update.CallbackQuery.Message.Text,
					"Name":        update.SentFrom().FirstName + " " + update.SentFrom().LastName,
					"ID":          supportID,
				},
			),
		)
		updateMsg.ParseMode = tgbotapi.ModeMarkdown
		_, err = b.bot.Request(updateMsg)
		if err != nil {
			log.Error().Err(err).Send()
			return
		}

		err = b.db.queueDB().delete(userID)
		if err != nil {
			log.Error().Err(err).Send()
			return
		}

		err = b.sendBufferMessages(supportID, userID)
		if err != nil {
			log.Error().Err(err).Send()
			return
		}

		msg := tgbotapi.NewMessage(supportID, b.tl.GetMessage(b.db.languageDB().get(supportID), "user_block_bot"))
		_, err = b.bot.Send(msg)
		if err != nil {
			log.Error().Err(err).Send()
			return
		}

		return
	} else if err != nil {
		log.Error().Err(err).Send()
		return
	}

	msg = tgbotapi.NewMessage(
		supportID,
		b.tl.GetMessage(
			b.db.languageDB().get(supportID), "user_connected", map[string]interface{}{
				"Name": user.FirstName + " " + user.LastName,
				"ID":   user.ID,
			},
		),
	)
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	_, err = b.bot.Send(msg)
	if err != nil && err.Error() == botWasBlockedError {
		msgDelete := tgbotapi.NewDeleteMessage(msgUser.Chat.ID, msgUser.MessageID)
		_, err = b.bot.Request(msgDelete)
		if err != nil {
			log.Error().Err(err).Send()
			return
		}

		callback := tgbotapi.NewCallback(
			update.CallbackQuery.ID,
			b.tl.GetMessageWithoutPrefix(b.db.languageDB().get(supportID), "you_block_bot"),
		)
		_, err := b.bot.Request(callback)
		if err != nil {
			log.Error().Err(err).Send()
			return
		}

		return
	} else if err != nil {
		log.Error().Err(err).Send()
		return
	}

	updateMsg := tgbotapi.NewEditMessageText(
		update.FromChat().ID,
		update.CallbackQuery.Message.MessageID,
		b.tl.GetMessageWithoutPrefix(
			"", "success_accept_appeal", map[string]interface{}{
				"LastMessage": update.CallbackQuery.Message.Text,
				"Name":        update.SentFrom().FirstName + " " + update.SentFrom().LastName,
				"ID":          supportID,
			},
		),
	)
	updateMsg.ParseMode = tgbotapi.ModeMarkdown
	_, err = b.bot.Request(updateMsg)
	if err != nil {
		log.Error().Err(err).Send()
		return
	}

	err = b.db.roomsDB().set(userID, supportID)
	if err != nil {
		log.Error().Err(err).Send()
		return
	}

	err = b.db.roomsDB().set(supportID, userID)
	if err != nil {
		log.Error().Err(err).Send()
		return
	}

	err = b.db.queueDB().delete(userID)
	if err != nil {
		log.Error().Err(err).Send()
		return
	}

	err = b.sendBufferMessages(supportID, userID)
	if err != nil {
		log.Error().Err(err).Send()
		return
	}
}

func (b *bot) sendBufferMessages(supportID, userID int64) error {
	bufferMessages, err := b.db.bufferDB().get(userID)
	if err != nil {
		return err
	}

	for _, message := range bufferMessages {
		msg := tgbotapi.NewCopyMessage(
			supportID,
			message.Chat.ID,
			message.MessageID,
		)
		rMsg, err := b.bot.Send(msg)
		if err != nil {
			return err
		}

		err = b.db.messagesIDsDB().set(message.MessageID, rMsg.MessageID)
		if err != nil {
			return err
		}

		err = b.db.messagesIDsDB().set(rMsg.MessageID, message.MessageID)
		if err != nil {
			return err
		}
	}

	err = b.db.bufferDB().delete(userID)
	if err != nil {
		return err
	}

	return nil
}

func (b *bot) DeclineCallback() {

}

func createCallbackData(key string, userID int64) string {
	return fmt.Sprintf("%s-%d", key, userID)
}

func parseCallbackData(data string) (string, int64, error) {
	dataSplit := strings.Split(data, "-")
	if len(dataSplit) != 2 {
		return "", 0, fmt.Errorf("bad callback data")
	}

	userID, err := strconv.ParseInt(dataSplit[1], 10, 64)
	if err != nil {
		return "", 0, err
	}

	return dataSplit[0], userID, nil
}
