package socket_io_emitter

import (
	"fmt"

	"github.com/shamaton/msgpack"
)

const uid = "emitter"
const event = 2

// EmitterRedis 广播相关的redis方法接口
type EmitterRedis interface {
	Publish(channel string, data []byte) error
}

type flags struct {
	json      bool
	volatile  bool
	broadcast bool
}

//Emitter 广播实例定义
type Emitter struct {
	Prefix  string
	Nsp     string
	channel string
	_rooms  []string
	_flags  flags
	_except []string
	redis   EmitterRedis
}

// NewEmitter 初始化Emitter实例
func NewEmitter(redis EmitterRedis, prefix string) *Emitter {
	if prefix == "" {
		prefix = "socket.io" // default socket.io
	}
	emitter := &Emitter{
		Prefix: prefix,
		Nsp:    "/",
		_flags: flags{true, true, true},
		redis:  redis,
	}
	return emitter
}

//Of 指定命名空间
func (emitter *Emitter) Of(nsp string) *Emitter {
	emitter.Nsp = nsp
	emitter.channel = emitter.Prefix + "#" + emitter.Nsp + "#"
	return emitter
}

// Except 排除定向的socketid
func (emitter *Emitter) Except(sid []string) *Emitter {
	emitter._except = sid
	return emitter
}

//To 指定单个房间或者socketid
func (emitter *Emitter) To(room string) *Emitter {
	emitter._rooms = append(emitter._rooms, room)
	return emitter
}

//ToRooms 指定多个房间
func (emitter *Emitter) ToRooms(rooms []string) *Emitter {
	emitter._rooms = append(emitter._rooms, rooms...)
	return emitter
}

//Emit 发送广播
func (emitter *Emitter) Emit(args ...interface{}) error {
	packet := map[string]interface{}{"type": event, "data": args, "nsp": emitter.Nsp}
	opts := map[string]interface{}{"rooms": emitter._rooms, "flags": emitter._flags}
	if len(emitter._except) > 0 {
		opts["except"] = emitter._except
	}
	fmt.Println(uid, packet, opts)
	msg, err := msgpack.Encode([]interface{}{uid, packet, opts})
	if err != nil {
		return err
	}
	channel := emitter.channel + emitter._rooms[0] + "#"
	fmt.Println(channel)
	err = emitter.redis.Publish(channel, msg)
	return err
}
