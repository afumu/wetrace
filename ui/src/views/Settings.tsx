import { useState, useEffect, useMemo } from "react"
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { systemApi, sessionApi } from "@/api"
import { toast } from "sonner"
import type {
  AIConfigUpdate,
  SyncConfigUpdate,
  BackupConfigUpdate,
  TTSConfigUpdate,
} from "@/api/system"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Switch } from "@/components/ui/switch"
import {
  Bot,
  RefreshCw,
  Lock,
  HardDrive,
  Loader2,
  CheckCircle,
  XCircle,
  MessageSquare,
  RotateCcw,
  Search,
  X,
  Mic,
} from "lucide-react"

/* ============================================================
 * AI Config Section
 * ============================================================ */
function AIConfigSection() {
  const queryClient = useQueryClient()
  const [form, setForm] = useState<AIConfigUpdate>({
    enabled: false,
    provider: "openai",
    model: "",
    base_url: "",
    api_key: "",
  })
  const [testStatus, setTestStatus] = useState<"idle" | "testing" | "success" | "error">("idle")
  const [showPromptsDialog, setShowPromptsDialog] = useState(false)

  const { data: config, isLoading } = useQuery({
    queryKey: ["ai-config"],
    queryFn: () => systemApi.getAIConfig(),
  })

  useEffect(() => {
    if (config) {
      setForm({
        enabled: config.enabled,
        provider: config.provider || "openai",
        model: config.model || "",
        base_url: config.base_url || "",
        api_key: "",
      })
    }
  }, [config])

  const updateMutation = useMutation({
    mutationFn: (data: AIConfigUpdate) => systemApi.updateAIConfig(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["ai-config"] })
      toast.success("AI 配置已保存")
    },
    onError: (err: Error) => toast.error("保存失败: " + err.message),
  })

  const handleTest = async () => {
    setTestStatus("testing")
    try {
      await systemApi.testAIConfig()
      setTestStatus("success")
    } catch {
      setTestStatus("error")
    }
  }

  if (isLoading) {
    return (
      <Card>
        <CardContent className="p-6 flex items-center justify-center">
          <Loader2 className="w-5 h-5 animate-spin text-muted-foreground" />
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base flex items-center gap-2">
          <Bot className="w-4 h-4 text-primary" />
          AI 大模型配置
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex items-center justify-between">
          <label className="text-sm font-medium leading-none">启用 AI 功能</label>
          <Switch
            checked={form.enabled}
            onCheckedChange={(checked) => setForm((f) => ({ ...f, enabled: checked }))}
          />
        </div>

        {form.enabled && (
          <>
            <div className="space-y-1.5">
              <label className="text-sm font-medium leading-none">提供商</label>
              <Input
                value={form.provider || ""}
                onChange={(e) => setForm((f) => ({ ...f, provider: e.target.value }))}
                placeholder="openai / deepseek / custom"
                className="h-9"
              />
            </div>
            <div className="space-y-1.5">
              <label className="text-sm font-medium leading-none">模型名称</label>
              <Input
                value={form.model || ""}
                onChange={(e) => setForm((f) => ({ ...f, model: e.target.value }))}
                placeholder="gpt-4o / deepseek-chat"
                className="h-9"
              />
            </div>
            <div className="space-y-1.5">
              <label className="text-sm font-medium leading-none">API 地址</label>
              <Input
                value={form.base_url || ""}
                onChange={(e) => setForm((f) => ({ ...f, base_url: e.target.value }))}
                placeholder="https://api.openai.com/v1"
                className="h-9"
              />
            </div>
            <div className="space-y-1.5">
              <label className="text-sm font-medium leading-none">API Key</label>
              <Input
                type="password"
                value={form.api_key || ""}
                onChange={(e) => setForm((f) => ({ ...f, api_key: e.target.value }))}
                placeholder={config?.api_key_masked || "输入 API Key"}
                className="h-9"
              />
            </div>
          </>
        )}

        <div className="flex items-center gap-2 pt-2">
          <Button size="sm" onClick={() => updateMutation.mutate(form)} disabled={updateMutation.isPending}>
            {updateMutation.isPending && <Loader2 className="w-4 h-4 animate-spin mr-1" />}
            保存配置
          </Button>
          {form.enabled && (
            <>
              <Button variant="outline" size="sm" onClick={handleTest} disabled={testStatus === "testing"}>
                {testStatus === "testing" && <Loader2 className="w-4 h-4 animate-spin mr-1" />}
                {testStatus === "success" && <CheckCircle className="w-4 h-4 text-green-500 mr-1" />}
                {testStatus === "error" && <XCircle className="w-4 h-4 text-destructive mr-1" />}
                测试连接
              </Button>
              <Button variant="outline" size="sm" onClick={() => setShowPromptsDialog(true)}>
                <MessageSquare className="w-4 h-4 mr-1" />
                修改默认提示词
              </Button>
            </>
          )}
        </div>
      </CardContent>

      {showPromptsDialog && (
        <AIPromptsDialog onClose={() => setShowPromptsDialog(false)} />
      )}
    </Card>
  )
}

