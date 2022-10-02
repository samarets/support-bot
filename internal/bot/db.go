package bot

import (
	"fmt"

	"github.com/dgraph-io/badger/v3"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/samarets/support-bot/internal/log"

	"github.com/samarets/support-bot/internal/db"
)

const (
	queue       = "queue"
	rooms       = "rooms"
	buffer      = "buffer"
	messagesIDs = "messagesIDs"
	language    = "language"
)

type botDB struct {
	db *db.DB
}

func newBotDB(db *db.DB) *botDB {
	return &botDB{
		db: db,
	}
}

type queueDB struct {
	*botDB
}

func (db *botDB) queueDB() *queueDB {
	return &queueDB{
		botDB: db,
	}
}

func (db *queueDB) set(userID int64, user *tgbotapi.User) error {
	return db.db.Set(mergePrefixDB(queue, userID), user)
}

func (db *queueDB) get(userID int64) (*tgbotapi.User, error) {
	var fromQueue tgbotapi.User
	err := db.db.Get(mergePrefixDB(queue, userID), &fromQueue)
	if err != nil {
		switch err {
		case badger.ErrKeyNotFound:
			return nil, nil
		default:
			return nil, err
		}
	}

	return &fromQueue, nil
}

func (db *queueDB) getFirst() (*tgbotapi.User, error) {
	var fromQueue tgbotapi.User
	err := db.db.GetFirstWherePrefix([]byte(queue), &fromQueue)
	if err != nil {
		switch err {
		case badger.ErrKeyNotFound:
			return nil, nil
		default:
			return nil, err
		}
	}

	return &fromQueue, nil
}

func (db *queueDB) delete(userID int64) error {
	return db.db.Drop(mergePrefixDB(queue, userID))
}

type roomsDB struct {
	*botDB
}

func (db *botDB) roomsDB() *roomsDB {
	return &roomsDB{
		botDB: db,
	}
}

func (db *roomsDB) set(userID, secondUserID int64) error {
	return db.db.Set(mergePrefixDB(rooms, userID), secondUserID)
}

func (db *roomsDB) get(userID int64) (*int64, error) {
	var secondUserID int64
	err := db.db.Get(mergePrefixDB(rooms, userID), &secondUserID)
	if err != nil {
		switch err {
		case badger.ErrKeyNotFound:
			return nil, nil
		default:
			return nil, err
		}
	}

	return &secondUserID, nil
}

func (db *roomsDB) delete(userID int64) error {
	return db.db.Drop(mergePrefixDB(rooms, userID))
}

type bufferDB struct {
	*botDB
}

func (db *botDB) bufferDB() *bufferDB {
	return &bufferDB{
		botDB: db,
	}
}

func (db *bufferDB) set(userID int64, messages []tgbotapi.Message) error {
	return db.db.Set(mergePrefixDB(buffer, userID), messages)
}

func (db *bufferDB) get(userID int64) ([]tgbotapi.Message, error) {
	var bufferMessages []tgbotapi.Message
	err := db.db.Get(mergePrefixDB(buffer, userID), &bufferMessages)
	if err != nil {
		switch err {
		case badger.ErrKeyNotFound:
			return nil, nil
		default:
			return nil, err
		}
	}

	return bufferMessages, nil
}

func (db *bufferDB) delete(userID int64) error {
	return db.db.Drop(mergePrefixDB(buffer, userID))
}

type messagesIDsDB struct {
	*botDB
}

func (db *botDB) messagesIDsDB() *messagesIDsDB {
	return &messagesIDsDB{
		botDB: db,
	}
}

func (db *messagesIDsDB) set(messageID, secondMessageID int) error {
	return db.db.Set(mergePrefixDB(messagesIDs, messageID), secondMessageID)
}

func (db *messagesIDsDB) get(messageID int) (*int, error) {
	var replyToID int
	err := db.db.Get(mergePrefixDB(messagesIDs, messageID), &replyToID)
	if err != nil {
		switch err {
		case badger.ErrKeyNotFound:
			return nil, nil
		default:
			return nil, err
		}
	}

	return &replyToID, nil
}

type languageDB struct {
	*botDB
}

func (db *botDB) languageDB() *languageDB {
	return &languageDB{
		botDB: db,
	}
}

func (db *languageDB) set(userID int64, lang string) error {
	return db.db.Set(mergePrefixDB(language, userID), lang)
}

func (db *languageDB) get(userID int64) string {
	var lang string
	err := db.db.Get(mergePrefixDB(language, userID), &lang)
	if err != nil {
		switch err {
		case badger.ErrKeyNotFound:
			return ""
		default:
			log.Error().Err(err).Send()
			return ""
		}
	}

	return lang
}

func mergePrefixDB(prefix string, id interface{}) []byte {
	return []byte(fmt.Sprintf("%s-%d", prefix, id))
}
