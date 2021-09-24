package message

// Header types
const (
	HeaderTypeNewClient = iota
	HeaderTypeClientList
	HeaderTypeDisconnectClient
	HeaderTypeClientMessage
)

// Header message prefix
const (
	NewClientHeaderPrefix        = "[new-client]"
	ClientsListHeaderPrefix      = "[clients-list]"
	ClientDisconnectHeaderPrefix = "[client-disconnect]"
	ClientMessageHeaderPrefix    = "[client-message]"
)