/* ============================================================
 * AI Prompts Dialog (Tab-based)
 * ============================================================ */
const PROMPT_LABELS: Record<string, string> = {
  summarize: "聊天总结",
  simulate: "模拟对话",
  sentiment: "情感分析",
  summary: "结构化摘要",
  extract_todos: "待办提取",
  extract_info: "关键信息抽取",
}

const PROMPT_KEYS = Object.keys(PROMPT_LABELS)

function AIPromptsDialog({ onClose }: { onClose: () => void }) {
  const queryClient = useQueryClient()
  const [prompts, setPrompts] = useState<Record<string, string>>({})
  const [defaults, setDefaults] = useState<Record<string, string>>({})
  const [activeTab, setActiveTab] = useState(PROMPT_KEYS[0])

  const { data, isLoading } = useQuery({
    queryKey: ["ai-prompts"],
    queryFn: () => systemApi.getAIPrompts(),
  })

  useEffect(() => {
    if (data) {
      setPrompts(data.prompts || {})
      setDefaults(data.defaults || {})
    }
  }, [data])

  const updateMutation = useMutation({
    mutationFn: (p: Record<string, string>) => systemApi.updateAIPrompts(p),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["ai-prompts"] })
      toast.success("AI 提示词配置已保存")
      onClose()
    },
    onError: (err: Error) => toast.error("保存失败: " + err.message),
  })

  const handleReset = (key: string) => {
    if (defaults[key]) {
      setPrompts((prev) => ({ ...prev, [key]: defaults[key] }))
    }
  }

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/60 backdrop-blur-sm p-4 animate-in fade-in duration-200">
      <div
        className="bg-background border shadow-2xl rounded-2xl w-full max-w-2xl flex flex-col overflow-hidden animate-in zoom-in-95 duration-200 max-h-[85vh]"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="px-6 py-4 border-b flex items-center justify-between">
          <div>
            <h3 className="font-bold text-lg">AI 提示词配置</h3>
            <p className="text-xs text-muted-foreground mt-0.5">
              自定义各 AI 功能的提示词，留空将使用默认值
            </p>
          </div>
          <Button variant="ghost" size="icon" onClick={onClose} className="rounded-full">
            <X className="h-5 w-5" />
          </Button>
        </div>

        {isLoading ? (
          <div className="p-12 flex items-center justify-center">
            <Loader2 className="w-6 h-6 animate-spin text-muted-foreground" />
          </div>
        ) : (
          <>
            {/* Tabs */}
            <div className="flex border-b px-4 overflow-x-auto">
              {PROMPT_KEYS.map((key) => (
                <button
                  key={key}
                  onClick={() => setActiveTab(key)}
                  className={`px-3 py-2.5 text-sm whitespace-nowrap border-b-2 transition-colors ${
                    activeTab === key
                      ? "border-primary text-primary font-medium"
                      : "border-transparent text-muted-foreground hover:text-foreground"
                  }`}
                >
                  {PROMPT_LABELS[key]}
                </button>
              ))}
            </div>

            {/* Content */}
            <div className="flex-1 overflow-y-auto p-6 space-y-3">
              <div className="flex items-center justify-between">
                <p className="text-xs text-muted-foreground">
                  模板变量：模拟对话支持 {"{{target_name}}"} 和 {"{{history}}"}；情感分析支持 {"{{monthly_texts}}"}；关键信息抽取支持 {"{{types_hint}}"}。
                </p>
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-7 px-2 text-xs text-muted-foreground shrink-0"
                  onClick={() => handleReset(activeTab)}
                >
                  <RotateCcw className="w-3 h-3 mr-1" />
                  重置默认
                </Button>
              </div>
              <textarea
                className="w-full min-h-[260px] rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 resize-y"
                value={prompts[activeTab] || ""}
                onChange={(e) =>
                  setPrompts((prev) => ({ ...prev, [activeTab]: e.target.value }))
                }
                placeholder={defaults[activeTab]?.slice(0, 200) + "..."}
              />
            </div>

            {/* Footer */}
            <div className="px-6 py-4 border-t flex items-center justify-end gap-2">
              <Button variant="outline" size="sm" onClick={onClose}>
                取消
              </Button>
              <Button
                size="sm"
                onClick={() => updateMutation.mutate(prompts)}
                disabled={updateMutation.isPending}
              >
                {updateMutation.isPending && <Loader2 className="w-4 h-4 animate-spin mr-1" />}
                保存提示词
              </Button>
            </div>
          </>
        )}
      </div>
      <div className="absolute inset-0 -z-10" onClick={onClose} />
    </div>
  )
}

