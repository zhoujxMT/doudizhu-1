package rpcs

import (
	"context"
	"net"
	"reflect"

	"github.com/smallnest/rpcx/server"

	eng "kelei.com/doudizhu/engine"
	"kelei.com/utils/frame"
	"kelei.com/utils/logger"
	"kelei.com/utils/rpcs"
)

var (
	engine *eng.Engine
)

func Inject(engine_ *eng.Engine) {
	engine = engine_
}

type DoudizhuS struct {
}

func (this *DoudizhuS) Call(ctx context.Context, args *rpcs.Args, reply *rpcs.Reply) error {
	clientConn := ctx.Value(server.RemoteConnContextKey).(net.Conn)
	args.V["clientConn"] = clientConn
	v := reflect.ValueOf(engine)
	funcName := args.V["funcname"].(string)
	logger.Debugf("收到请求:%s%v", funcName, args.V["args"])
	mv := v.MethodByName(funcName)
	res := mv.Call([]reflect.Value{reflect.ValueOf(args)})
	r := res[0].Interface().(*rpcs.Reply)
	reply.RS = r.RS
	reply.SC = r.SC
	if reply.RS == nil {
		logger.Debugf("   回发结果:%s,%v", reply.SC, reply.RS)
	} else {
		logger.Debugf("   回发结果:%s,%s", reply.SC, *reply.RS)
	}
	return nil
}

//给所有的订阅者推送消息
func PushMessage(userids, funcName, messages string) {
	xclient := frame.NewRpcxClient(frame.GetArgs().RpcxClient)
	defer xclient.Close()
	args := &rpcs.Args{}
	p := []string{userids, funcName, messages}
	logger.Debugf("推送数据:%s", p)
	args.V = map[string]interface{}{"funcname": "PushMessage", "args": p}
	reply := &rpcs.Reply{}
	err := xclient.Fork(context.Background(), "Call", args, reply)
	logger.CheckError(err)
}
