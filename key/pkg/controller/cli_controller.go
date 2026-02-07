package controller

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/afumu/wetrace/key/pkg/dllloader"
	"github.com/afumu/wetrace/key/pkg/imagekey"
	"github.com/afumu/wetrace/key/pkg/logger"
	"github.com/afumu/wetrace/key/pkg/options"
	"github.com/afumu/wetrace/key/pkg/process"
	"github.com/afumu/wetrace/key/pkg/utils"
	"github.com/afumu/wetrace/wxkey"
)

type ExecutionResult struct {
	Success       bool   `json:"success"`
	Key           string `json:"key,omitempty"`
	ImageXorKey   int    `json:"image_xor_key,omitempty"`
	ImageAesKey   string `json:"image_aes_key,omitempty"`
	Pid           uint32 `json:"pid,omitempty"`
	WechatVersion string `json:"wechat_version,omitempty"`
	ElapsedTimeMs int64  `json:"elapsed_time_ms"`
	Timestamp     string `json:"timestamp"`
	Error         string `json:"error,omitempty"`
	ErrorCode     string `json:"error_code,omitempty"`
}

type CliController struct {
	options        options.CliOptions
	logger         *logger.Logger
	processManager *process.ProcessManager
	dllLoader      *dllloader.DllLoader
	startTime      time.Time
}

func NewCliController(opts options.CliOptions) *CliController {
	log := logger.NewLogger(opts.Verbose, opts.Quiet, opts.NoColor)
	exeDir, _ := utils.GetExecutableDirectory()
	log.SetLogFile(utils.JoinPath(exeDir, "wx_key_cli_go_log.txt"))

	return &CliController{
		options:        opts,
		logger:         log,
		processManager: process.NewProcessManager(),
		dllLoader:      dllloader.NewDllLoader(),
	}
}

func (c *CliController) Run() int {
	c.startTime = time.Now()
	result := ExecutionResult{
		Timestamp: utils.GetCurrentTimestamp(),
	}

	defer c.Cleanup()

	// Special Mode: Image Key
	if c.options.ImageKeyMode {
		return c.runImageKeyMode(&result)
	}

	// 1. Prepare WeChat Process
	c.logger.Info("正在准备微信进程...")
	pid, err := c.prepareWeChatProcess()
	if err != nil {
		return c.handleError(err, &result)
	}
	result.Pid = pid
	c.logger.Success(fmt.Sprintf("微信进程准备完成，PID: %d", pid))

	// 2. Load DLL
	c.logger.Info("正在加载 wx_key.dll...")
	if err := c.loadDll(); err != nil {
		return c.handleError(err, &result)
	}
	c.logger.Success("DLL 加载成功")

	// 3. Install Hook
	c.logger.Info("正在安装 Hook...")
	if err := c.installHook(pid); err != nil {
		return c.handleError(err, &result)
	}
	c.logger.Success("Hook 安装成功")

	// 4. Wait for Key
	c.logger.Info("正在等待密钥...")
	key, err := c.waitForKey()
	if err != nil {
		return c.handleError(err, &result)
	}
	result.Key = key

	// Success
	result.Success = true
	result.ElapsedTimeMs = time.Since(c.startTime).Milliseconds()

	// Get version if possible (not implemented fully)
	result.WechatVersion = "" // Placeholder

	c.outputResult(result)
	return 0
}

func (c *CliController) prepareWeChatProcess() (uint32, error) {
	// Manual Mode
	if c.options.ManualPid > 0 {
		c.logger.Debug(fmt.Sprintf("使用手动指定的 PID: %d", c.options.ManualPid))
		// Verification skipped for simplicity, or we can check if it exists
		return c.options.ManualPid, nil
	}

	// Auto Mode
	if c.options.AutoMode {
		isRunning := c.processManager.IsProcessRunning("Weixin.exe")

		if isRunning && !c.options.NoRestart {
			c.logger.Info("检测到微信正在运行，正在关闭...")
			if err := c.processManager.KillProcess("Weixin.exe"); err != nil {
				return 0, fmt.Errorf("关闭微信失败: %v", err)
			}
			c.logger.Success("微信已关闭")
			time.Sleep(2 * time.Second)
		}

		if !isRunning || !c.options.NoRestart {
			c.logger.Info("正在启动微信...")
			wechatPath := c.findWeChatPath()
			if wechatPath == "" {
				return 0, fmt.Errorf("未找到微信安装路径，请使用 --wechat-path 手动指定")
			}
			if err := c.processManager.LaunchWeChat(wechatPath); err != nil {
				return 0, fmt.Errorf("微信启动失败: %v", err)
			}
			c.logger.Success("微信启动成功")

			c.logger.Info("等待微信窗口出现...")
			if !c.processManager.WaitForWeChatWindow(c.options.StartupWaitTimeout) {
				c.logger.Warning("等待微信窗口超时或窗口未显示（可能是登录界面或最小化），尝试直接获取 PID...")
			} else {
				c.logger.Success("微信窗口已出现")
			}

			// Wait a bit more for components
			c.logger.Info("等待微信进程初始化...")
			time.Sleep(2 * time.Second)
		}

		c.logger.Info("正在获取微信 PID...")
		mainPid := c.processManager.FindMainWeChatPid()
		if mainPid == 0 {
			c.logger.Warning("未找到主窗口关联 PID，切换为进程扫描模式...")
			pid, err := c.processManager.GetProcessId("Weixin.exe")
			if err != nil {
				return 0, fmt.Errorf("未找到微信进程: %v", err)
			}
			return pid, nil
		}
		return mainPid, nil
	}

	return 0, fmt.Errorf("无效的工作模式")
}

