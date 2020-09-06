package redis

import (
	"fmt"
	"testing"
)

func TestRedisPool(t *testing.T) {

	conn:=RedisPool().Get()
	defer conn.Close()
	data,_ := conn.Do("hgetall",1)
	datal := data.([]interface{})
	for i := 0; i<len(datal);i++  {
		v := string(datal[i].([]byte))
		fmt.Println(v)
	}
}
