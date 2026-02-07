package api

import (
	"github.com/afumu/wetrace/pkg/util"
	"github.com/afumu/wetrace/web/transport"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// SelectPath 让用户在服务端（本地）选择路径
func (a *API) SelectPath(c *gin.Context) {
	type SelectReq struct {
		Type string `json:"type"` // "file" or "folder"
	}
	var req SelectReq
	if err := c.ShouldBindJSON(&req); err != nil {
		transport.BadRequest(c, "参数错误")
		return
	}

	var path string
	var err error

	if req.Type == "file" {
		path, err = util.OpenFileDialog("选择微信可执行文件", "WeChat Executable|WeChat.exe;Weixin.exe|All Files|*.*")
	} else {
		path, err = util.OpenFolderDialog("选择微信数据目录 (xwechat_files)")
	}

	if err != nil {
		// 用户取消
		if err.Error() == "cancelled" {
			transport.SendSuccess(c, gin.H{"path": ""})
			return
		}
		// 其他错误
		transport.InternalServerError(c, "打开对话框失败: "+err.Error())
		return
	}

	transport.SendSuccess(c, gin.H{"path": path})
}

// GetSystemStatus 返回应用程序的当前状态。
// 目前，它只确认服务正在运行。
func (a *API) GetSystemStatus(c *gin.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 获取当前配置中的密钥，用于前端判断是否存在
	status := gin.H{
		"store_initialized": true,
		"config": gin.H{
			"wechat_db_key":      a.Conf.WechatDbKey,
			"image_key":          a.Media.ImageKey,
			"xor_key":            a.Media.XorKey,
			"wechat_path":        a.Conf.WechatPath,
			"wechat_db_src_path": a.Conf.WechatDbSrcPath,
		},
	}
	transport.SendSuccess(c, status)
}

// DetectWeChatInstallPath 检测微信安装路径
func (a *API) DetectWeChatInstallPath(c *gin.Context) {
	paths := util.FindWeChatInstallPaths()
	transport.SendSuccess(c, paths)
}

// DetectWeChatDataPath 检测微信数据路径
func (a *API) DetectWeChatDataPath(c *gin.Context) {
	paths := util.FindWeChatDataPaths()
	transport.SendSuccess(c, paths)
}

// UpdateConfig 更新系统配置
func (a *API) UpdateConfig(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		transport.BadRequest(c, "参数错误")
		return
	}

	// 允许更新的键白名单
	allowedKeys := map[string]bool{
		"WXKEY_WECHAT_PATH":  true,
		"WECHAT_DB_SRC_PATH": true,
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	changed := false
	for k, v := range req {
		if allowedKeys[k] {
			viper.Set(k, v)
			// 同步更新内存中的配置对象
			if k == "WXKEY_WECHAT_PATH" {
				a.Conf.WechatPath = v
			} else if k == "WECHAT_DB_SRC_PATH" {
				a.Conf.WechatDbSrcPath = v
			}
			changed = true
		}
	}

	if changed {
		if err := viper.WriteConfig(); err != nil {
			// 如果文件不存在，尝试创建
			if err := viper.WriteConfigAs(".env"); err != nil {
				transport.InternalServerError(c, "保存配置文件失败: "+err.Error())
				return
			}
		}
	}

	transport.SendSuccess(c, gin.H{"status": "ok"})
}
