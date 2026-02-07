import { useState, useEffect, useRef } from "react"
import { X, Key, Image as ImageIcon, Copy, Check, Loader2, AlertCircle, FolderSearch } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { systemApi } from "@/api"

interface Props {
  onClose: () => void
}

export function KeyManagerModal({ onClose }: Props) {
  const [dbKey, setDbKey] = useState<string | null>(null)
  const [imageKey, setImageKey] = useState<{ xor: string; aes: string } | null>(null)
  
  const [wechatPath, setWechatPath] = useState<string>("")
  const [dbSrcPath, setDbSrcPath] = useState<string>("")
  const [detectedWechatPaths, setDetectedWechatPaths] = useState<string[] | null>(null)
  const [detectedDbPaths, setDetectedDbPaths] = useState<string[] | null>(null)

  const [loading, setLoading] = useState<'db' | 'image' | 'detect_wechat' | 'detect_db' | 'init' | null>('init')
  const [error, setError] = useState<string | null>(null)
  const [copied, setCopied] = useState<string | null>(null)

  const isInitialMount = useRef(true)

  // 初始化检查现有配置
  useEffect(() => {
    const checkExistingKeys = async () => {
      try {
        const res: any = await systemApi.getStatus()
        const config = res?.config || res?.data?.config
        
        if (config) {
          if (config.wechat_db_key) {
            setDbKey(config.wechat_db_key)
          }
          if (config.image_key) {
            setImageKey({ aes: config.image_key, xor: config.xor_key || "" })
          }
          if (config.wechat_path) setWechatPath(config.wechat_path)
          if (config.wechat_db_src_path) setDbSrcPath(config.wechat_db_src_path)
          
          // 如果路径为空，自动触发检测
          if (!config.wechat_path) handleDetectWeChat();
          if (!config.wechat_db_src_path) handleDetectDbPath();
        }
      } catch (err) {
        console.error("Failed to fetch initial status:", err)
      } finally {
        setLoading(null)
        // 延迟标记初始化完成，避免初始化赋值触发自动保存
        setTimeout(() => { isInitialMount.current = false }, 100)
      }
    }
    checkExistingKeys()
  }, [])

  // 微信路径自动保存
  useEffect(() => {
    if (isInitialMount.current) return
    const timer = setTimeout(() => {
      handleSaveConfig('WXKEY_WECHAT_PATH', wechatPath)
    }, 800)
    return () => clearTimeout(timer)
  }, [wechatPath])

  // 数据库路径自动保存
  useEffect(() => {
    if (isInitialMount.current) return
    const timer = setTimeout(() => {
      handleSaveConfig('WECHAT_DB_SRC_PATH', dbSrcPath)
    }, 800)
    return () => clearTimeout(timer)
  }, [dbSrcPath])

  const handleGetDbKey = async () => {
    setLoading('db')
    setError(null)
    try {
      const res: any = await systemApi.getWeChatDbKey()
      console.log('DB Key Response:', res)
      
      const data = res?.data || res
      
      if (data && data.key) {
        setDbKey(data.key)
      } else {
        setError("获取失败：响应中未包含密钥字段")
      }
    } catch (err: any) {
      console.error('DB Key Error:', err)
      setError(err.message || "请求失败")
    } finally {
      setLoading(null)
    }
  }

  const handleGetImageKey = async () => {
    setLoading('image')
    setError(null)
    try {
      const res: any = await systemApi.getWeChatImageKey()
      console.log('Image Key Response:', res)
      
      const data = res?.data || res
      
      if (data && data.image_aes_key) {
        setImageKey({ xor: data.image_xor_key, aes: data.image_aes_key })
      } else {
        setError("获取失败：响应中未包含图片密钥数据")
      }
    } catch (err: any) {
      console.error('Image Key Error:', err)
      setError(err.message || "请求失败")
    } finally {
      setLoading(null)
    }
  }

  const handleDetectWeChat = async () => {
    setLoading('detect_wechat')
    setError(null)
    setDetectedWechatPaths(null)
    try {
      const res: any = await systemApi.detectWeChatPath()
      const paths = res?.data || res
      if (Array.isArray(paths) && paths.length > 0) {
        if (paths.length === 1) {
          // 只有一个结果，自动选择
          setWechatPath(paths[0])
          await handleSaveConfig('WXKEY_WECHAT_PATH', paths[0])
        } else {
          // 多个结果，让用户选择
          setDetectedWechatPaths(paths)
        }
      } else {
        setError("未检测到微信安装路径，请手动输入")
      }
    } catch (err: any) {
      setError("检测微信路径失败: " + (err.message || "未知错误"))
    } finally {
      setLoading(null)
    }
  }

  const handleDetectDbPath = async () => {
    setLoading('detect_db')
    setError(null)
    setDetectedDbPaths(null)
    try {
      const res: any = await systemApi.detectDbPath()
      const paths = res?.data || res
      if (Array.isArray(paths) && paths.length > 0) {
        if (paths.length === 1) {
          // 只有一个结果，自动选择
          setDbSrcPath(paths[0])
          await handleSaveConfig('WECHAT_DB_SRC_PATH', paths[0])
        } else {
          // 多个结果，让用户选择
          setDetectedDbPaths(paths)
        }
      } else {
        setError("未检测到数据存储路径")
      }
    } catch (err: any) {
      setError("检测数据路径失败: " + (err.message || "未知错误"))
    } finally {
      setLoading(null)
    }
  }

  const handleSaveConfig = async (key: string, value: string) => {
    try {
      await systemApi.updateConfig({ [key]: value })
      if (key === 'WXKEY_WECHAT_PATH') {
        setDetectedWechatPaths(null)
      } else if (key === 'WECHAT_DB_SRC_PATH') {
        setDetectedDbPaths(null)
      }
    } catch (err) {
      console.error("Failed to save config:", err)
    }
  }

  const copyToClipboard = (text: string, type: string) => {
    navigator.clipboard.writeText(text)
    setCopied(type)
    setTimeout(() => setCopied(null), 2000)
  }

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/60 backdrop-blur-sm p-4 animate-in fade-in duration-200">
      <div className="bg-background border shadow-2xl rounded-2xl w-full max-w-lg flex flex-col overflow-hidden animate-in zoom-in-95 duration-200 max-h-[90vh]">
        {/* Header */}
        <div className="p-6 border-b flex items-center justify-between bg-muted/20">
          <div className="flex items-center gap-3">
            <div className="bg-primary/10 p-2 rounded-lg text-primary">
              <Key className="w-5 h-5" />
            </div>
            <div>
              <h3 className="font-bold text-xl">系统配置与密钥</h3>
              <p className="text-xs text-muted-foreground mt-1">配置微信路径与获取解密密钥</p>
            </div>
          </div>
          <Button variant="ghost" size="icon" onClick={onClose} className="rounded-full">
            <X className="h-5 w-5" />
          </Button>
        </div>

        {/* Content */}
        <div className="p-6 space-y-6 overflow-y-auto">
          {error && (
            <div className="bg-destructive/10 border border-destructive/20 text-destructive text-sm p-3 rounded-xl flex items-start gap-3 animate-in slide-in-from-top-2">
              <AlertCircle className="w-5 h-5 shrink-0 mt-0.5" />
              <p>{error}</p>
            </div>
          )}

          {/* Path Configuration Section */}
          <div className="space-y-4">
             <h4 className="text-sm font-bold flex items-center gap-2">
                <FolderSearch className="w-4 h-4 text-purple-500" />
                路径配置
              </h4>
             
             {/* WeChat Path */}
             <div className="space-y-2">
                <label className="text-xs font-medium text-muted-foreground">微信程序路径 (WeChat.exe/Weixin.exe)</label>
                {(error?.includes("微信安装路径") || (!loading && !wechatPath)) && (
                    <p className="text-[11px] text-destructive font-bold animate-pulse">
                        {"⚠️ 自动检测未果？请右键点击桌面微信图标 -> \"打开文件所在位置\"，将该目录下的 WeChat.exe 全路径复制到下方。"}
                    </p>
                )}
                <div className="flex gap-2">
                    <Input 
                        value={wechatPath} 
                        onChange={(e) => setWechatPath(e.target.value)}
                        placeholder="请输入或粘贴微信安装路径"
                        className="text-xs font-mono h-8"
                    />
                    <Button 
                        size="sm" 
                        variant="outline"
                        onClick={handleDetectWeChat}
                        disabled={!!loading}
                        className="h-8 whitespace-nowrap"
                    >
                        {loading === 'detect_wechat' ? <Loader2 className="w-3 h-3 animate-spin" /> : "自动检测"}
                    </Button>
                </div>
                {detectedWechatPaths && (
                    <div className="bg-muted/50 p-2 rounded-lg border space-y-2 animate-in slide-in-from-top-2">
                        <div className="text-xs text-muted-foreground font-medium px-1">检测到以下路径，请选择：</div>
                        {detectedWechatPaths.map((path, i) => (
                            <button 
                                key={i}
                                onClick={() => {
                                    setWechatPath(path)
                                    handleSaveConfig('WXKEY_WECHAT_PATH', path)
                                }}
                                className="w-full text-left text-xs p-2 rounded hover:bg-background border border-transparent hover:border-border transition-colors truncate font-mono"
                                title={path}
                            >
                                {path}
                            </button>
                        ))}
                    </div>
                )}
             </div>

             {/* DB Path */}
             <div className="space-y-2">
                <label className="text-xs font-medium text-muted-foreground flex items-center gap-1">
                    数据存储路径 (包含 wxid_xxx 文件夹)
                    <span className="text-destructive font-bold">* 核心配置</span>
                </label>
                {(error?.includes("数据存储路径") || (!loading && !dbSrcPath)) && (
                    <p className="text-[11px] text-destructive font-bold animate-pulse">
                        {"⚠️ 找不到数据？请在手机微信中点击：设置 -> 文件管理 -> 打开文件夹，在电脑窗口中向上退一级，找到包含 wxid_xxxx 的目录路径复制到下方。"}
                    </p>
                )}
                <div className="flex gap-2">
                    <Input 
                        value={dbSrcPath} 
                        onChange={(e) => setDbSrcPath(e.target.value)}
                        placeholder="请输入或粘贴微信数据存储路径"
                        className="text-xs font-mono h-8 border-orange-200 focus-visible:ring-orange-500"
                    />
                    <Button 
                        size="sm" 
                        variant="outline"
                        onClick={handleDetectDbPath}
                        disabled={!!loading}
                        className="h-8 whitespace-nowrap"
                    >
                         {loading === 'detect_db' ? <Loader2 className="w-3 h-3 animate-spin" /> : "自动检测"}
                    </Button>
                </div>
                <div className="py-1 space-y-2">
                    <div className="flex gap-2 items-start text-destructive dark:text-red-400">
                        <AlertCircle className="w-4 h-4 shrink-0 mt-0.5" />
                        <div className="text-xs font-medium leading-relaxed">
                            <p className="font-bold mb-1 text-sm">⚠️ 该路径配置至关重要：</p>
                            <p>1. 路径必须指向包含 <code className="bg-destructive/10 dark:bg-red-900/20 px-1 rounded font-bold border border-destructive/20">wxid_xxxx</code> 文件夹的目录。</p>
                            <p>2. 路径错误或未包含用户信息目录将导致<span className="underline font-bold">无法读取聊天记录</span>。</p>
                            <p>3. 若解密后<span className="underline font-bold">没有看到任何数据</span>，请核对该目录下是否存在以 <code className="font-mono bg-destructive/5 px-1">wxid_</code> 开头的文件夹。</p>
                        </div>
                    </div>
                </div>
                {detectedDbPaths && (
                    <div className="bg-muted/50 p-2 rounded-lg border space-y-2 animate-in slide-in-from-top-2">
                        <div className="text-xs text-muted-foreground font-medium px-1">检测到以下路径，请选择：</div>
                        {detectedDbPaths.map((path, i) => (
                            <button 
                                key={i}
                                onClick={() => {
                                    setDbSrcPath(path)
                                    handleSaveConfig('WECHAT_DB_SRC_PATH', path)
                                }}
                                className="w-full text-left text-xs p-2 rounded hover:bg-background border border-transparent hover:border-border transition-colors truncate font-mono"
                                title={path}
                            >
                                {path}
                            </button>
                        ))}
                    </div>
                )}
             </div>
             <p className="text-[10px] text-muted-foreground">提示：输入路径后系统将自动保存配置。</p>
          </div>

          <div className="border-t border-dashed" />

          {/* Database Key Section */}
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <h4 className="text-sm font-bold flex items-center gap-2">
                <Key className="w-4 h-4 text-blue-500" />
                数据库解密密钥
              </h4>
              <Button 
                size="sm" 
                onClick={handleGetDbKey} 
                disabled={!!loading}
                className="rounded-full h-8 px-4"
              >
                {loading === 'db' ? (
                  <>
                    <Loader2 className="w-3 h-3 mr-2 animate-spin" />
                    正在获取...
                  </>
                ) : (dbKey ? "重新获取" : "点击获取")}
              </Button>
            </div>
            
            {dbKey && (
              <div className="relative group animate-in slide-in-from-left-2">
                <div className="bg-muted font-mono text-xs p-4 rounded-xl break-all pr-10 border border-border/50">
                  {dbKey}
                </div>
                <button 
                  onClick={() => copyToClipboard(dbKey, 'db')}
                  className="absolute right-3 top-1/2 -translate-y-1/2 p-1.5 hover:bg-background rounded-md transition-colors"
                >
                  {copied === 'db' ? <Check className="w-4 h-4 text-green-500" /> : <Copy className="w-4 h-4 text-muted-foreground" />}
                </button>
              </div>
            )}

            <div className="py-1 flex gap-2 items-start text-destructive dark:text-red-400">
              <AlertCircle className="w-4 h-4 shrink-0 mt-0.5" />
              <p className="text-xs font-medium leading-relaxed">
                <span className="font-bold">注意：</span>获取过程中<span className="font-bold underline">微信将自动重启</span>。成功捕获密钥后微信可能会自动退出，这属于正常情况，您只需<span className="font-bold underline">再次手动登录</span>即可正常使用。
              </p>
            </div>
          </div>

          <div className="border-t border-dashed" />

          {/* Image Key Section */}
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <h4 className="text-sm font-bold flex items-center gap-2">
                <ImageIcon className="w-4 h-4 text-emerald-500" />
                图片查看密钥
              </h4>
              <Button 
                size="sm" 
                variant="outline"
                onClick={handleGetImageKey} 
                disabled={!!loading}
                className="rounded-full h-8 px-4"
              >
                {loading === 'image' ? (
                  <>
                    <Loader2 className="w-3 h-3 mr-2 animate-spin" />
                    扫描中...
                  </>
                ) : (imageKey ? "重新获取" : "点击获取")}
              </Button>
            </div>

            {imageKey && (
              <div className="space-y-2 animate-in slide-in-from-right-2">
                <div className="relative group">
                  <div className="text-[10px] text-muted-foreground mb-1 ml-1 font-bold uppercase tracking-wider">Image Key</div>
                  <div className="bg-muted font-mono text-xs p-3 rounded-xl break-all pr-10 border border-border/50">
                    {imageKey.aes}
                  </div>
                  <button 
                    onClick={() => copyToClipboard(imageKey.aes, 'aes')}
                    className="absolute right-3 bottom-2.5 p-1.5 hover:bg-background rounded-md transition-colors"
                  >
                    {copied === 'aes' ? <Check className="w-4 h-4 text-green-500" /> : <Copy className="w-4 h-4 text-muted-foreground" />}
                  </button>
                </div>
                <div className="relative group">
                  <div className="text-[10px] text-muted-foreground mb-1 ml-1 font-bold uppercase tracking-wider">XOR Key</div>
                  <div className="bg-muted font-mono text-xs p-3 rounded-xl break-all pr-10 border border-border/50">
                    {imageKey.xor}
                  </div>
                  <button 
                    onClick={() => copyToClipboard(imageKey.xor, 'xor')}
                    className="absolute right-3 bottom-2.5 p-1.5 hover:bg-background rounded-md transition-colors"
                  >
                    {copied === 'xor' ? <Check className="w-4 h-4 text-green-500" /> : <Copy className="w-4 h-4 text-muted-foreground" />}
                  </button>
                </div>
              </div>
            )}
            <div className="py-1 flex gap-2 items-start text-destructive dark:text-red-400">
              <ImageIcon className="w-4 h-4 shrink-0 mt-0.5" />
              <p className="text-xs font-medium leading-relaxed">
                <span className="font-bold">提示：</span>点击扫描后，请在微信中<span className="font-bold underline">连续打开多张</span>聊天图片（或左右切换浏览），直到密钥显示。
              </p>
            </div>
          </div>
        </div>

        {/* Footer */}
        <div className="p-4 bg-muted/10 border-t text-center text-[10px] text-muted-foreground">
          获取成功后将自动保存至服务器配置中。
        </div>
      </div>
      <div className="absolute inset-0 -z-10" onClick={onClose} />
    </div>
  )
}