/* ============================================================
 * Sync Config Section
 * ============================================================ */
function SyncConfigSection() {
  const queryClient = useQueryClient()
  const [enabled, setEnabled] = useState(false)
  const [interval, setInterval] = useState(30)

  const { data: config, isLoading } = useQuery({
    queryKey: ["sync-config"],
    queryFn: () => systemApi.getSyncConfig(),
  })

  useEffect(() => {
    if (config) {
      setEnabled(config.enabled)
      setInterval(config.interval_minutes)
    }
  }, [config])

  const updateMutation = useMutation({
    mutationFn: (data: SyncConfigUpdate) => systemApi.updateSyncConfig(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["sync-config"] })
      toast.success("同步配置已保存")
    },
    onError: (err: Error) => toast.error("保存失败: " + err.message),
  })

  const syncMutation = useMutation({
    mutationFn: () => systemApi.triggerSync(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["sync-config"] })
      toast.success("同步已触发")
    },
    onError: (err: Error) => toast.error("同步失败: " + err.message),
  })

  if (isLoading) {
    return (
      <Card>
        <CardContent className="p-6 flex items-center justify-center">
          <Loader2 className="w-5 h-5 animate-spin text-muted-foreground" />
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base flex items-center gap-2">
          <RefreshCw className="w-4 h-4 text-primary" />
          自动同步
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex items-center justify-between">
          <label className="text-sm font-medium leading-none">启用自动同步</label>
          <Switch checked={enabled} onCheckedChange={setEnabled} />
        </div>

        {enabled && (
          <div className="space-y-1.5">
            <label className="text-sm font-medium leading-none">同步间隔（分钟）</label>
            <Input
              type="number"
              min={5}
              max={1440}
              value={interval}
              onChange={(e) => setInterval(Number(e.target.value))}
              className="h-9 w-32"
            />
            <p className="text-xs text-muted-foreground">最小 5 分钟，最大 1440 分钟（24小时）</p>
          </div>
        )}

        {config?.last_sync_time && (
          <div className="text-xs text-muted-foreground">
            上次同步: {new Date(config.last_sync_time).toLocaleString()}
            {config.last_sync_status && ` (${config.last_sync_status})`}
          </div>
        )}

        <div className="flex items-center gap-2 pt-2">
          <Button
            size="sm"
            onClick={() => updateMutation.mutate({ enabled, interval_minutes: interval })}
            disabled={updateMutation.isPending}
          >
            {updateMutation.isPending && <Loader2 className="w-4 h-4 animate-spin mr-1" />}
            保存配置
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => syncMutation.mutate()}
            disabled={syncMutation.isPending || config?.is_syncing}
          >
            {(syncMutation.isPending || config?.is_syncing) && (
              <Loader2 className="w-4 h-4 animate-spin mr-1" />
            )}
            立即同步
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}

