package bot

type state string

const (
	defaultState state = "default-state"
	unknownState state = "unknown-state"
	queueState   state = "queue-state"
	roomState    state = "room-state"
)

const (
	botWasBlockedError string = "Forbidden: bot was blocked by the user"
)
