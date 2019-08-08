package example

import (
	"errors"
	"fmt"

	"github.com/gomodule/redigo/redis"
	socketIOEmitter "socket.io-emitter"
)

type emitterClient struct {
}

var eC = &emitterClient{}

var (
	redisPool *redis.Pool
	ErrRedis  = errors.New("get redis conn error")
)

func init() {
	initRedis()
}

func initRedis() {
	// 建立连接池
	redisPool = &redis.Pool{
		// 从配置文件获取maxidle以及maxactive，取不到则用后面的默认值
		MaxIdle:     100,
		MaxActive:   800,
		IdleTimeout: 60 * 1000,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", "redis host")
			if err != nil {
				return nil, err
			}
			if _, err := c.Do("AUTH", "redis password"); err != nil {
				c.Close()
				return nil, err
			}
			// 选择db
			if _, err := c.Do("SELECT", 0); err != nil {
				c.Close()
				return nil, err
			}
			return c, nil
		},
	}
}

// RedisControlConn redis操作
func RedisControlConn(f func(redis.Conn) error) error {
	conn := redisPool.Get()
	err := conn.Err()
	if err != nil {
		fmt.Println(fmt.Errorf("redis error: %v", err))
		return ErrRedis
	}
	defer conn.Close()
	return f(conn)
}

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
