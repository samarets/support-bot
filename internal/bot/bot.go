package bot

import (
	"fmt"

	"github.com/dgraph-io/badger/v3"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/samarets/support-bot/internal/translations"

	"github.com/samarets/support-bot/internal/db"
	"github.com/samarets/support-bot/internal/log"
)

type bot struct {
	bot     *tgbotapi.BotAPI
	db      *botDB
	tl      *translations.Translator
	adminID int64
}

func InitBot(token string, adminID int64, translator *translations.Translator, db *db.DB) error {
	botAPI, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return err
	}

	dbState := newBotDB(db)

	botState := &bot{
		bot:     botAPI,
		db:      dbState,
		tl:      translator,
		adminID: adminID,
	}
	botState.InitUpdates()

	return nil
}

func (b *bot) InitUpdates() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.SentFrom() != nil {
			user := update.SentFrom()
			err := b.db.languageDB().set(user.ID, user.LanguageCode)
			if err != nil {
				log.Error().Err(err).Send()
			}
		}

		if update.CallbackQuery != nil {
			if !b.hasRight(update.CallbackQuery.From.ID) {
				callback := tgbotapi.NewCallback(
					update.CallbackQuery.ID,
					b.tl.GetMessageWithoutPrefix(b.db.languageDB().get(update.SentFrom().ID), "no_rights"),
				)
				_, err := b.bot.Request(callback)
				if err != nil {
					log.Error().Err(err).Send()
					continue
				}

				continue
			}

			callbackKey, userID, err := parseCallbackData(update.CallbackData())
			if err != nil {
				log.Error().Err(err).Send()
				continue
			}

			switch callbackKey {
			case acceptCallback:
				b.AcceptCallback(update, update.CallbackQuery.From.ID, userID)
			case declineCallback:
				b.DeclineCallback(update, update.CallbackQuery.From.ID, userID)
			}

			continue
		}

		if update.Message != nil {
			if update.Message.MigrateToChatID != 0 && b.db.groupDB().get() == update.Message.Chat.ID {
				err := b.db.groupDB().set(update.Message.MigrateToChatID)
				if err != nil {
					log.Error().Err(err).Send()
				}
			}

			userState, err := b.db.getUserState(update.SentFrom().ID)
			if err != nil {
				log.Error().Err(err).Send()
				continue
			}

			if update.Message.IsCommand() {
				switch update.Message.Command() {
				case startCommand:
					b.StartCommand(update, userState)
				case breakCommand:
					b.BreakCommand(update)
				case getIDCommand:
					b.GetID(update)
				case setGroupCommand:
					b.SetGroup(update)
				case addSupportCommand:
					b.AddSupport(update)
				case delSupportCommand:
					b.DelSupport(update)
				case getSupportsCommand:
					b.GetSupports(update)
				}

				continue
			}

			if !update.FromChat().IsPrivate() {
				continue
			}

			if userState == defaultState {
				err = b.defaultStateFunc(update)
				if err != nil {
					log.Error().Err(err).Send()
				}

				continue
			}
			if userState == queueState {
				err = b.queueStateFunc(update)
				if err != nil {
					log.Error().Err(err).Send()
				}

				continue
			}
			if userState == roomState {
				err = b.roomStateFunc(update)
				if err != nil {
					log.Error().Err(err).Send()
				}

				continue
			}

		}
	}
}

func (b *bot) defaultStateFunc(update tgbotapi.Update) error {
	err := b.db.queueDB().set(update.Message.From.ID, update.Message.From)
	if err != nil {
		return err
	}

	var bufferMessages []tgbotapi.Message
	bufferMessages = append(bufferMessages, *update.Message)
	err = b.db.bufferDB().set(update.Message.From.ID, bufferMessages)
	if err != nil {
		return err
	}

	msg := tgbotapi.NewMessage(
		update.Message.From.ID,
		b.tl.GetMessage(b.db.languageDB().get(update.SentFrom().ID), "got_message"),
	)
	_, err = b.bot.Send(msg)
	if err != nil {
		return err
	}

	b.sendSupportRequest(update)

	return nil
}

func (b *bot) queueStateFunc(update tgbotapi.Update) error {
	bufferMessages, err := b.db.bufferDB().get(update.Message.From.ID)
	if err != nil {
		return err
	}

	bufferMessages = append(bufferMessages, *update.Message)

	err = b.db.bufferDB().set(update.Message.From.ID, bufferMessages)
	if err != nil {
		return err
	}

	return nil
}

func (b *bot) roomStateFunc(update tgbotapi.Update) error {
	whoomSend, err := b.db.roomsDB().get(update.Message.From.ID)
	if err != nil {
		return err
	}
	if whoomSend == nil {
		return fmt.Errorf("whoomSend is empty")
	}

	msg := tgbotapi.NewCopyMessage(*whoomSend, update.Message.From.ID, update.Message.MessageID)
	if update.Message.ReplyToMessage != nil {
		replyToID, err := b.db.messagesIDsDB().get(update.Message.ReplyToMessage.MessageID)
		if err != nil && err != badger.ErrKeyNotFound {
			return err
		}

		if replyToID != nil {
			msg.ReplyToMessageID = *replyToID
		}
	}

	rMsg, err := b.bot.Send(msg)
	if err != nil {
		return err
	}

	// todo: if err == bot was blocked by the user then break the connection

	err = b.db.messagesIDsDB().set(update.Message.MessageID, rMsg.MessageID)
	if err != nil {
		return err
	}

	err = b.db.messagesIDsDB().set(rMsg.MessageID, update.Message.MessageID)
	if err != nil {
		return err
	}

	return nil
}

func (b *bot) sendSupportRequest(update tgbotapi.Update) {
	groupID := b.db.groupDB().get()
	if groupID == 0 {
		groupID = b.adminID
	}

	msg := tgbotapi.NewMessage(
		groupID,
		b.tl.GetMessage(
			"", "new_appeal", map[string]interface{}{
				"UserID": update.SentFrom().ID,
			},
		),
	)

	var keyboardRows [][]tgbotapi.InlineKeyboardButton
	buttonFirst := tgbotapi.NewInlineKeyboardButtonData(
		b.tl.GetMessageWithoutPrefix("", "confirm_appeal"),
		createCallbackData(acceptCallback, update.SentFrom().ID),
	)
	buttonSecond := tgbotapi.NewInlineKeyboardButtonData(
		b.tl.GetMessageWithoutPrefix("", "decline_appeal"),
		createCallbackData(declineCallback, update.SentFrom().ID),
	)
	keyboard := tgbotapi.NewInlineKeyboardRow(buttonFirst, buttonSecond)
	keyboardRows = append(keyboardRows, keyboard)

	msg.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: keyboardRows,
	}
	msg.ParseMode = tgbotapi.ModeMarkdown

	_, err := b.bot.Send(msg)
	if err != nil {
		log.Error().Err(err).Send()
		return
	}
}

func (b *bot) hasRight(userID int64) bool {
	if userID == b.adminID || b.db.supportDB().get(userID) {
		return true
	}

	return false
}
