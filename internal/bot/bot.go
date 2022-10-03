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
				case connectCommand:
					b.ConnectCommand(update)
				case breakCommand:
					b.BreakCommand(update)
				case cancelCommand:
					b.CancelCommand(update)
				case getID:
					b.GetID(update)
				case setGroup:
					b.SetGroup(update)
				case event:
					b.Event(update)
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
			if userState != roomState {
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
		"🤖 Ваше повідомлення було отримано, якщо у вас є що додати - напишіть.\n\nСкоро до вас доєднається оператор технічної підтримки",
	)
	_, err = b.bot.Send(msg)
	if err != nil {
		return err
	}

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
