package transaction

import (
	//"fmt"
	"github.com/KalbiProject/Kalbi/log"
	"github.com/KalbiProject/Kalbi/sip/message"
	"github.com/KalbiProject/Kalbi/sip/method"
	"github.com/KalbiProject/Kalbi/transport"
	"github.com/looplab/fsm"
)

const (
	serverInputRequest      = "server_input_request"
	serverInputAck          = "server_input_ack"
	serverInputUser1xx      = "server_input_user_1xx"
	serverInputUser2xx      = "server_input_user_2xx"
	serverInputUser300Plus  = "server_input_user_300_plus"
	serverInputTimerG       = "server_input_timer_g"
	serverInputTimerH       = "server_input_timer_h"
	serverInputTimerI       = "server_input_timer_i"
	serverInputTransportErr = "server_input_transport_err"
	serverInputDelete       = "server_input_delete"
)

type ServerTransaction struct {
	ID             string
	BranchID       string
	TransManager   *TransactionManager
	Origin         *message.SipMsg
	FSM            *fsm.FSM
	msgHistory     []*message.SipMsg
	ListeningPoint transport.ListeningPoint
	Host           string
	Port           string
	LastMessage    *message.SipMsg
}

func (st *ServerTransaction) InitFSM(msg *message.SipMsg) {

	switch string(msg.Req.Method) {
	case method.INVITE:
		st.FSM = fsm.NewFSM("", fsm.Events{
			{Name: serverInputRequest, Src: []string{""}, Dst: "Proceeding"},
			{Name: serverInputUser1xx, Src: []string{"Proceeding"}, Dst: "Proceeding"},
			{Name: serverInputUser300Plus, Src: []string{"Proceeding"}, Dst: "Completed"},
			{Name: serverInputAck, Src: []string{"Completed"}, Dst: "Confirmed"},
			{Name: serverInputUser2xx, Src: []string{"Proceeding"}, Dst: "Terminated"},
		}, fsm.Callbacks{serverInputUser1xx: st.actRespond,
			serverInputTransportErr: st.actTransErr,
			serverInputUser2xx:      st.actRespondDelete,
			serverInputUser300Plus:  st.actRespond})
	default:
		st.FSM = fsm.NewFSM("", fsm.Events{
			{Name: serverInputRequest, Src: []string{""}, Dst: "Proceeding"},
			{Name: serverInputUser1xx, Src: []string{"Proceeding"}, Dst: "Proceeding"},
			{Name: serverInputUser300Plus, Src: []string{"Proceeding"}, Dst: "Completed"},
			{Name: serverInputAck, Src: []string{"Completed"}, Dst: "Confirmed"},
			{Name: serverInputUser2xx, Src: []string{"Proceeding"}, Dst: "Terminated"},
		}, fsm.Callbacks{serverInputUser1xx: st.actRespond,
			serverInputTransportErr: st.actTransErr,
			serverInputUser2xx:      st.actRespondDelete,
			serverInputUser300Plus:  st.actRespond})
	}
}

func (st *ServerTransaction) SetListeningPoint(lp transport.ListeningPoint) {
	st.ListeningPoint = lp
}

func (st *ServerTransaction) GetBranchId() string {
	return st.BranchID
}

func (st *ServerTransaction) GetOrigin() *message.SipMsg {
	return st.Origin
}

func (st *ServerTransaction) Receive(msg *message.SipMsg) {
	st.LastMessage = msg
	log.Log.Info("Message Received for transactionId " + st.BranchID + ": \n" + string(msg.Src))
	log.Log.Info(message.MessageDetails(msg))
	if msg.Req.Method != nil || string(msg.Req.Method) != method.ACK {
		st.FSM.Event(serverInputRequest)
	}

}

func (st *ServerTransaction) Respond(msg *message.SipMsg) {
	//TODO: this will change due to issue https://github.com/KalbiProject/Kalbi/issues/20
	log.Log.Info("Message Sent for transactionId " + st.BranchID + ": \n" + message.MessageDetails(msg))
	if msg.GetStatusCode() < 200 {
		st.FSM.Event(serverInputUser1xx)
	} else if msg.GetStatusCode() < 300 {
		st.FSM.Event(serverInputUser2xx)
	} else {
		st.FSM.Event(serverInputUser300Plus)
	}

}

func (st *ServerTransaction) Send(msg *message.SipMsg, host string, port string) {
	st.LastMessage = msg
	st.Host = host
	st.Port = port
	st.Respond(msg)
}

func (st *ServerTransaction) actRespond(event *fsm.Event) {
	err := st.ListeningPoint.Send(st.Host, st.Port, st.LastMessage.Export())
	if err != nil {
		st.FSM.Event(serverInputTransportErr)
	}

}

func (st *ServerTransaction) actRespondDelete(event *fsm.Event) {
	err := st.ListeningPoint.Send(st.Host, st.Port, st.LastMessage.Export())
	if err != nil {
		st.FSM.Event(serverInputTransportErr)
	}
	st.TransManager.DeleteTransaction(st.BranchID)
}

func (st *ServerTransaction) actTransErr(event *fsm.Event) {
	log.Log.Error("Transport error for transactionID : " + st.BranchID)
	st.FSM.Event(serverInputDelete)
}

func (st *ServerTransaction) actDelete(event *fsm.Event) {
	st.TransManager.DeleteTransaction(st.BranchID)
}
