package transport

import (
	"github.com/KalbiProject/Kalbi/sip/message"
	"github.com/KalbiProject/Kalbi/interfaces"
)

type ListeningPoint interface {
	Read() SipEventObject
	Build(string, int)
	Start()
	SetTransportChannel(chan interfaces.SipEventObject)
	Send(string, string, string) error
}

type SipEventObject interface{
	GetSipMessage() *message.SipMsg
	SetSipMessage(*message.SipMsg)
	GetTransaction() interfaces.Transaction
	SetTransaction(interfaces.Transaction)
}


type Transaction interface {
	GetBranchID() string
	GetOrigin() *message.SipMsg
	SetListeningPoint(ListeningPoint)
	Send(*message.SipMsg, string, string)
	Receive(*message.SipMsg)
	GetLastMessage() *message.SipMsg
	GetServerTransactionID() string
    SetLastMessage(*message.SipMsg)
    GetListeningPoint() ListeningPoint
}