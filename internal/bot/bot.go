package bot

import (
	"fmt"

	"github.com/dgraph-io/badger/v3"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/samarets/support-bot/internal/db"
	"github.com/samarets/support-bot/internal/log"
)

type bot struct {
	bot *tgbotapi.BotAPI
	db  *botDB
}

func InitBot(token string, db *db.DB) error {
	botAPI, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return err
	}

	dbState := newBotDB(db)

	botState := &bot{
		bot: botAPI,
		db:  dbState,
	}
	botState.InitUpdates()

	return nil
}

func (b *bot) InitUpdates() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			if update.Message.IsCommand() {
				switch update.Message.Command() {
				case startCommand:
					b.StartCommand(update)
				case connectCommand:
					b.ConnectCommand(update)
				case breakCommand:
					b.BreakCommand(update)
				case cancelCommand:
					b.CancelCommand(update)
				}

				continue
			}

			fromQueue, err := b.db.queueDB().get(update.Message.From.ID)
			if err != nil {
				log.Error().Err(err).Send()
				continue
			}

			roomsID, err := b.db.roomsDB().get(update.Message.From.ID)
			if err != nil {
				log.Error().Err(err).Send()
				continue
			}

			if fromQueue == nil && roomsID == nil {
				err = b.db.queueDB().set(update.Message.From.ID, update.Message.From)
				if err != nil {
					log.Error().Err(err).Send()
					continue
				}

				var bufferMessages []tgbotapi.Message
				bufferMessages = append(bufferMessages, *update.Message)
				err = b.db.bufferDB().set(update.Message.From.ID, bufferMessages)
				if err != nil {
					log.Error().Err(err).Send()
					continue
				}

				msg := tgbotapi.NewMessage(
					update.Message.Chat.ID,
					"🤖 Ваше повідомлення було отримано, якщо у вас є що додати - напишіть.\n\nСкоро до вас доєднається оператор технічної підтримки",
				)
				_, err = b.bot.Send(msg)
				if err != nil {
					log.Error().Err(err).Send()
					continue
				}

				continue
			}

			if fromQueue != nil && roomsID == nil {
				bufferMessages, err := b.db.bufferDB().get(update.Message.From.ID)
				if err != nil {
					log.Error().Err(err).Send()
					continue
				}

				bufferMessages = append(bufferMessages, *update.Message)

				err = b.db.bufferDB().set(update.Message.From.ID, bufferMessages)
				if err != nil {
					log.Error().Err(err).Send()
					continue
				}

				continue
			}

			if roomsID != nil {
				whoomSend, err := b.db.roomsDB().get(update.Message.From.ID)
				if err != nil {
					log.Error().Err(err).Send()
					continue
				}
				if whoomSend == nil {
					log.Error().Err(fmt.Errorf("whoomSend is empty")).Send()
				}

				msg := tgbotapi.NewCopyMessage(*whoomSend, update.Message.Chat.ID, update.Message.MessageID)
				if update.Message.ReplyToMessage != nil {
					replyToID, err := b.db.messagesIDsDB().get(update.Message.ReplyToMessage.MessageID)
					if err != nil && err != badger.ErrKeyNotFound {
						log.Error().Err(err).Send()
						continue
					}

					if replyToID != nil {
						msg.ReplyToMessageID = *replyToID
					}
				}

				rMsg, err := b.bot.Send(msg)
				if err != nil {
					log.Error().Err(err).Send()
					continue
				}

				err = b.db.messagesIDsDB().set(update.Message.MessageID, rMsg.MessageID)
				if err != nil {
					log.Error().Err(err).Send()
					continue
				}

				err = b.db.messagesIDsDB().set(rMsg.MessageID, update.Message.MessageID)
				if err != nil {
					log.Error().Err(err).Send()
					continue
				}

				continue
			}

		}
	}
}
