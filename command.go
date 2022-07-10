package simple_lsm_db

const (
	ERROR byte = iota
	GET
	SET
	DEL
)

const CommandCode = "Code"

type command interface {
	commandType() byte
}

type errorCommand struct {
	Code byte
}

type getCommand struct {
	Code byte
	Key  string
}

type setCommand struct {
	Code byte
	Key  string
	Val  string
}

type delCommand struct {
	Code byte
	Key  string
	Val  string
}

func (g getCommand) commandType() byte {
	return GET
}

func (s setCommand) commandType() byte {
	return SET
}

func (d delCommand) commandType() byte {
	return DEL
}

func (e errorCommand) commandType() byte {
	return ERROR
}

func NewErrorCommand() *errorCommand {
	return &errorCommand{Code: ERROR}
}

func NewGetCommand(key string) *getCommand {
	return &getCommand{Code: GET, Key: key}
}

func NewSetCommand(key, val string) *setCommand {
	return &setCommand{Code: SET, Key: key, Val: val}
}

func NewDelCommand(key, val string) *delCommand {
	return &delCommand{Code: DEL, Key: key, Val: val}
}