func (c *CliController) findWeChatPath() string {
	if c.options.WechatPath != "" {
		if utils.FileExists(c.options.WechatPath) {
			return c.options.WechatPath
		}
	}
	path := c.processManager.FindWeChatPath()
	if path == "" {
		// Fallback error
		return ""
	}
	return path
}

func (c *CliController) loadDll() error {
	dllPath := c.options.DllPath
	if dllPath == "" {
		if path, err := wxkey.GetDllPath(); err == nil {
			dllPath = path
			c.logger.Debug("使用嵌入的 DLL: " + dllPath)
		} else {
			exeDir, _ := utils.GetExecutableDirectory()
			dllPath = utils.JoinPath(exeDir, "wx_key.dll")
			c.logger.Warning("获取嵌入 DLL 失败，回退到默认路径")
		}
	}
	c.logger.Debug("DLL 路径: " + dllPath)
	return c.dllLoader.Load(dllPath)
}

func (c *CliController) installHook(pid uint32) error {
	return c.dllLoader.InitializeHook(pid)
}

func (c *CliController) waitForKey() (string, error) {
	deadline := time.Now().Add(time.Duration(c.options.KeyWaitTimeout) * time.Second)

	for time.Now().Before(deadline) {
		// Poll Key
		if key, ok := c.dllLoader.PollKeyData(); ok {
			c.logger.Success("密钥获取成功")
			return key, nil
		}

		// Process Messages
		c.processStatusMessages()

		time.Sleep(100 * time.Millisecond)
	}
	return "", fmt.Errorf("密钥获取超时")
}

func (c *CliController) processStatusMessages() {
	for i := 0; i < 5; i++ {
		msg, level, ok := c.dllLoader.GetStatusMessage()
		if !ok {
			break
		}
		switch level {
		case 0:
			c.logger.Info("[DLL] " + msg)
		case 1:
			c.logger.Success("[DLL] " + msg)
		case 2:
			c.logger.Error("[DLL] " + msg)
		}
	}
}

func (c *CliController) runImageKeyMode(result *ExecutionResult) int {
	c.logger.Info("进入图片密钥获取模式...")

	firstAttempt := true
	// Create a logger for background polling: quiet on console, but writes to file
	pollingLogger := *c.logger
	pollingLogger.Quiet = true
	pollingLogger.MuteAll = false

	for {
		var pid uint32
		var err error
		var currentLogger *logger.Logger

		if firstAttempt {
			currentLogger = c.logger
		} else {
			currentLogger = &pollingLogger
		}

		currentLogger.Info("开始新一轮图片密钥扫描...")

		// 1. Determine PID
		if c.options.ManualPid > 0 {
			if firstAttempt {
				currentLogger.Debug(fmt.Sprintf("使用手动指定的 PID: %d", c.options.ManualPid))
			}
			pid = c.options.ManualPid
		} else {
			// Try to find running WeChat
			if firstAttempt {
				currentLogger.Info("正在查找运行中的微信进程...")
			}

			// First try finding by window (more reliable for main process)
			pid = c.processManager.FindMainWeChatPid()

			if pid == 0 {
				// Fallback to process name search
				if firstAttempt {
					currentLogger.Debug("未检测到微信主窗口，尝试直接搜索进程名...")
				}
				pid, err = c.processManager.GetProcessId("Weixin.exe")
				if err != nil || pid == 0 {
					if firstAttempt {
						c.logger.Warning("未找到运行中的微信进程，请启动微信并登录")
					}
					currentLogger.Info("未找到运行中的微信进程 (Weixin.exe)")
				}
			}
		}

		if pid > 0 {
			if firstAttempt {
				c.logger.Success(fmt.Sprintf("目标微信进程 PID: %d", pid))
			}
			currentLogger.Info(fmt.Sprintf("正在对 PID %d 执行密钥提取...", pid))

			// 2. Get Keys
			svc := imagekey.NewImageKeyService(currentLogger)
			keyResult := svc.GetImageKeys(pid, c.options.WechatDataPath)

			if keyResult.Success {
				result.Pid = pid
				result.Success = true
				result.ImageXorKey = keyResult.XorKey
				result.ImageAesKey = keyResult.AesKey
				result.ElapsedTimeMs = time.Since(c.startTime).Milliseconds()

				if !firstAttempt {
					if !c.options.Quiet {
						fmt.Println()
					}
					c.logger.Success("成功获取图片密钥！")
				}

				c.outputResult(*result)
				return 0
			} else {
				if firstAttempt {
					c.logger.Warning(fmt.Sprintf("获取失败: %v", keyResult.Error))
				}
				currentLogger.Info(fmt.Sprintf("该轮尝试失败: %v", keyResult.Error))
			}
		}

		if firstAttempt {
			c.logger.Info("正在等待图片密钥（每 3 秒扫描一次）...")
			c.logger.Info("请在微信中打开任意图片...")
			c.logger.Info("按 Ctrl+C 退出")
			firstAttempt = false
		} else {
			if !c.options.Quiet {
				fmt.Print(".")
			}
		}

		time.Sleep(3 * time.Second)
	}
}

