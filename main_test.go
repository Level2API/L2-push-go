package main

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"l2-push-go/rpc/entity"
	"l2-push-go/rpc/proxy"
	"testing"
)

func mustNewProxyClient() *proxy.Client {
	server := `localhost:8090` //推送代理服务器地址
	return proxy.MustNewClient(server)
}

// 测试行情推送
func TestTxMsgPush(t *testing.T) {
	r := require.New(t)
	ctx := context.TODO()

	client := mustNewProxyClient()
	defer client.Close()

	//创建逐笔成交推送流，可调用其他接口创建不同类型的推送流
	tickStream, err := client.NewTickRecordStream(ctx, &entity.Void{})
	r.NoError(err)

	//读取推送消息
	for {
		tick, err := tickStream.Recv()
		if err != nil {
			if err == io.EOF {
				fmt.Println(`服务端已关闭推送流`)
				return
			}

			if status.Code(err) == codes.Unavailable {
				fmt.Println(`连接已端开:`, err)
				return
			}

			fmt.Println(`推送消息读取出错:`, err)
			return
		}

		//全订阅仅逐笔成交每秒大概推送6000条
		fmt.Println(tick)
	}
}

// 测试用户订阅
func TestSubscription(t *testing.T) {
	r := require.New(t)
	ctx := context.TODO()

	client := mustNewProxyClient()
	defer client.Close()

	printlnCurrentSub := func(tag string) {
		//查询用户当前订阅
		rsp, err := client.GetSubscription(ctx, &entity.Void{})
		r.NoError(err)
		r.True(rsp.Rsp.Code == 1)  //成功响应
		fmt.Println(tag, rsp.Data) //当前订阅信息
	}
	printlnCurrentSub(`t1`)

	//新增订阅
	rsp2, err := client.AddSubscription(ctx, &entity.String{Value: `2_000002_1,1_600276_1`})
	r.NoError(err)
	r.True(rsp2.Code == 1)
	printlnCurrentSub(`t2`)

	//删除订阅
	rsp3, err := client.DelSubscription(ctx, &entity.String{Value: `2_000002_1`})
	r.NoError(err)
	r.True(rsp3.Code == 1)
	printlnCurrentSub(`t3`)
}

// 测试行情推送-并发处理
func TestTxMsgPushParallel(t *testing.T) {
	r := require.New(t)
	ctx := context.TODO()

	client := mustNewProxyClient()
	defer client.Close()

	//stream.Recv()非线程安全，以下启用1个协程读取消息并写入channel
	//其他协程读取channel获取消息并处理，1个消息仅发送给其中1个协程
	newTickStream := func() <-chan *entity.TickRecord {
		ch := make(chan *entity.TickRecord)

		//启动单协程处理推送消息
		go func() {
			defer close(ch)

			tickStream, err := client.NewTickRecordStream(ctx, &entity.Void{})
			r.NoError(err)

			for {
				tick, err := tickStream.Recv()
				if err != nil {
					fmt.Println(`连接已断开：`, err)
				}

				ch <- tick
			}
		}()

		return ch
	}

	tickStream := newTickStream()

	//启动多协程从管道获取消息并处理
	for i := 0; i < 10; i++ {
		go func() {
			for tick := range tickStream {
				fmt.Println(tick)
			}
		}()
	}

	select {}

}
