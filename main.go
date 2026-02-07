package main

import (
	"embed"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/afumu/wetrace/store"
	"github.com/afumu/wetrace/web"
	"github.com/spf13/viper"
)

//go:embed ui/dist
var uiDist embed.FS

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// --- 加载配置 ---
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 文件不存在，尝试创建默认配置
			if err := viper.SafeWriteConfig(); err != nil {
				log.Printf("无法创建默认 .env 文件: %v", err)
			} else {
				log.Println("已自动创建并初始化 .env 配置文件")
			}
		} else {
			log.Printf("注意: 读取 .env 文件出错: %v. 将使用默认值或环境变量。", err)
		}
	}

	// --- 配置 ---
	// workDir 是包含已解密数据库文件的目录。
	workDir := viper.GetString("WORK_DIR")
	if workDir == "" {
		workDir = "data"
	}

	// 端口配置：优先使用 LISTEN_ADDR，其次使用 PORT，最后默认 127.0.0.1:8080
	listenAddr := viper.GetString("LISTEN_ADDR")
	port := viper.GetString("PORT")
	if listenAddr == "" {
		if port != "" {
			listenAddr = "127.0.0.1:" + port
		} else {
			listenAddr = "127.0.0.1:5200"
		}
	}

	imageKey := viper.GetString("IMAGE_KEY")
	if imageKey == "" {
		imageKey = ""
	}

	xorKey := viper.GetString("XOR_KEY")
	if xorKey == "" {
		xorKey = ""
	}

	log.Printf("使用工作目录: %s", workDir)

	// 确保工作目录存在
	if err := os.MkdirAll(workDir, 0755); err != nil {
		log.Fatalf("创建工作目录失败: %v", err)
	}

	// --- 初始化 Store ---
	newStore, err := store.NewStore(workDir)
	if err != nil {
		log.Fatalf("初始化 store 失败: %v", err)
	}
	defer newStore.Close()
	log.Println("Store 初始化成功。")

	// --- 准备静态文件系统 ---
	staticFS, err := fs.Sub(uiDist, "ui/dist")
	if err != nil {
		log.Fatalf("无法加载嵌入的 UI 文件: %v", err)
	}

	// --- 初始化 Web 服务 ---
	webConf := web.Config{
		ListenAddr:      listenAddr,
		DataDir:         workDir,
		ImageKey:        imageKey,
		XorKey:          xorKey,
		WechatDbSrcPath: viper.GetString("WECHAT_DB_SRC_PATH"),
		WechatDbKey:     viper.GetString("WECHAT_DB_KEY"),
		WxKeyDllPath:    viper.GetString("WXKEY_DLL_PATH"),
		WechatPath:      viper.GetString("WXKEY_WECHAT_PATH"),
		WechatDataPath:  viper.GetString("WXKEY_WECHAT_DATA_PATH"),
	}
	webService := web.NewService(newStore, &webConf, staticFS)

	// --- 启动服务 ---
	if err := webService.Start(); err != nil {
		log.Fatalf("启动 web 服务失败: %v", err)
	}

	// 打印访问地址并自动打开浏览器
	baseURL := listenAddr
	if len(baseURL) > 0 && baseURL[0] == ':' {
		baseURL = "127.0.0.1" + baseURL
	}
	url := "http://" + baseURL
	log.Printf("服务已启动，请访问: %s", url)
	openBrowser(url)

	// --- 等待中断信号以实现优雅关闭 ---
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("接收到关闭信号，正在关闭服务...")

	// --- 关闭服务 ---
	if err := webService.Stop(); err != nil {
		log.Fatalf("关闭 web 服务时出错: %v", err)
	}
	log.Println("服务已成功关闭。")
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = nil
	}
	if err != nil {
		log.Printf("无法自动打开浏览器: %v", err)
	}
}
