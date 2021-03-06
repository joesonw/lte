package websocket

import (
	"bufio"
	"context"
	"net"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	lua "github.com/yuin/gopher-lua"

	luacontext "github.com/joesonw/lte/pkg/lua/context"
	libasync "github.com/joesonw/lte/pkg/lua/lib/async"
	libbytes "github.com/joesonw/lte/pkg/lua/lib/bytes"
	libpool "github.com/joesonw/lte/pkg/lua/lib/pool"
	luautil "github.com/joesonw/lte/pkg/lua/util"
	"github.com/joesonw/lte/pkg/stat"
)

const connMetaName = "*WEBSOCKET*CONN*"

type connContext struct {
	addr     string
	messages []wsutil.Message
	conn     net.Conn
	br       *bufio.Reader
	guard    *libpool.Guard
	luaCtx   *luacontext.Context
}

var connFuncs = map[string]lua.LGFunction{
	"read":  connRead,
	"write": connWrite,
	"close": connClose,
}

func connRead(L *lua.LState) int {
	c := L.CheckUserData(1).Value.(*connContext)
	return libasync.DeferredResult(L, c.luaCtx.AsyncPool(), func(ctx context.Context) (lua.LGFunction, error) {
		if c.br != nil {
			messages, err := wsutil.ReadServerMessage(c.br, nil)
			ws.PutReader(c.br)
			c.br = nil
			if err != nil {
				return nil, err
			}

			for i := range messages {
				if messages[i].OpCode.IsData() {
					c.messages = append(c.messages, messages[i])
				}
			}
		}

		if len(c.messages) > 0 {
			m := c.messages[len(c.messages)-1]
			c.messages = c.messages[:len(c.messages)-1]
			luautil.ReportContextStat(c.luaCtx, stat.New("websocket").Tag("addr", c.addr).IntField("read", len(m.Payload)))
			return func(L *lua.LState) int {
				L.Push(libbytes.New(L, m.Payload))
				return 1
			}, nil
		}

		b, err := wsutil.ReadServerText(c.conn)
		if err != nil {
			return nil, err
		}
		luautil.ReportContextStat(c.luaCtx, stat.New("websocket").Tag("addr", c.addr).IntField("read", len(b)))
		return func(L *lua.LState) int {
			L.Push(libbytes.New(L, b))
			return 1
		}, nil
	})
}

func connWrite(L *lua.LState) int {
	c := L.CheckUserData(1).Value.(*connContext)
	bytes := libbytes.Check(L, 2)
	luautil.ReportContextStat(c.luaCtx, stat.New("websocket").Tag("addr", c.addr).IntField("write", len(bytes)))
	return libasync.Deferred(L, c.luaCtx.AsyncPool(), func(ctx context.Context) error {
		return wsutil.WriteClientText(c.conn, bytes)
	})
}

func connClose(L *lua.LState) int {
	c := L.CheckUserData(1).Value.(*connContext)
	return libasync.Deferred(L, c.luaCtx.AsyncPool(), func(ctx context.Context) error {
		if c.br != nil {
			ws.PutReader(c.br)
		}
		c.guard.Done()
		return c.conn.Close()
	})
}
