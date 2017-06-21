package stnet

type MsgParse interface {
	//*Session:session which recved message
	//[]byte:recved data now;
	//int:length of recved data parsed;
	ParseMsg(*Session, []byte) int
	OnOpen(*Session)
	OnClose(*Session)
}

//a simple example of "MsgParse interface"
//it's a echo server or client,when it recv a message, it send the message back at once
type SimpleEchoMsgParse struct {
	//you can put any data here,and use Session to access it
	Mydata int
}

func (SimpleEchoMsgParse) ParseMsg(sess *Session, buf []byte) int {
	sess.Send(buf)
	return len(buf)
}

func (SimpleEchoMsgParse) OnOpen(sess *Session) {
}

func (SimpleEchoMsgParse) OnClose(sess *Session) {
}
