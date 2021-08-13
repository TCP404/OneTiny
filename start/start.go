package start

import (
	"fmt"
	"io"
	"log"
	"net"
	"oneTiny/config"
	"oneTiny/controller"
	"oneTiny/middleware"
	"os"
	"strconv"

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	"github.com/urfave/cli/v2"
)

func initCLI() {
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"v"},
		Usage:   "打印版本信息",
	}
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Println("当前版本: ", c.App.Version)
		os.Exit(0)
	}
	cli.HelpFlag = &cli.BoolFlag{
		Name:    "help",
		Aliases: []string{"h"},
		Usage:   "打印帮助信息",
	}
	cli.HelpPrinter = func(w io.Writer, templ string, data interface{}) {
		cli.HelpPrinterCustom(w, templ, data, nil)

		os.Exit(0)
	}
}

// Run 函数作为程序入口，主要负责处理命令和 flag
func Run() {
	initCLI()
	globalFlag := []cli.Flag{
		&cli.PathFlag{
			Name:        "road",
			Aliases:     []string{"r"},
			Usage:       "指定对外开放的目录`路径`",
			Value:       config.RootPath,
			Required:    false,
			DefaultText: config.RootPath,
		},
		&cli.IntFlag{
			Name:        "port",
			Aliases:     []string{"p"},
			Usage:       "指定开放的`端口`",
			Value:       config.Port,
			Required:    false,
			DefaultText: strconv.Itoa(config.Port),
		},
		&cli.BoolFlag{
			Name:        "allow",
			Aliases:     []string{"a"},
			Usage:       "指定`是否`允许访问者上传",
			Value:       config.IsAllowUpload,
			Required:    false,
			DefaultText: "false",
		},
		&cli.IntFlag{
			Name:        "max",
			Aliases:     []string{"x"},
			Usage:       "指定允许访问的`深度`，默认仅限访问共享目录",
			Value:       int(config.MaxLevel),
			Required:    false,
			DefaultText: "0",
		},
		&cli.StringFlag{
			Name:  "user",
			Usage: "指定访问`帐号`",
			Value: "admin",
		},
		&cli.StringFlag{
			Name:  "pass",
			Usage: "指定访问`密码`",
			Value: "admin",
		},
		&cli.BoolFlag{
			Name:     "secure",
			Aliases:  []string{"s"},
			Usage:    "指定是否开启访问登录",
			Value:    false,
			Required: false,
		},
	}

	updateCmd := &cli.Command{
		Name:    "update",
		Aliases: []string{"u", "up"},
		Usage:   "更新 OneTiny 到最新版",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "list",
				Aliases:     []string{"l"},
				Usage:       "列出远程服务器上所有可用版本",
				Required:    false,
				DefaultText: "false",
			},
		},
	}

	configCmd := &cli.Command{
		Name:        "config",
		Aliases:     []string{"c", "cf", "cfg", "conf"},
		Usage:       "设置默认配置",
		UsageText:   "onetiny config [OPTIONS]",
		Flags:       globalFlag,
		Description: "使用 onetiny config 命令可以将设置写入配置文件。\n使用方式与 onetiny 命令相同，仅多了一个 config 关键字，如：\n  onetiny config -p 10240  可以将端口设置为 10240 写入配置\n  onetiny config -a false  可以设置不允许访问者上传并写入配置",
		Action: func(c *cli.Context) error {
			if err := config.Set(c); err != nil {
				return cli.Exit(err.Error(), 1)
			}
			return cli.Exit("配置成功~", 0)
		},
		HelpName: "help",
	}

	app := &cli.App{
		Name:            "OneTiny",
		Usage:           "一个用于局域网内共享文件的FTP程序",
		UsageText:       "onetiny [GLOBAL OPTIONS] COMMAND [COMMAND OPTIONS] [参数...]",
		Version:         config.VERSION,
		Flags:           globalFlag,
		Authors:         []*cli.Author{{Name: "Boii", Email: "i@tcp404.com"}},
		Commands:        []*cli.Command{updateCmd, configCmd},
		CommandNotFound: func(c *cli.Context, s string) { cli.ShowAppHelpAndExit(c, 10) },
		Action: func(c *cli.Context) error {
			config.Port = c.Int("port")
			config.RootPath = c.String("road")
			config.MaxLevel = uint8(c.Int("max"))
			config.IsAllowUpload = c.Bool("allow")
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

// StartGin 函数负责启动 gin 实例，开始提供 HTTP 服务
func StartGin() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.LoggerWithWriter(config.Output), gin.Recovery())
	r.Use(middleware.InterceptICO)
	r.Use(middleware.CheckLevel)

	// 注册路由
	r.NoRoute(controller.NotFound)
	r.GET("/*filename", controller.Handler)
	r.POST("/upload", controller.Upload)

	printInfo()

	err := r.Run(":" + strconv.Itoa(config.Port))
	if _, ok := err.(*net.OpError); ok {
		log.Fatal(color.RedString("指定的 %d 端口已被占用，请换一个端口号。", config.Port))
	}
}

// printInfo 会在程序启动后打印本机 IP、共享目录、是否允许上传的信息
func printInfo() {
	log.SetOutput(color.Output)
	// Print IP infomation
	if config.IP != "" {
		log.Printf("Run on   [ %s ]", color.BlueString("http://%s:%d", config.IP, config.Port))
	} else {
		log.Printf("%s", color.YellowString("Warning: [ 暂时获取不到您的IP，可以打开新的命令行窗口输入 ->  ipconfig , 查看您的IP。]"))
	}

	// Print RootPath infomation
	log.Printf("Run with [ %s ]", color.BlueString("%s", config.RootPath))

	// Print Max allow access level
	log.Printf("Allow access level: [ %s ]", color.BlueString("%d", config.MaxLevel))

	// Print Allow upload Status
	status := color.RedString("%t", config.IsAllowUpload)
	if config.IsAllowUpload {
		status = color.GreenString("%t", config.IsAllowUpload)
	}
	log.Printf("Allow upload : [ %s ]", status)
}
