package bot

import (
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/samarets/support-bot/internal/log"
)

const (
	startCommand = "start"
	breakCommand = "break"
	getIDCommand = "get_id"

	setGroupCommand    = "set_group"
	addSupportCommand  = "add_support"
	delSupportCommand  = "del_support"
	getSupportsCommand = "get_supports"
)

func (b *bot) StartCommand(update tgbotapi.Update, userState state) {
	if !update.FromChat().IsPrivate() {
		return
	}

	var message string
	switch userState {
	case queueState:
		message = b.tl.GetMessage(b.db.languageDB().get(update.SentFrom().ID), "queue_start")
	case roomState:
		message = b.tl.GetMessage(b.db.languageDB().get(update.SentFrom().ID), "room_start")
	default:
		message = b.tl.GetMessage(
			b.db.languageDB().get(update.SentFrom().ID), "hello", map[string]interface{}{
				"Name": update.SentFrom().FirstName,
			},
		)
	}

	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		message,
	)

	_, err := b.bot.Send(msg)
	if err != nil {
		log.Error().Err(err).Send()
	}
}

func (b *bot) BreakCommand(update tgbotapi.Update) {
	if !b.hasRight(update.SentFrom().ID) {
		return
	}

	whomBreak, err := b.db.roomsDB().get(update.SentFrom().ID)
	if err != nil {
		log.Error().Err(err).Send()
		return
	}
	if whomBreak == nil {
		return
	}

	err = b.db.roomsDB().delete(*whomBreak)
	if err != nil {
		log.Error().Err(err).Send()
		return
	}

	err = b.db.roomsDB().delete(update.SentFrom().ID)
	if err != nil {
		log.Error().Err(err).Send()
		return
	}

	msg := tgbotapi.NewMessage(
		update.SentFrom().ID,
		b.tl.GetMessage(b.db.languageDB().get(update.SentFrom().ID), "chat_end"),
	)
	_, err = b.bot.Send(msg)
	if err != nil {
		log.Error().Err(err).Send()
		return
	}

	msg = tgbotapi.NewMessage(
		*whomBreak,
		b.tl.GetMessage(b.db.languageDB().get(*whomBreak), "operator_leave"),
	)
	_, err = b.bot.Send(msg)
	if err != nil {
		log.Error().Err(err).Send()
		return
	}
}

func (b *bot) GetID(update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(b.tl.Prefix+" %d", update.SentFrom().ID))
	_, err := b.bot.Send(msg)
	if err != nil {
		log.Error().Err(err).Send()
		return
	}
}

func (b *bot) SetGroup(update tgbotapi.Update) {
	user := update.SentFrom()
	if user.ID != b.adminID {
		return
	}

	err := b.db.groupDB().set(update.FromChat().ID)
	if err != nil {
		log.Error().Err(err).Send()
		return
	}

	msg := tgbotapi.NewMessage(
		update.FromChat().ID,
		b.tl.GetMessage(b.db.languageDB().get(user.ID), "channel_saved"),
	)
	_, err = b.bot.Send(msg)
	if err != nil {
		log.Error().Err(err).Send()
		return
	}
}

func (b *bot) AddSupport(update tgbotapi.Update) {
	if update.SentFrom().ID != b.adminID {
		return
	}

	var userID int64
	if update.Message.ReplyToMessage != nil && !update.Message.ReplyToMessage.From.IsBot {
		if update.Message.ReplyToMessage.From.ID == b.adminID {
			return
		}
		userID = update.Message.ReplyToMessage.From.ID
	} else {
		argumentID, err := strconv.ParseInt(update.Message.CommandArguments(), 10, 64)
		if err != nil {
			msg := tgbotapi.NewMessage(
				update.Message.Chat.ID,
				b.tl.GetMessage(b.db.languageDB().get(update.SentFrom().ID), "add_support_fail"),
			)
			msg.ReplyToMessageID = update.Message.MessageID
			msg.ParseMode = tgbotapi.ModeMarkdown
			_, err := b.bot.Send(msg)
			if err != nil {
				log.Error().Err(err).Send()
				return
			}
			return
		}
		if argumentID == b.adminID {
			return
		}
		userID = argumentID
	}

	err := b.db.supportDB().set(userID, true)
	if err != nil {
		log.Error().Err(err).Send()
		return
	}

	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		b.tl.GetMessage(
			b.db.languageDB().get(update.SentFrom().ID), "add_support_success", map[string]interface{}{
				"UserID": userID,
			},
		),
	)
	msg.ReplyToMessageID = update.Message.MessageID
	_, err = b.bot.Send(msg)
	if err != nil {
		log.Error().Err(err).Send()
		return
	}
}

