import { useState } from "react"
import { createPortal } from "react-dom"
import { ShieldCheck, AlertTriangle } from "lucide-react"
import { Button } from "./ui/button"
import { systemApi } from "@/api"

const COMPLIANCE_VERSION = "1.0"

interface ComplianceDialogProps {
  onAgreed: () => void
}

export function ComplianceDialog({ onAgreed }: ComplianceDialogProps) {
  const [checked, setChecked] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  const handleAgree = async () => {
    if (!checked) return
    setSubmitting(true)
    try {
      await systemApi.agreeCompliance(COMPLIANCE_VERSION)
      onAgreed()
    } catch {
      onAgreed()
    } finally {
      setSubmitting(false)
    }
  }

  return createPortal(
    <div className="fixed inset-0 z-[200] flex items-center justify-center bg-black/60 backdrop-blur-sm">
      <div className="bg-background border shadow-2xl rounded-xl p-6 w-[520px] max-h-[85vh] overflow-y-auto animate-in zoom-in-95 duration-200">
        <ComplianceHeader />
        <ComplianceBody />
        <ComplianceFooter
          checked={checked}
          onCheckedChange={setChecked}
          onAgree={handleAgree}
          submitting={submitting}
        />
      </div>
    </div>,
    document.body
  )
}

function ComplianceHeader() {
  return (
    <div className="flex items-center gap-3 mb-5 pb-4 border-b">
      <div className="w-10 h-10 rounded-full bg-primary/10 flex items-center justify-center">
        <ShieldCheck className="w-5 h-5 text-primary" />
      </div>
      <div>
        <h3 className="text-lg font-bold">使用须知与合规声明</h3>
        <p className="text-xs text-muted-foreground">请仔细阅读以下内容后再使用本软件</p>
      </div>
    </div>
  )
}

function ComplianceBody() {
  return (
    <div className="space-y-4 text-sm leading-relaxed text-foreground/80">
      <div className="flex items-start gap-2 p-3 rounded-lg bg-warning/10 border border-warning/20">
        <AlertTriangle className="w-4 h-4 text-warning mt-0.5 shrink-0" />
        <p className="text-xs text-warning">
          本软件仅供个人合法用途，使用前请确保已获得相关方的知情同意。
        </p>
      </div>

      <div className="space-y-3">
        <ComplianceSection
          title="1. 数据本地处理"
          content="所有聊天数据仅在您的本地设备上处理和存储，不会上传至任何第三方服务器。AI分析功能会将部分文本发送至您配置的AI服务提供商。"
        />
        <ComplianceSection
          title="2. 授权使用"
          content="在查看、分析或导出他人的聊天记录前，您应当确保已获得聊天对方的知情同意或具有合法授权。未经授权查看他人隐私信息可能违反相关法律法规。"
        />
        <ComplianceSection
          title="3. 合法用途"
          content="本软件不得用于非法监控、骚扰、侵犯他人隐私或其他违法目的。导出的聊天记录如用于法律取证，请确保符合当地法律对电子证据的要求。"
        />
        <ComplianceSection
          title="4. 免责声明"
          content="开发者不对因使用本软件而产生的任何法律纠纷或损失承担责任。用户应自行承担使用本软件的全部法律风险。"
        />
      </div>
    </div>
  )
}

function ComplianceSection({ title, content }: { title: string; content: string }) {
  return (
    <div>
      <h4 className="font-medium text-foreground mb-1">{title}</h4>
      <p className="text-muted-foreground text-xs leading-relaxed">{content}</p>
    </div>
  )
}

function ComplianceFooter({
  checked,
  onCheckedChange,
  onAgree,
  submitting,
}: {
  checked: boolean
  onCheckedChange: (v: boolean) => void
  onAgree: () => void
  submitting: boolean
}) {
  return (
    <div className="mt-6 pt-4 border-t space-y-4">
      <label className="flex items-start gap-2 cursor-pointer select-none">
        <input
          type="checkbox"
          checked={checked}
          onChange={(e) => onCheckedChange(e.target.checked)}
          className="mt-0.5 rounded border-border"
        />
        <span className="text-xs text-muted-foreground leading-relaxed">
          我已阅读并理解上述声明，确认在合法授权的前提下使用本软件，并自行承担使用风险。
        </span>
      </label>
      <Button
        className="w-full"
        disabled={!checked || submitting}
        onClick={onAgree}
      >
        {submitting ? "提交中..." : "同意并继续"}
      </Button>
    </div>
  )
}
