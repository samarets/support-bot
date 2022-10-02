package messages

type Code string

const (
	HelloMessage      Code = "hello"
	OperatorConnected Code = "op-conn"
)

var LocalMessages = map[Code]map[Locale]string{
	HelloMessage: {
		LocaleUK: "Привіт, напиши своє питання - ми допоможемо",
		LocaleEN: "Hello, write your question and we will help",
	},
	OperatorConnected: {
		LocaleUK: "До вас доєднався оператор",
		LocaleEN: "An operator has joined you",
	},
}
