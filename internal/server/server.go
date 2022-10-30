package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/TCP404/OneTiny-cli/internal/conf"
	"github.com/TCP404/OneTiny-cli/internal/server/middleware"
	"github.com/TCP404/OneTiny-cli/internal/server/routes"
	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
)

// RunCore 函数负责启动 gin 实例，开始提供 HTTP 服务
func RunCore() {
	var (
		srv = initServer()
		q   = make(chan os.Signal, 1)
	)
	signal.Notify(q, syscall.SIGINT, syscall.SIGTERM)

	{
		go run(srv)
		<-q
		exit(srv)
	}
}

func initServer() *http.Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	middleware.Setup(r)
	routes.Setup(r)
	s := &http.Server{
		Addr:    ":" + strconv.Itoa(conf.Config.Port),
		Handler: r,
	}
	return s
}

func run(srv *http.Server) {
	printInfo()
	err := srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Println(color.RedString(err.Error()))
	}
}

func exit(srv *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Println(color.RedString(err.Error()))
	}
	fmt.Println(color.GreenString("\nbye~"))
	os.Exit(0)
}

// printInfo 会在程序启动后打印本机 IP、共享目录、是否允许上传的信息
func printInfo() {
	log.SetOutput(color.Output)
	// Print IP information
	if conf.Config.IP != "" {
		log.Printf("Run on   [ %s ]", color.BlueString("http://%s:%d", conf.Config.IP, conf.Config.Port))
	} else {
		log.Printf("%s", color.YellowString("Warning: [ 暂时获取不到您的IP，可以打开新的命令行窗口输入 ->  ipconfig , 查看您的IP。]"))
	}

	// Print RootPath information
	log.Printf("Run with [ %s ]", color.BlueString("%s", conf.Config.RootPath))

	// Print Max allow access level
	log.Printf("Allow access level: [ %s ]", color.BlueString("%d", conf.Config.MaxLevel))

	// Print Allow upload Status
	status := color.RedString("%t", conf.Config.IsAllowUpload)
	if conf.Config.IsAllowUpload {
		status = color.GreenString("%t", conf.Config.IsAllowUpload)
	}
	log.Printf("Allow upload: [ %s ]", status)

	// Print Secure status
	status = color.RedString("%t", conf.Config.IsSecure)
	if conf.Config.IsSecure {
		status = color.GreenString("%t", conf.Config.IsSecure)
	}
	log.Printf("Need Login: [ %s ]\n\n", status)
}
