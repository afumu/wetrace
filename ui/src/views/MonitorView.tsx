import { useState, useEffect, useMemo } from "react"
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { monitorApi } from "@/api/monitor"
import { sessionApi } from "@/api"
import { toast } from "sonner"
import type { MonitorConfig, MonitorConfigCreate, FeishuConfigUpdate } from "@/api/monitor"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Switch } from "@/components/ui/switch"
import {
  Shield,
  Plus,
  Trash2,
  Pencil,
  Loader2,
  CheckCircle,
  XCircle,
  Send,
  X,
  AlertTriangle,
  Search,
} from "lucide-react"

type EditingConfig = MonitorConfigCreate & { id?: number }

const emptyForm: EditingConfig = {
  name: "",
  type: "keyword",
  prompt: "",
  keywords: [],
  platform: "webhook",
  webhook_url: "",
  feishu_url: "",
  enabled: true,
  session_ids: [],
  interval_minutes: 5,
}

/* ============================================================
 * Config Form Dialog (inline overlay)
 * ============================================================ */
function ConfigFormDialog({
  initial,
  onSave,
  onCancel,
  saving,
}: {
  initial: EditingConfig
  onSave: (data: EditingConfig) => void
  onCancel: () => void
  saving: boolean
}) {
  const [form, setForm] = useState<EditingConfig>(initial)
  const [keywordInput, setKeywordInput] = useState("")
  const [sessionSearch, setSessionSearch] = useState("")

  const { data: sessionData } = useQuery({
    queryKey: ["monitor-sessions"],
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

  const handleToggleSession = (id: string) => {
    setForm((f) => {
      const ids = f.session_ids || []
      return {
        ...f,
        session_ids: ids.includes(id) ? ids.filter((s) => s !== id) : [...ids, id],
      }
    })
  }

  const addKeyword = () => {
    const kw = keywordInput.trim()
    if (kw && !form.keywords?.includes(kw)) {
      setForm((f) => ({ ...f, keywords: [...(f.keywords || []), kw] }))
    }
    setKeywordInput("")
  }

  const removeKeyword = (kw: string) => {
    setForm((f) => ({ ...f, keywords: (f.keywords || []).filter((k) => k !== kw) }))
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base flex items-center gap-2">
          <Shield className="w-4 h-4 text-primary" />
          {initial.id ? "编辑监控配置" : "新建监控配置"}
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Name */}
        <div className="space-y-1.5">
          <label className="text-sm font-medium leading-none">配置名称</label>
          <Input
            value={form.name}
            onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
            placeholder="例如：敏感词监控"
            className="h-9"
          />
        </div>

        {/* Type */}
        <div className="space-y-1.5">
          <label className="text-sm font-medium leading-none">监控类型</label>
          <div className="flex gap-2">
            <Button
              size="sm"
              variant={form.type === "keyword" ? "default" : "outline"}
              onClick={() => setForm((f) => ({ ...f, type: "keyword" }))}
            >
              关键词匹配
            </Button>
            <Button
              size="sm"
              variant={form.type === "ai" ? "default" : "outline"}
              onClick={() => setForm((f) => ({ ...f, type: "ai" }))}
            >
              AI 智能监控
            </Button>
          </div>
        </div>

        {/* Keywords (keyword type) */}
        {form.type === "keyword" && (
          <div className="space-y-1.5">
            <label className="text-sm font-medium leading-none">敏感词列表</label>
            <div className="flex gap-2">
              <Input
                value={keywordInput}
                onChange={(e) => setKeywordInput(e.target.value)}
                placeholder="输入敏感词后回车添加"
                className="h-9"
                onKeyDown={(e) => {
                  if (e.key === "Enter") {
                    e.preventDefault()
                    addKeyword()
                  }
                }}
              />
              <Button size="sm" variant="outline" onClick={addKeyword}>
                添加
              </Button>
            </div>
            {(form.keywords || []).length > 0 && (
              <div className="flex flex-wrap gap-1.5 pt-1">
                {(form.keywords || []).map((kw) => (
                  <span
                    key={kw}
                    className="inline-flex items-center gap-1 px-2 py-0.5 rounded-md bg-muted text-sm"
                  >
                    {kw}
                    <button
                      onClick={() => removeKeyword(kw)}
                      className="text-muted-foreground hover:text-destructive"
                    >
                      <X className="w-3 h-3" />
                    </button>
                  </span>
                ))}
              </div>
            )}
          </div>
        )}

        {/* Prompt (ai type) */}
        {form.type === "ai" && (
          <div className="space-y-1.5">
            <label className="text-sm font-medium leading-none">AI 提示词</label>
            <textarea
              value={form.prompt || ""}
              onChange={(e) => setForm((f) => ({ ...f, prompt: e.target.value }))}
              placeholder="例如：判断以下消息是否包含负面情绪或投诉内容"
              className="w-full h-24 rounded-md border border-input bg-background px-3 py-2 text-sm resize-none focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            />
          </div>
        )}

        {/* Platform */}
        <div className="space-y-1.5">
          <label className="text-sm font-medium leading-none">推送平台</label>
          <div className="flex gap-2">
            <Button
              size="sm"
              variant={form.platform === "webhook" ? "default" : "outline"}
              onClick={() => setForm((f) => ({ ...f, platform: "webhook" }))}
            >
              Webhook
            </Button>
            <Button
              size="sm"
              variant={form.platform === "feishu" ? "default" : "outline"}
              onClick={() => setForm((f) => ({ ...f, platform: "feishu" }))}
            >
              飞书
            </Button>
          </div>
        </div>

        {/* Webhook URL */}
        {form.platform === "webhook" && (
          <div className="space-y-1.5">
            <label className="text-sm font-medium leading-none">Webhook URL</label>
            <Input
              value={form.webhook_url || ""}
              onChange={(e) => setForm((f) => ({ ...f, webhook_url: e.target.value }))}
              placeholder="https://example.com/webhook"
              className="h-9"
            />
          </div>
        )}

        {/* Feishu URL */}
        {form.platform === "feishu" && (
          <div className="space-y-1.5">
            <label className="text-sm font-medium leading-none">飞书机器人 Webhook URL</label>
            <Input
              value={form.feishu_url || ""}
              onChange={(e) => setForm((f) => ({ ...f, feishu_url: e.target.value }))}
              placeholder="https://open.feishu.cn/open-apis/bot/v2/hook/xxx"
              className="h-9"
            />
          </div>
        )}

        {/* Interval */}
        <div className="space-y-1.5">
          <label className="text-sm font-medium leading-none">检查间隔（分钟）</label>
          <Input
            type="number"
            min={1}
            max={1440}
            value={form.interval_minutes || 5}
            onChange={(e) => setForm((f) => ({ ...f, interval_minutes: Number(e.target.value) }))}
            className="h-9 w-32"
          />
          <p className="text-xs text-muted-foreground">每隔多少分钟检查一次新消息，最小1分钟</p>
        </div>

        {/* Session Selector */}
        <div className="space-y-2 border rounded-md p-3">
          <label className="text-sm font-medium leading-none">监控会话</label>
          <p className="text-xs text-muted-foreground">
            选择要监控的会话，未选择则不监控任何会话。已选 {(form.session_ids || []).length} 个。
          </p>
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
          </div>
          <div className="max-h-40 overflow-y-auto border rounded-md divide-y">
            {filteredSessions.map((s) => (
              <label
                key={s.talker}
                className="flex items-center gap-2 px-3 py-1.5 hover:bg-muted/50 cursor-pointer text-sm"
              >
                <input
                  type="checkbox"
                  checked={(form.session_ids || []).includes(s.talker)}
                  onChange={() => handleToggleSession(s.talker)}
                  className="accent-primary"
                />
                <span className="truncate">{s.name || s.talkerName || s.talker}</span>
              </label>
            ))}
            {filteredSessions.length === 0 && (
              <div className="px-3 py-4 text-center text-xs text-muted-foreground">
                {sessionSearch ? "未找到匹配的会话" : "暂无会话数据"}
              </div>
            )}
          </div>
        </div>

        {/* Enabled */}
        <div className="flex items-center justify-between">
          <label className="text-sm font-medium leading-none">启用</label>
          <Switch
            checked={form.enabled}
            onCheckedChange={(checked) => setForm((f) => ({ ...f, enabled: checked }))}
          />
        </div>

        {/* Actions */}
        <div className="flex items-center gap-2 pt-2">
          <Button size="sm" onClick={() => onSave(form)} disabled={saving}>
            {saving && <Loader2 className="w-4 h-4 animate-spin mr-1" />}
            保存
          </Button>
          <Button size="sm" variant="outline" onClick={onCancel}>
            取消
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}

/* ============================================================
 * Feishu Config Section
 * ============================================================ */
function FeishuConfigSection() {
  const queryClient = useQueryClient()
  const [form, setForm] = useState<FeishuConfigUpdate>({
    bot_webhook: "",
    sign_secret: "",
    enabled: false,
    app_id: "",
    app_secret: "",
    app_token: "",
    table_id: "",
    push_type: "bot",
  })
  const [testStatus, setTestStatus] = useState<
    "idle" | "testing" | "success" | "error"
  >("idle")
  const [bitableTestStatus, setBitableTestStatus] = useState<
    "idle" | "testing" | "success" | "error"
  >("idle")

  const { data: config, isLoading } = useQuery({
    queryKey: ["feishu-config"],
    queryFn: () => monitorApi.getFeishuConfig(),
  })

  useEffect(() => {
    if (config) {
      setForm({
        bot_webhook: config.bot_webhook || "",
        sign_secret: config.sign_secret || "",
        enabled: config.enabled,
        app_id: config.app_id || "",
        app_secret: config.app_secret || "",
        app_token: config.app_token || "",
        table_id: config.table_id || "",
        push_type: config.push_type || "bot",
      })
    }
  }, [config])

  const updateMutation = useMutation({
    mutationFn: (data: FeishuConfigUpdate) =>
      monitorApi.updateFeishuConfig(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["feishu-config"] })
      toast.success("飞书配置已保存")
    },
    onError: (err: Error) => toast.error("保存失败: " + err.message),
  })

  const handleTest = async () => {
    setTestStatus("testing")
    try {
      await monitorApi.testFeishu()
      setTestStatus("success")
    } catch {
      setTestStatus("error")
    }
  }

  const handleBitableTest = async () => {
    setBitableTestStatus("testing")
    try {
      await monitorApi.testFeishuBitable()
      setBitableTestStatus("success")
    } catch {
      setBitableTestStatus("error")
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
          <Send className="w-4 h-4 text-primary" />
          飞书平台配置
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex items-center justify-between">
          <label className="text-sm font-medium leading-none">
            启用飞书推送
          </label>
          <Switch
            checked={form.enabled}
            onCheckedChange={(v) => setForm((f) => ({ ...f, enabled: v }))}
          />
        </div>

        {form.enabled && (
          <>
            {/* 推送方式选择 */}
            <div className="space-y-1.5">
              <label className="text-sm font-medium leading-none">推送方式</label>
              <div className="flex gap-2">
                {(["bot", "bitable", "both"] as const).map((t) => (
                  <Button
                    key={t}
                    size="sm"
                    variant={form.push_type === t ? "default" : "outline"}
                    onClick={() => setForm((f) => ({ ...f, push_type: t }))}
                  >
                    {t === "bot" ? "机器人" : t === "bitable" ? "多维表格" : "两者都用"}
                  </Button>
                ))}
              </div>
            </div>

            {/* 机器人配置 */}
            {(form.push_type === "bot" || form.push_type === "both") && (
              <div className="space-y-3 border rounded-md p-3">
                <p className="text-xs font-medium text-muted-foreground">机器人推送配置</p>
                <div className="space-y-1.5">
                  <label className="text-sm font-medium leading-none">
                    机器人 Webhook URL
                  </label>
                  <Input
                    value={form.bot_webhook}
                    onChange={(e) =>
                      setForm((f) => ({ ...f, bot_webhook: e.target.value }))
                    }
                    placeholder="https://open.feishu.cn/open-apis/bot/v2/hook/xxx"
                    className="h-9"
                  />
                </div>
                <div className="space-y-1.5">
                  <label className="text-sm font-medium leading-none">
                    签名密钥（可选）
                  </label>
                  <Input
                    type="password"
                    value={form.sign_secret || ""}
                    onChange={(e) =>
                      setForm((f) => ({ ...f, sign_secret: e.target.value }))
                    }
                    placeholder="签名校验密钥"
                    className="h-9"
                  />
                </div>
              </div>
            )}

            {/* 多维表格配置 */}
            {(form.push_type === "bitable" || form.push_type === "both") && (
              <div className="space-y-3 border rounded-md p-3">
                <p className="text-xs font-medium text-muted-foreground">多维表格推送配置</p>
                <div className="space-y-1.5">
                  <label className="text-sm font-medium leading-none">App ID</label>
                  <Input
                    value={form.app_id || ""}
                    onChange={(e) => setForm((f) => ({ ...f, app_id: e.target.value }))}
                    placeholder="飞书应用 App ID"
                    className="h-9"
                  />
                </div>
                <div className="space-y-1.5">
                  <label className="text-sm font-medium leading-none">App Secret</label>
                  <Input
                    type="password"
                    value={form.app_secret || ""}
                    onChange={(e) => setForm((f) => ({ ...f, app_secret: e.target.value }))}
                    placeholder="飞书应用 App Secret"
                    className="h-9"
                  />
                </div>
                <div className="space-y-1.5">
                  <label className="text-sm font-medium leading-none">多维表格 App Token</label>
                  <Input
                    value={form.app_token || ""}
                    onChange={(e) => setForm((f) => ({ ...f, app_token: e.target.value }))}
                    placeholder="多维表格链接中的 app_token"
                    className="h-9"
                  />
                </div>
                <div className="space-y-1.5">
                  <label className="text-sm font-medium leading-none">数据表 Table ID</label>
                  <Input
                    value={form.table_id || ""}
                    onChange={(e) => setForm((f) => ({ ...f, table_id: e.target.value }))}
                    placeholder="数据表 ID"
                    className="h-9"
                  />
                </div>
                <p className="text-xs text-muted-foreground">
                  需要在飞书开放平台创建应用并授权多维表格权限。表格需包含字段：发送人、会话、消息内容、触发规则、消息时间、告警时间。
                </p>
              </div>
            )}
          </>
        )}

        <div className="flex items-center gap-2 pt-2 flex-wrap">
          <Button
            size="sm"
            onClick={() => updateMutation.mutate(form)}
            disabled={updateMutation.isPending}
          >
            {updateMutation.isPending && (
              <Loader2 className="w-4 h-4 animate-spin mr-1" />
            )}
            保存配置
          </Button>
          {form.enabled && (form.push_type === "bot" || form.push_type === "both") && (
            <Button
              variant="outline"
              size="sm"
              onClick={handleTest}
              disabled={testStatus === "testing"}
            >
              {testStatus === "testing" && (
                <Loader2 className="w-4 h-4 animate-spin mr-1" />
              )}
              {testStatus === "success" && (
                <CheckCircle className="w-4 h-4 text-green-500 mr-1" />
              )}
              {testStatus === "error" && (
                <XCircle className="w-4 h-4 text-destructive mr-1" />
              )}
              测试机器人
            </Button>
          )}
          {form.enabled && (form.push_type === "bitable" || form.push_type === "both") && (
            <Button
              variant="outline"
              size="sm"
              onClick={handleBitableTest}
              disabled={bitableTestStatus === "testing"}
            >
              {bitableTestStatus === "testing" && (
                <Loader2 className="w-4 h-4 animate-spin mr-1" />
              )}
              {bitableTestStatus === "success" && (
                <CheckCircle className="w-4 h-4 text-green-500 mr-1" />
              )}
              {bitableTestStatus === "error" && (
                <XCircle className="w-4 h-4 text-destructive mr-1" />
              )}
              测试多维表格
            </Button>
          )}
        </div>
      </CardContent>
    </Card>
  )
}

/* ============================================================
 * Config List Item
 * ============================================================ */
function ConfigListItem({
  cfg,
  testingId,
  testResult,
  onEdit,
  onDelete,
  onTest,
}: {
  cfg: MonitorConfig
  testingId: number | null
  testResult?: "success" | "error"
  onEdit: () => void
  onDelete: () => void
  onTest: () => void
}) {
  const isTesting = testingId === cfg.id

  return (
    <div className="border rounded-lg p-4 flex items-start justify-between gap-4">
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2 mb-1">
          <span className="text-sm font-medium truncate">{cfg.name}</span>
          <span
            className={
              cfg.enabled
                ? "text-[10px] px-1.5 py-0.5 rounded bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400"
                : "text-[10px] px-1.5 py-0.5 rounded bg-muted text-muted-foreground"
            }
          >
            {cfg.enabled ? "已启用" : "已禁用"}
          </span>
        </div>
        <div className="flex flex-wrap gap-2 text-xs text-muted-foreground">
          <span>类型: {cfg.type === "keyword" ? "关键词" : "AI"}</span>
          <span>平台: {cfg.platform === "feishu" ? "飞书" : "Webhook"}</span>
          {cfg.type === "keyword" && cfg.keywords?.length > 0 && (
            <span>关键词: {cfg.keywords.slice(0, 3).join(", ")}{cfg.keywords.length > 3 ? "..." : ""}</span>
          )}
        </div>
      </div>
      <div className="flex items-center gap-1 flex-shrink-0">
        <Button variant="ghost" size="icon" onClick={onTest} disabled={isTesting} title="测试推送">
          {isTesting ? (
            <Loader2 className="w-4 h-4 animate-spin" />
          ) : testResult === "success" ? (
            <CheckCircle className="w-4 h-4 text-green-500" />
          ) : testResult === "error" ? (
            <XCircle className="w-4 h-4 text-destructive" />
          ) : (
            <Send className="w-4 h-4" />
          )}
        </Button>
        <Button variant="ghost" size="icon" onClick={onEdit} title="编辑">
          <Pencil className="w-4 h-4" />
        </Button>
        <Button variant="ghost" size="icon" onClick={onDelete} title="删除">
          <Trash2 className="w-4 h-4 text-destructive" />
        </Button>
      </div>
    </div>
  )
}

/* ============================================================
 * Main MonitorView
 * ============================================================ */
export default function MonitorView() {
  const queryClient = useQueryClient()
  const [editing, setEditing] = useState<EditingConfig | null>(null)
  const [testingId, setTestingId] = useState<number | null>(null)
  const [testResult, setTestResult] = useState<Record<number, "success" | "error">>({})

  const { data: configs, isLoading } = useQuery({
    queryKey: ["monitor-configs"],
    queryFn: () => monitorApi.getConfigs(),
  })

  const createMutation = useMutation({
    mutationFn: (data: MonitorConfigCreate) => monitorApi.createConfig(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["monitor-configs"] })
      setEditing(null)
    },
    onError: (err: Error) => toast.error("创建失败: " + err.message),
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: MonitorConfigCreate }) =>
      monitorApi.updateConfig(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["monitor-configs"] })
      setEditing(null)
    },
    onError: (err: Error) => toast.error("更新失败: " + err.message),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: number) => monitorApi.deleteConfig(id),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["monitor-configs"] }),
    onError: (err: Error) => toast.error("删除失败: " + err.message),
  })

  const handleSave = (form: EditingConfig) => {
    const { id, ...data } = form
    if (id) {
      updateMutation.mutate({ id, data })
    } else {
      createMutation.mutate(data)
    }
  }

  const handleDelete = (id: number) => {
    deleteMutation.mutate(id)
  }

  const handleTestPush = async (cfg: MonitorConfig) => {
    const url = cfg.platform === "feishu" ? cfg.feishu_url : cfg.webhook_url
    if (!url) {
      toast.warning("推送地址为空，无法测试")
      return
    }
    setTestingId(cfg.id)
    try {
      await monitorApi.testPush({ url })
      setTestResult((r) => ({ ...r, [cfg.id]: "success" }))
    } catch {
      setTestResult((r) => ({ ...r, [cfg.id]: "error" }))
    } finally {
      setTestingId(null)
    }
  }

  const handleEdit = (cfg: MonitorConfig) => {
    setEditing({
      id: cfg.id,
      name: cfg.name,
      type: cfg.type,
      prompt: cfg.prompt,
      keywords: cfg.keywords || [],
      platform: cfg.platform,
      webhook_url: cfg.webhook_url,
      feishu_url: cfg.feishu_url,
      enabled: cfg.enabled,
    })
  }

  return (
    <ScrollArea className="h-full">
      <div className="max-w-3xl mx-auto p-6 space-y-6 pb-20">
        {/* Page header */}
        <div>
          <h2 className="text-2xl font-bold tracking-tight">监控配置</h2>
          <p className="text-sm text-muted-foreground mt-1">
            管理敏感词监控规则和推送通知
          </p>
        </div>

        {/* Config form (editing / creating) */}
        {editing && (
          <ConfigFormDialog
            initial={editing}
            onSave={handleSave}
            onCancel={() => setEditing(null)}
            saving={createMutation.isPending || updateMutation.isPending}
          />
        )}

        {/* Config list */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle className="text-base flex items-center gap-2">
              <Shield className="w-4 h-4 text-primary" />
              监控规则列表
            </CardTitle>
            {!editing && (
              <Button
                size="sm"
                onClick={() => setEditing({ ...emptyForm })}
              >
                <Plus className="w-4 h-4 mr-1" />
                新建
              </Button>
            )}
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <div className="flex items-center justify-center py-8">
                <Loader2 className="w-5 h-5 animate-spin text-muted-foreground" />
              </div>
            ) : !configs || configs.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-12 gap-4">
                <div className="w-16 h-16 bg-muted rounded-full flex items-center justify-center">
                  <AlertTriangle className="w-8 h-8 text-muted-foreground/30" />
                </div>
                <p className="text-muted-foreground text-sm font-medium">
                  暂无监控配置
                </p>
              </div>
            ) : (
              <div className="space-y-3">
                {configs.map((cfg) => (
                  <ConfigListItem
                    key={cfg.id}
                    cfg={cfg}
                    testingId={testingId}
                    testResult={testResult[cfg.id]}
                    onEdit={() => handleEdit(cfg)}
                    onDelete={() => handleDelete(cfg.id)}
                    onTest={() => handleTestPush(cfg)}
                  />
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Feishu config */}
        <FeishuConfigSection />
      </div>
    </ScrollArea>
  )
}