func (b *bot) DelSupport(update tgbotapi.Update) {
	if update.SentFrom().ID != b.adminID {
		return
	}

	var userID int64
	if update.Message.ReplyToMessage != nil && !update.Message.ReplyToMessage.From.IsBot {
		if update.Message.ReplyToMessage.From.ID == b.adminID {
			return
		}
		userID = update.Message.ReplyToMessage.From.ID
	} else {
		argumentID, err := strconv.ParseInt(update.Message.CommandArguments(), 10, 64)
		if err != nil {
			msg := tgbotapi.NewMessage(
				update.Message.Chat.ID,
				b.tl.GetMessage(b.db.languageDB().get(update.SentFrom().ID), "del_support_fail"),
			)
			msg.ReplyToMessageID = update.Message.MessageID
			msg.ParseMode = tgbotapi.ModeMarkdown
			_, err := b.bot.Send(msg)
			if err != nil {
				log.Error().Err(err).Send()
				return
			}
			return
		}
		if argumentID == b.adminID {
			return
		}
		userID = argumentID
	}

	st, err := b.db.getUserState(userID)
	if err != nil {
		log.Error().Err(err).Send()
		return
	}

	if st == roomState {
		whomBreak, err := b.db.roomsDB().get(userID)
		if err != nil {
			log.Error().Err(err).Send()
			return
		}

		err = b.db.roomsDB().delete(*whomBreak)
		if err != nil {
			log.Error().Err(err).Send()
			return
		}

		err = b.db.roomsDB().delete(userID)
		if err != nil {
			log.Error().Err(err).Send()
			return
		}

		msg := tgbotapi.NewMessage(
			userID,
			b.tl.GetMessage(b.db.languageDB().get(userID), "chat_end"),
		)
		_, err = b.bot.Send(msg)
		if err != nil {
			log.Error().Err(err).Send()
			return
		}

		msg = tgbotapi.NewMessage(
			*whomBreak,
			b.tl.GetMessage(b.db.languageDB().get(*whomBreak), "operator_leave"),
		)
		_, err = b.bot.Send(msg)
		if err != nil {
			log.Error().Err(err).Send()
			return
		}
	}

	err = b.db.supportDB().set(userID, false)
	if err != nil {
		log.Error().Err(err).Send()
		return
	}

	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		b.tl.GetMessage(
			b.db.languageDB().get(update.SentFrom().ID), "del_support_success", map[string]interface{}{
				"UserID": userID,
			},
		),
	)
	msg.ReplyToMessageID = update.Message.MessageID
	_, err = b.bot.Send(msg)
	if err != nil {
		log.Error().Err(err).Send()
		return
	}
}

func (b *bot) GetSupports(update tgbotapi.Update) {
	supports, err := b.db.supportDB().getAll()
	if err != nil {
		log.Error().Err(err).Send()
		return
	}

	if len(supports) == 0 {
		msg := tgbotapi.NewMessage(
			update.Message.Chat.ID,
			b.tl.GetMessage(b.db.languageDB().get(update.SentFrom().ID), "supports_empty"),
		)
		msg.ReplyToMessageID = update.Message.MessageID
		_, err := b.bot.Send(msg)
		if err != nil {
			log.Error().Err(err).Send()
			return
		}

		return
	}

	var messageStr strings.Builder
	messageStr.WriteString(b.tl.GetMessage(b.db.languageDB().get(update.SentFrom().ID), "supports_message"))
	for i, userID := range supports {
		messageStr.WriteString(fmt.Sprintf("%d. [%d](tg://user?id=%d)\n", i+1, userID, userID))
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, messageStr.String())
	msg.ReplyToMessageID = update.Message.MessageID
	msg.ParseMode = tgbotapi.ModeMarkdown
	_, err = b.bot.Send(msg)
	if err != nil {
		log.Error().Err(err).Send()
		return
	}
}
