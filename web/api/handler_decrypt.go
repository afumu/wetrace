package api

import (
	"github.com/afumu/wetrace/decrypt"
	"github.com/afumu/wetrace/web/transport"
	"github.com/gin-gonic/gin"
)

// HandleEnvDecrypt 处理基于环境变量的本地解密请求
func (a *API) HandleEnvDecrypt(c *gin.Context) {
	count, outputDir, err := decrypt.RunTask(a.Conf.WechatDbSrcPath, a.Conf.WechatDbKey)
	if err != nil {
		transport.InternalServerError(c, "解密失败: "+err.Error())
		return
	}

	// 解密成功后，触发 Store 的重载，以便立即发现新文件
	if err := a.Store.Reload(); err != nil {
		transport.InternalServerError(c, "解密成功，但重载数据存储失败: "+err.Error())
		return
	}

	transport.SendSuccess(c, gin.H{
		"processed_files": count,
		"output_dir":      outputDir,
	})
}
