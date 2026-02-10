import { useState } from "react"
import { createPortal } from "react-dom"
import { Lock } from "lucide-react"
import { Button } from "./ui/button"
import { systemApi } from "@/api"

interface PasswordDialogProps {
  onUnlocked: () => void
}

export function PasswordDialog({ onUnlocked }: PasswordDialogProps) {
  const [password, setPassword] = useState("")
  const [error, setError] = useState("")
  const [submitting, setSubmitting] = useState(false)

  const handleVerify = async () => {
    if (!password.trim()) {
      setError("请输入密码")
      return
    }
    setError("")
    setSubmitting(true)
    try {
      const data = await systemApi.verifyPassword(password) as any
      if (data?.token) {
        localStorage.setItem("auth_token", data.token)
      }
      onUnlocked()
    } catch (err: any) {
      setError(err?.message || "密码错误")
    } finally {
      setSubmitting(false)
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") {
      handleVerify()
    }
  }

  return createPortal(
    <div className="fixed inset-0 z-[200] flex items-center justify-center bg-black/60 backdrop-blur-sm">
      <div className="bg-background border shadow-2xl rounded-xl p-6 w-[400px] animate-in zoom-in-95 duration-200">
        <div className="flex items-center gap-3 mb-5 pb-4 border-b">
          <div className="w-10 h-10 rounded-full bg-primary/10 flex items-center justify-center">
            <Lock className="w-5 h-5 text-primary" />
          </div>
          <div>
            <h3 className="text-lg font-bold">密码验证</h3>
            <p className="text-xs text-muted-foreground">请输入密码以继续使用</p>
          </div>
        </div>

        <div className="space-y-4">
          <div>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="请输入密码"
              autoFocus
              className="w-full px-3 py-2 rounded-md border border-border bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary/50"
            />
            {error && (
              <p className="text-xs text-destructive mt-1.5">{error}</p>
            )}
          </div>

          <Button
            className="w-full"
            disabled={submitting}
            onClick={handleVerify}
          >
            {submitting ? "验证中..." : "解锁"}
          </Button>
        </div>
      </div>
    </div>,
    document.body
  )
}
