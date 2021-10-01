package util

import (
	"fmt"
	"net"
	"sync"
)

var connPool *sync.Pool

func InitConnPool(network, address string) {
	connPool = &sync.Pool{}
	connPool.New = func() interface{} {
		conn, err := net.Dial(network, address)
		if err != nil {
			fmt.Printf("net dial err: %s\n", err)
			return nil
		}
		return conn
	}
}

func GetConn() net.Conn {
	connInterface := connPool.Get()
	if connInterface == nil {
		fmt.Printf("get conn failed.")
		return nil
	}
	conn, ok := connInterface.(net.Conn)
	if !ok {
		fmt.Printf("get conn failed.")
		return nil
	}
	return conn
}

func ReleaseConn(conn net.Conn) {
	connPool.Put(conn)
}
