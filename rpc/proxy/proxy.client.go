package proxy

import (
	"context"
	"google.golang.org/grpc"
	"l2-push-go/util"
	"time"
)

type Client struct {
	ProxyClient
	conn *grpc.ClientConn //仅保存内部创建的conn
}

func MustNewClient(server string,opts ...grpc.DialOption)*Client{
	ctx,cancel:=context.WithTimeout(context.Background(),3*time.Second)
	defer cancel()

	opts=append(opts,grpc.WithBlock(),grpc.WithInsecure())

	conn,err:=grpc.DialContext(ctx,server,opts...)
	util.AssertNilErr(err,`GRPC客户端创建出错[server=%v]`,server)

	return &Client{ProxyClient:NewProxyClient(conn),conn:conn}
}

func (c *Client)Close()error{
	if conn := c.conn; conn != nil {
		c.conn = nil
		return conn.Close()
	}

	return nil
}