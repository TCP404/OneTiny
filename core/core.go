package core

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/TCP404/OneTiny-cli/config"
	"github.com/TCP404/OneTiny-cli/internal/middleware"
	"github.com/TCP404/OneTiny-cli/routes"

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
)

var (
	srv = initServer()
	q   = make(chan os.Signal)
)

// RunCore 函数负责启动 gin 实例，开始提供 HTTP 服务
func RunCore() {
	signal.Notify(q, syscall.SIGINT, syscall.SIGTERM)

	go run(srv)
	<-q
	exit(srv)
}

func initServer() *http.Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	middleware.Setup(r)
	routes.Setup(r)
	s := &http.Server{
		Addr:    ":" + strconv.Itoa(config.Port),
		Handler: r,
	}
	return s
}

func run(srv *http.Server) {
	printInfo()
	err := srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Println(color.RedString(err.Error()))
		q <- syscall.SIGINT
	}
}

func exit(srv *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Println(color.RedString(err.Error()))
	}
	select {
	case <-ctx.Done():
		log.Println(ctx.Err())
	default:
		log.Println(color.GreenString("exiting..."))
	}
	os.Exit(0)
}

// printInfo 会在程序启动后打印本机 IP、共享目录、是否允许上传的信息
func printInfo() {
	log.SetOutput(color.Output)
	// Print IP information
	if config.IP != "" {
		log.Printf("Run on   [ %s ]", color.BlueString("http://%s:%d", config.IP, config.Port))
	} else {
		log.Printf("%s", color.YellowString("Warning: [ 暂时获取不到您的IP，可以打开新的命令行窗口输入 ->  ipconfig , 查看您的IP。]"))
	}

	// Print RootPath information
	log.Printf("Run with [ %s ]", color.BlueString("%s", config.RootPath))

	// Print Max allow access level
	log.Printf("Allow access level: [ %s ]", color.BlueString("%d", config.MaxLevel))

	// Print Allow upload Status
	status := color.RedString("%t", config.IsAllowUpload)
	if config.IsAllowUpload {
		status = color.GreenString("%t", config.IsAllowUpload)
	}
	log.Printf("Allow upload: [ %s ]", status)

	// Print Secure status
	status = color.RedString("%t", config.IsSecure)
	if config.IsSecure {
		status = color.GreenString("%t", config.IsSecure)
	}
	log.Printf("Need Login: [ %s ]\n\n", status)
}