/* ============================================================
 * Password Section
 * ============================================================ */
function PasswordSection() {
  const [oldPwd, setOldPwd] = useState("")
  const [newPwd, setNewPwd] = useState("")
  const [confirmPwd, setConfirmPwd] = useState("")
  const [disablePwd, setDisablePwd] = useState("")

  const { data: status, isLoading, refetch } = useQuery({
    queryKey: ["password-status"],
    queryFn: () => systemApi.getPasswordStatus(),
  })

  const setMutation = useMutation({
    mutationFn: () => systemApi.setPassword(oldPwd, newPwd),
    onSuccess: () => {
      toast.success("密码已设置")
      setOldPwd("")
      setNewPwd("")
      setConfirmPwd("")
      refetch()
    },
    onError: (err: Error) => toast.error("设置失败: " + err.message),
  })

  const disableMutation = useMutation({
    mutationFn: () => systemApi.disablePassword(disablePwd),
    onSuccess: () => {
      toast.success("密码保护已关闭")
      setDisablePwd("")
      refetch()
    },
    onError: (err: Error) => toast.error("关闭失败: " + err.message),
  })

  const handleSetPassword = () => {
    if (newPwd.length < 4) {
      toast.warning("密码至少 4 位")
      return
    }
    if (newPwd !== confirmPwd) {
      toast.warning("两次输入的密码不一致")
      return
    }
    setMutation.mutate()
  }

  if (isLoading) {
    return (
      <Card>
        <CardContent className="p-6 flex items-center justify-center">
          <Loader2 className="w-5 h-5 animate-spin text-muted-foreground" />
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base flex items-center gap-2">
          <Lock className="w-4 h-4 text-primary" />
          密码保护
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <p className="text-xs text-muted-foreground">
          {status?.enabled ? "密码保护已开启，每次打开应用需要输入密码。" : "密码保护未开启。"}
        </p>

        {status?.enabled ? (
          <div className="space-y-4">
            <div className="space-y-1.5">
              <label className="text-sm font-medium leading-none">输入当前密码以关闭保护</label>
              <Input
                type="password"
                value={disablePwd}
                onChange={(e) => setDisablePwd(e.target.value)}
                placeholder="当前密码"
                className="h-9 w-64"
              />
            </div>
            <Button
              variant="destructive"
              size="sm"
              onClick={() => disableMutation.mutate()}
              disabled={!disablePwd || disableMutation.isPending}
            >
              {disableMutation.isPending && <Loader2 className="w-4 h-4 animate-spin mr-1" />}
              关闭密码保护
            </Button>
          </div>
        ) : (
          <div className="space-y-4">
            <div className="space-y-1.5">
              <label className="text-sm font-medium leading-none">设置新密码</label>
              <Input
                type="password"
                value={newPwd}
                onChange={(e) => setNewPwd(e.target.value)}
                placeholder="新密码（至少 4 位）"
                className="h-9 w-64"
              />
            </div>
            <div className="space-y-1.5">
              <label className="text-sm font-medium leading-none">确认密码</label>
              <Input
                type="password"
                value={confirmPwd}
                onChange={(e) => setConfirmPwd(e.target.value)}
                placeholder="再次输入密码"
                className="h-9 w-64"
              />
            </div>
            <Button
              size="sm"
              onClick={handleSetPassword}
              disabled={!newPwd || !confirmPwd || setMutation.isPending}
            >
              {setMutation.isPending && <Loader2 className="w-4 h-4 animate-spin mr-1" />}
              启用密码保护
            </Button>
          </div>
        )}
      </CardContent>
    </Card>
  )
}

/* ============================================================
 * Backup Config Section
 * ============================================================ */
function BackupConfigSection() {
  const queryClient = useQueryClient()
  const [form, setForm] = useState<BackupConfigUpdate>({
    enabled: false,
    interval_hours: 24,
    backup_path: "",
    format: "html",
  })
  const [selectedSessionIds, setSelectedSessionIds] = useState<string[]>([])
  const [backupAll, setBackupAll] = useState(true)
  const [sessionSearch, setSessionSearch] = useState("")

  const { data: config, isLoading } = useQuery({
    queryKey: ["backup-config"],
    queryFn: () => systemApi.getBackupConfig(),
  })

  // Fetch session list for selective backup
  const { data: sessionData } = useQuery({
    queryKey: ["backup-sessions"],
    queryFn: () => sessionApi.getSessions({ limit: 10000 }),
  })

  const filteredSessions = useMemo(() => {
    if (!sessionData?.items) return []
    if (!sessionSearch.trim()) return sessionData.items
    const kw = sessionSearch.trim().toLowerCase()
    return sessionData.items.filter(
      (s) =>
        (s.name || "").toLowerCase().includes(kw) ||
        (s.talkerName || "").toLowerCase().includes(kw) ||
        s.talker.toLowerCase().includes(kw)
    )
  }, [sessionData, sessionSearch])

  useEffect(() => {
    if (config) {
      setForm({
        enabled: config.enabled,
        interval_hours: config.interval_hours,
        backup_path: config.backup_path,
        format: config.format || "html",
      })
    }
  }, [config])

  const updateMutation = useMutation({
    mutationFn: (data: BackupConfigUpdate) => systemApi.updateBackupConfig(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["backup-config"] })
      toast.success("备份配置已保存")
    },
    onError: (err: Error) => toast.error("保存失败: " + err.message),
  })

  const backupMutation = useMutation({
    mutationFn: (sessionIds?: string[]) => systemApi.runBackup(sessionIds),
    onSuccess: () => toast.success("备份任务已启动"),
    onError: (err: Error) => toast.error("备份失败: " + err.message),
  })

  const handleRunBackup = () => {
    if (backupAll) {
      backupMutation.mutate(undefined)
    } else {
      if (selectedSessionIds.length === 0) {
        toast.warning("请至少选择一个会话")
        return
      }
      backupMutation.mutate(selectedSessionIds)
    }
  }

  const handleToggleSession = (id: string) => {
    setSelectedSessionIds((prev) =>
      prev.includes(id) ? prev.filter((s) => s !== id) : [...prev, id]
    )
  }

  const handleSelectAllFiltered = () => {
    const ids = filteredSessions.map((s) => s.talker)
    setSelectedSessionIds((prev) => {
      const set = new Set(prev)
      ids.forEach((id) => set.add(id))
      return Array.from(set)
    })
  }

  const handleDeselectAllFiltered = () => {
    const ids = new Set(filteredSessions.map((s) => s.talker))
    setSelectedSessionIds((prev) => prev.filter((id) => !ids.has(id)))
  }

  if (isLoading) {
    return (
      <Card>
        <CardContent className="p-6 flex items-center justify-center">
          <Loader2 className="w-5 h-5 animate-spin text-muted-foreground" />
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base flex items-center gap-2">
          <HardDrive className="w-4 h-4 text-primary" />
          自动备份
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex items-center justify-between">
          <label className="text-sm font-medium leading-none">启用自动备份</label>
          <Switch
            checked={form.enabled}
            onCheckedChange={(checked) => setForm((f) => ({ ...f, enabled: checked }))}
          />
        </div>

        {form.enabled && (
          <>
            <div className="space-y-1.5">
              <label className="text-sm font-medium leading-none">备份间隔（小时）</label>
              <Input
                type="number"
                min={1}
                value={form.interval_hours}
                onChange={(e) => setForm((f) => ({ ...f, interval_hours: Number(e.target.value) }))}
                className="h-9 w-32"
              />
            </div>
            <div className="space-y-1.5">
              <label className="text-sm font-medium leading-none">备份保存路径</label>
              <Input
                value={form.backup_path}
                onChange={(e) => setForm((f) => ({ ...f, backup_path: e.target.value }))}
                placeholder="/path/to/backups"
                className="h-9"
              />
            </div>
            <div className="space-y-1.5">
              <label className="text-sm font-medium leading-none">备份格式</label>
              <Input
                value={form.format || "html"}
                onChange={(e) => setForm((f) => ({ ...f, format: e.target.value }))}
                placeholder="html / txt / csv"
                className="h-9 w-32"
              />
            </div>
          </>
        )}

        {config?.last_backup_time && (
          <div className="text-xs text-muted-foreground">
            上次备份: {new Date(config.last_backup_time).toLocaleString()}
            {config.last_backup_status && ` (${config.last_backup_status})`}
          </div>
        )}

        {/* Session selection for manual backup */}
        <div className="space-y-2 border rounded-md p-3">
          <label className="text-sm font-medium leading-none">手动备份范围</label>
          <div className="flex items-center gap-3">
            <label className="flex items-center gap-1.5 text-sm cursor-pointer">
              <input
                type="radio"
                name="backup-scope"
                checked={backupAll}
                onChange={() => setBackupAll(true)}
                className="accent-primary"
              />
              备份所有会话
            </label>
            <label className="flex items-center gap-1.5 text-sm cursor-pointer">
              <input
                type="radio"
                name="backup-scope"
                checked={!backupAll}
                onChange={() => setBackupAll(false)}
                className="accent-primary"
              />
              选择会话备份
            </label>
          </div>

          {!backupAll && (
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <div className="relative flex-1">
                  <Search className="absolute left-2 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-muted-foreground" />
                  <Input
                    value={sessionSearch}
                    onChange={(e) => setSessionSearch(e.target.value)}
                    placeholder="搜索会话..."
                    className="h-8 pl-7 text-xs"
                  />
                </div>
                <Button variant="ghost" size="sm" className="h-8 text-xs px-2" onClick={handleSelectAllFiltered}>
                  全选
                </Button>
                <Button variant="ghost" size="sm" className="h-8 text-xs px-2" onClick={handleDeselectAllFiltered}>
                  取消全选
                </Button>
              </div>
              <p className="text-xs text-muted-foreground">
                已选择 {selectedSessionIds.length} 个会话
                {sessionData?.items ? ` / 共 ${sessionData.items.length} 个` : ""}
              </p>
              <div className="max-h-48 overflow-y-auto border rounded-md divide-y">
                {filteredSessions.map((s) => (
                  <label
                    key={s.talker}
                    className="flex items-center gap-2 px-3 py-1.5 hover:bg-muted/50 cursor-pointer text-sm"
                  >
                    <input
                      type="checkbox"
                      checked={selectedSessionIds.includes(s.talker)}
                      onChange={() => handleToggleSession(s.talker)}
                      className="accent-primary"
                    />
                    <span className="truncate">{s.name || s.talkerName || s.talker}</span>
                    {s.type && (
                      <span className="ml-auto text-xs text-muted-foreground shrink-0">
                        {s.type === "group" ? "群聊" : s.type === "private" ? "私聊" : s.type === "official" ? "公众号" : ""}
                      </span>
                    )}
                  </label>
                ))}
                {filteredSessions.length === 0 && (
                  <div className="px-3 py-4 text-center text-xs text-muted-foreground">
                    {sessionSearch ? "未找到匹配的会话" : "暂无会话数据"}
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
        <div className="flex items-center gap-2 pt-2">
          <Button
            size="sm"
            onClick={() => updateMutation.mutate(form)}
            disabled={updateMutation.isPending}
          >
            {updateMutation.isPending && <Loader2 className="w-4 h-4 animate-spin mr-1" />}
            保存配置
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={handleRunBackup}
            disabled={backupMutation.isPending}
          >
            {backupMutation.isPending && <Loader2 className="w-4 h-4 animate-spin mr-1" />}
            {backupAll ? "立即备份全部" : `立即备份 (${selectedSessionIds.length})`}
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}

/* ============================================================
 * TTS Voice-to-Text Config Section
 * ============================================================ */
function TTSConfigSection() {
  const queryClient = useQueryClient()
  const [form, setForm] = useState<TTSConfigUpdate>({
    enabled: false,
    provider: "openai",
    base_url: "",
    api_key: "",
    model: "whisper-1",
  })

  const { data: config, isLoading } = useQuery({
    queryKey: ["tts-config"],
    queryFn: () => systemApi.getTTSConfig(),
  })

  useEffect(() => {
    if (config) {
      setForm({
        enabled: config.enabled,
        provider: config.provider || "openai",
        base_url: config.base_url || "",
        api_key: "",
        model: config.model || "whisper-1",
      })
    }
  }, [config])

  const updateMutation = useMutation({
    mutationFn: (data: TTSConfigUpdate) => systemApi.updateTTSConfig(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["tts-config"] })
      toast.success("语音转文字配置已保存")
    },
    onError: (err: Error) => toast.error("保存失败: " + err.message),
  })

  if (isLoading) {
    return (
      <Card>
        <CardContent className="p-6 flex items-center justify-center">
          <Loader2 className="w-5 h-5 animate-spin text-muted-foreground" />
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base flex items-center gap-2">
          <Mic className="w-4 h-4 text-primary" />
          语音转文字 (Whisper)
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex items-center justify-between">
          <label className="text-sm font-medium leading-none">启用语音转文字</label>
          <Switch
            checked={form.enabled}
            onCheckedChange={(checked) => setForm((f) => ({ ...f, enabled: checked }))}
          />
        </div>

        {form.enabled && (
          <>
            <div className="space-y-1.5">
              <label className="text-sm font-medium leading-none">提供商</label>
              <Input
                value={form.provider || ""}
                onChange={(e) => setForm((f) => ({ ...f, provider: e.target.value }))}
                placeholder="openai"
                className="h-9"
              />
            </div>
            <div className="space-y-1.5">
              <label className="text-sm font-medium leading-none">模型名称</label>
              <Input
                value={form.model || ""}
                onChange={(e) => setForm((f) => ({ ...f, model: e.target.value }))}
                placeholder="whisper-1"
                className="h-9"
              />
            </div>
            <div className="space-y-1.5">
              <label className="text-sm font-medium leading-none">API 地址</label>
              <Input
                value={form.base_url || ""}
                onChange={(e) => setForm((f) => ({ ...f, base_url: e.target.value }))}
                placeholder="https://api.openai.com/v1"
                className="h-9"
              />
            </div>
            <div className="space-y-1.5">
              <label className="text-sm font-medium leading-none">API Key</label>
              <Input
                type="password"
                value={form.api_key || ""}
                onChange={(e) => setForm((f) => ({ ...f, api_key: e.target.value }))}
                placeholder={config?.api_key_masked || "输入 API Key"}
                className="h-9"
              />
            </div>
            <p className="text-xs text-muted-foreground">
              配置后，语音消息旁会出现"转文字"按钮，点击即可将语音转为文字。支持 OpenAI Whisper 兼容接口。
            </p>
          </>
        )}

        <div className="flex items-center gap-2 pt-2">
          <Button
            size="sm"
            onClick={() => updateMutation.mutate(form)}
            disabled={updateMutation.isPending}
          >
            {updateMutation.isPending && <Loader2 className="w-4 h-4 animate-spin mr-1" />}
            保存配置
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}

/* ============================================================
 * Main Settings View
 * ============================================================ */
export default function SettingsView() {
  return (
    <ScrollArea className="h-full">
      <div className="max-w-3xl mx-auto p-6 space-y-6 pb-20">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">设置</h2>
          <p className="text-sm text-muted-foreground mt-1">管理应用配置</p>
        </div>

        <AIConfigSection />
        <TTSConfigSection />
        <SyncConfigSection />
        <PasswordSection />
        <BackupConfigSection />
      </div>
    </ScrollArea>
  )
}