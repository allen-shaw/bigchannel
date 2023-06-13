package interceptor

import "time"

type Interceptor interface {
	BeforeCall()
	AfterCall()
	BeforeSendMsg(m any)
	AfterSendMsg(m any, err error, d time.Duration)
	BeforeRecvMsg()
	AfterRecvMsg()
}
