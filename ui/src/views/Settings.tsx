import { useState, useEffect } from "react"
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { systemApi } from "@/api"
import { toast } from "sonner"
import type {
  AIConfigUpdate,
  SyncConfigUpdate,
  BackupConfigUpdate,
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
            <Button variant="outline" size="sm" onClick={handleTest} disabled={testStatus === "testing"}>
              {testStatus === "testing" && <Loader2 className="w-4 h-4 animate-spin mr-1" />}
              {testStatus === "success" && <CheckCircle className="w-4 h-4 text-green-500 mr-1" />}
              {testStatus === "error" && <XCircle className="w-4 h-4 text-destructive mr-1" />}
              测试连接
            </Button>
          )}
        </div>
      </CardContent>
    </Card>
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

  const { data: config, isLoading } = useQuery({
    queryKey: ["backup-config"],
    queryFn: () => systemApi.getBackupConfig(),
  })

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
    mutationFn: () => systemApi.runBackup(),
    onSuccess: () => toast.success("备份任务已启动"),
    onError: (err: Error) => toast.error("备份失败: " + err.message),
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
            onClick={() => backupMutation.mutate()}
            disabled={backupMutation.isPending}
          >
            {backupMutation.isPending && <Loader2 className="w-4 h-4 animate-spin mr-1" />}
            立即备份
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
        <SyncConfigSection />
        <PasswordSection />
        <BackupConfigSection />
      </div>
    </ScrollArea>
  )
}