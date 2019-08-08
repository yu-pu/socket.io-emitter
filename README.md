# socket.io-emitter
socket.io-emitter

### example

```
// PublishData 实现publish方法，发布通知
func (client *emitterClient) Publish(channel string, data []byte) error {
	return RedisControlConn(func(conn redis.Conn) error {
		_, err := conn.Do("PUBLISH", channel, data)
		return err
	})
}

// BroadcastRoom 广播到房间内全部连接
func BroadcastRoom(prefixName string, nsp string, room string, event string, data interface{}) error {
	return socketIOEmitter.NewEmitter(eC, prefixName).Of(nsp).To(room).Emit(event, data)
}

// EmitToSocketID 单播
func EmitToSocketID(prefixName string, nsp string, socketID string, event string, data interface{}) error {
	return socketIOEmitter.NewEmitter(eC, prefixName).Of(nsp).To(socketID).Emit(event, data)
}

// BroadcastRoomExcept 广播到房间内，排除部分连接
func BroadcastRoomExcept(prefixName string, nsp string, room string, except []string, event string, data interface{}) error {
	return socketIOEmitter.NewEmitter(eC, prefixName).Except(except).Of(nsp).To(room).Emit(event, data)
}

```
