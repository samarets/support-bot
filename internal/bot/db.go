package bot

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

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
	group       = "group"
	support     = "support"
)

type botDB struct {
	db *db.DB
}

func newBotDB(db *db.DB) *botDB {
	return &botDB{
		db: db,
	}
}

func (db *botDB) getUserState(userID int64) (state, error) {
	fromQueue, err := db.queueDB().get(userID)
	if err != nil {
		return unknownState, err
	}
	if fromQueue != nil {
		return queueState, nil
	}

	fromRoom, err := db.roomsDB().get(userID)
	if err != nil {
		log.Error().Err(err).Send()
		return unknownState, err
	}
	if fromRoom != nil {
		return roomState, nil
	}

	return defaultState, nil
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

type groupDB struct {
	*botDB
}

func (db *botDB) groupDB() *groupDB {
	return &groupDB{
		botDB: db,
	}
}

func (db *groupDB) set(groupID int64) error {
	return db.db.Set([]byte(group), groupID)
}

func (db *groupDB) get() int64 {
	var groupID int64
	err := db.db.Get([]byte(group), &groupID)
	if err != nil {
		switch err {
		case badger.ErrKeyNotFound:
			return 0
		default:
			log.Error().Err(err).Send()
			return 0
		}
	}

	return groupID
}

type supportDB struct {
	*botDB
}

func (db *botDB) supportDB() *supportDB {
	return &supportDB{
		botDB: db,
	}
}

func (db *supportDB) set(supportID int64, isSupport bool) error {
	return db.db.Set(mergePrefixDB(support, supportID), isSupport)
}

func (db *supportDB) get(userID int64) bool {
	var isSupport bool
	err := db.db.Get(mergePrefixDB(support, userID), &isSupport)
	if err != nil {
		switch err {
		case badger.ErrKeyNotFound:
			return false
		default:
			log.Error().Err(err).Send()
			return false
		}
	}

	return isSupport
}

func (db *supportDB) getAll() ([]int64, error) {
	m, err := db.db.GetAll([]byte(support))
	if err != nil {
		return nil, err
	}

	supports := make([]int64, 0, len(m))
	for key, value := range m {
		var isSupport bool
		err = json.Unmarshal(value, &isSupport)
		if err != nil {
			return nil, err
		}
		if !isSupport {
			continue
		}

		key = trimKeyPrefix(key, support)
		keyInt, err := strconv.ParseInt(key, 10, 64)
		if err != nil {
			return nil, err
		}

		supports = append(supports, keyInt)
	}

	return supports, nil
}

func mergePrefixDB(prefix string, id interface{}) []byte {
	return []byte(fmt.Sprintf("%s-%d", prefix, id))
}

func trimKeyPrefix(key, prefix string) string {
	return strings.TrimPrefix(key, prefix+"-")
}