func (c *CliController) Cleanup() {
	c.logger.Debug("正在清理资源...")
	c.dllLoader.CleanupHook()
	c.logger.Close()
}

func (c *CliController) handleError(err error, result *ExecutionResult) int {
	c.logger.Error(fmt.Sprintf("错误: %v", err))
	result.Success = false
	result.Error = err.Error()
	result.ErrorCode = "EXECUTION_FAILED"
	result.ElapsedTimeMs = time.Since(c.startTime).Milliseconds()

	if c.options.OutputFormat == "json" {
		c.outputJson(result)
	}
	return 1
}

func (c *CliController) outputResult(result ExecutionResult) {
	if c.options.OutputFormat == "json" {
		c.outputJson(&result)
	} else {
		c.outputText(&result)
	}
}

func (c *CliController) outputText(result *ExecutionResult) {
	fmt.Println()
	fmt.Println("========================================")
	if result.Key != "" {
		fmt.Println("           数据库密钥获取成功")
	} else if result.ImageAesKey != "" {
		fmt.Println("           图片密钥获取成功")
	} else {
		fmt.Println("           操作完成")
	}
	fmt.Println("========================================")
	fmt.Println()

	if result.Key != "" {
		fmt.Printf("数据库密钥: %s\n", result.Key)
	}

	if result.ImageAesKey != "" {
		fmt.Printf("图片 XOR 密钥: %02X\n", result.ImageXorKey)
		fmt.Printf("图片 AES 密钥: %s\n", result.ImageAesKey)
	}

	fmt.Printf("微信进程 PID: %d\n", result.Pid)
	fmt.Printf("耗时: %d 毫秒\n", result.ElapsedTimeMs)

	// Log result to file
	logMsg := fmt.Sprintf("PID: %d, Success: %v", result.Pid, result.Success)
	if result.Key != "" {
		logMsg += fmt.Sprintf(", Key: %s", result.Key)
	}
	if result.ImageAesKey != "" {
		logMsg += fmt.Sprintf(", ImageXor: %02X, ImageAes: %s", result.ImageXorKey, result.ImageAesKey)
	}
	c.logger.LogToFile(logMsg)

	if c.options.OutputFile != "" {
		content := ""
		if result.Key != "" {
			content += result.Key
		}
		if result.ImageAesKey != "" {
			if content != "" {
				content += "\n"
			}
			content += fmt.Sprintf("XOR: %02X\nAES: %s", result.ImageXorKey, result.ImageAesKey)
		}

		if err := os.WriteFile(c.options.OutputFile, []byte(content), 0666); err == nil {
			c.logger.Success("密钥已保存到文件: " + c.options.OutputFile)
		} else {
			c.logger.Warning("无法写入输出文件: " + c.options.OutputFile)
		}
	}
}

func (c *CliController) outputJson(result *ExecutionResult) {
	// If extended JSON is needed, we would add more fields here.
	// For now, standard JSON.
	data, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(data))

	if c.options.OutputFile != "" {
		os.WriteFile(c.options.OutputFile, data, 0666)
	}
}
