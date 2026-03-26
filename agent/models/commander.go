package models

// CommandReceiver is a callback function that handles received commands.
// It receives the command name, associated data, and bus message headers.
type CommandReceiver func(data any, headers BusHeaders)

type Commander interface {
	Send(command string, data any, headers BusHeaders) error
	SendToHive(command, hivePath string, data any, headers BusHeaders) error
	OnCommandReceived(command string, receiver CommandReceiver)
	Emit(command string, data any, headers BusHeaders, interval, duration int)
	EmitToHive(command, hivePath string, data any, headers BusHeaders, interval, duration int)
	CancelEmit(command string)
}
