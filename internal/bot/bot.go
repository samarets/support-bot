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
	db  *db.DB
}

func InitBot(token string, db *db.DB) error {
	botAPI, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return err
	}

	botState := &bot{
		bot: botAPI,
		db:  db,
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
				case "start":
					b.StartCommand(update.Message)
				case "connect":
					b.ConnectCommand(update)
				case "break":
					b.BreakCommand(update)

				}

				continue
			}

			var fromQueue tgbotapi.User
			err := b.db.Get(mergePrefixDB(queue, update.Message.From.ID), &fromQueue)
			if err != nil && err != badger.ErrKeyNotFound {
				log.Error().Err(err).Send()
				continue
			}

			var roomsID int64
			err = b.db.Get(mergePrefixDB(rooms, update.Message.From.ID), &roomsID)
			if err != nil && err != badger.ErrKeyNotFound {
				log.Error().Err(err).Send()
				continue
			}

			defaultUser := tgbotapi.User{}
			if fromQueue == defaultUser && roomsID == 0 {
				err = b.db.Set(mergePrefixDB(queue, update.Message.From.ID), update.Message.From)
				if err != nil {
					log.Error().Err(err).Send()
					continue
				}

				var bufferMessages []tgbotapi.Message
				bufferMessages = append(bufferMessages, *update.Message)

				err = b.db.Set(mergePrefixDB(buffer, update.Message.From.ID), bufferMessages)
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

				fmt.Println(update.Message.Chat.ID, "create request")
				continue
			}
			if fromQueue != defaultUser && roomsID == 0 {
				var bufferMessages []tgbotapi.Message
				err = b.db.Get(mergePrefixDB(buffer, update.Message.From.ID), &bufferMessages)
				if err != nil && err != badger.ErrKeyNotFound {
					log.Error().Err(err).Send()
					continue
				}

				bufferMessages = append(bufferMessages, *update.Message)

				err = b.db.Set(mergePrefixDB(buffer, update.Message.From.ID), bufferMessages)
				if err != nil {
					log.Error().Err(err).Send()
					continue
				}

				continue
			}
			if roomsID != 0 {
				var whoomSend int64
				err = b.db.Get(mergePrefixDB(rooms, update.Message.From.ID), &whoomSend)
				if err != nil && err != badger.ErrKeyNotFound {
					log.Error().Err(err).Send()
					continue
				}

				msg := tgbotapi.NewCopyMessage(whoomSend, update.Message.Chat.ID, update.Message.MessageID)
				if update.Message.ReplyToMessage != nil {
					var replyToID int
					err = b.db.Get(mergePrefixDB(messagesIDs, update.Message.ReplyToMessage.MessageID), &replyToID)
					if err != nil && err != badger.ErrKeyNotFound {
						log.Error().Err(err).Send()
						continue
					}

					if replyToID != 0 {
						fmt.Println(replyToID)
						msg.ReplyToMessageID = replyToID
					}
				}

				rMsg, err := b.bot.Send(msg)
				if err != nil {
					log.Error().Err(err).Send()
					continue
				}

				err = b.db.Set(mergePrefixDB(messagesIDs, update.Message.MessageID), rMsg.MessageID)
				if err != nil {
					log.Error().Err(err).Send()
					continue
				}

				err = b.db.Set(mergePrefixDB(messagesIDs, rMsg.MessageID), update.Message.MessageID)
				if err != nil {
					log.Error().Err(err).Send()
					continue
				}

				fmt.Println(update.Message.Chat.ID, "send msg")
				continue
			}

		}
	}
}

func mergePrefixDB(prefix string, id interface{}) []byte {
	return []byte(fmt.Sprintf("%s-%d", prefix, id))
}
