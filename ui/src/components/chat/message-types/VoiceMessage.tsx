import { useState, useRef, useEffect } from "react"
import { mediaApi } from "@/api/media"
import { cn } from "@/lib/utils"
import { Loader2, Type } from "lucide-react"

interface VoiceMessageProps {
  id: string
  isSelf?: boolean
  duration?: number
}

export function VoiceMessage({ id, isSelf, duration }: VoiceMessageProps) {
  const [isPlaying, setIsPlaying] = useState(false)
  const [animationStep, setAnimationStep] = useState(3)
  const audioRef = useRef<HTMLAudioElement | null>(null)
  const animationIntervalRef = useRef<number | null>(null)
  const [transcribeText, setTranscribeText] = useState<string | null>(null)
  const [isTranscribing, setIsTranscribing] = useState(false)

  const voiceUrl = mediaApi.getVoiceUrl(id)
  
  // Clean up duration for display
  // If duration is 0 or undefined, default to 1 for better UX? 
  // Or keep 0 if that's the data. The user said "shows 0 seconds", so data is likely 0.
  // We will display whatever is passed, but format it nicely.
  const displayDuration = duration ? Math.round(duration / 1000) : 0

  useEffect(() => {
    const audio = new Audio(voiceUrl)
    audio.onended = () => {
      setIsPlaying(false)
      stopAnimation()
    }
    audio.onpause = () => {
      setIsPlaying(false)
      stopAnimation()
    }
    audio.onplay = () => {
      setIsPlaying(true)
      startAnimation()
    }
    audio.onerror = () => {
        setIsPlaying(false)
        stopAnimation()
    }
    
    audioRef.current = audio

    return () => {
      audio.pause()
      audioRef.current = null
      stopAnimation()
    }
  }, [voiceUrl])

  const startAnimation = () => {
    if (animationIntervalRef.current) clearInterval(animationIntervalRef.current)
    setAnimationStep(1) 
    
    // Animation loop: 1 -> 2 -> 3 -> 1 ...
    // Step 0 is usually reserved for "off" or dot only, but standard wechat loop is:
    // Dot + 1 arc -> Dot + 2 arcs -> Dot + 3 arcs
    animationIntervalRef.current = setInterval(() => {
      setAnimationStep(prev => (prev % 3) + 1)
    }, 500)
  }

  const stopAnimation = () => {
    if (animationIntervalRef.current) clearInterval(animationIntervalRef.current)
    setAnimationStep(3) // Reset to full static
  }

  const togglePlay = async () => {
    if (!audioRef.current) return

    if (isPlaying) {
      audioRef.current.pause()
    } else {
      try {
        if (audioRef.current.ended) {
          audioRef.current.currentTime = 0
        }
        await audioRef.current.play()
      } catch (error) {
        console.error("Failed to play voice:", error)
      }
    }
  }

  const handleTranscribe = async (e: React.MouseEvent) => {
    e.stopPropagation()
    if (isTranscribing || transcribeText !== null) return
    setIsTranscribing(true)
    try {
      const res = await mediaApi.transcribeVoice(id)
      setTranscribeText(res.text || "(无识别结果)")
    } catch (err: any) {
      setTranscribeText("转文字失败: " + (err?.message || "未知错误"))
    } finally {
      setIsTranscribing(false)
    }
  }

  // Calculate width. Base 60px, + px per second. Max 200px.
  // WeChat logic roughly: min 40px, max ~160px.
  const width = Math.min(200, Math.max(70, 70 + (displayDuration * 5)))

  return (
    <div className="flex flex-col gap-1">
      <div
        className={cn(
          "flex items-center gap-2 py-1 px-3 cursor-pointer select-none transition-all rounded-md hover:bg-black/5 dark:hover:bg-white/5",
          isSelf ? "flex-row-reverse" : "flex-row"
        )}
        style={{ width: duration ? `${width}px` : 'auto' }}
        onClick={(e) => {
          e.stopPropagation()
          togglePlay()
        }}
      >
        {/* Icon */}
        <div className={cn(
          "shrink-0 flex items-center justify-center w-5 h-5",
          isSelf ? "rotate-180" : ""
        )}>
          <VoiceIcon step={animationStep} />
        </div>

        {/* Duration */}
        <span className="text-sm text-foreground/80 min-w-[16px] text-center">
          {displayDuration}"
        </span>

        {/* Transcribe button */}
        {transcribeText === null && (
          <button
            className="shrink-0 ml-1 p-0.5 rounded hover:bg-black/10 dark:hover:bg-white/10 text-muted-foreground hover:text-foreground transition-colors"
            onClick={handleTranscribe}
            disabled={isTranscribing}
            title="转文字"
          >
            {isTranscribing ? (
              <Loader2 className="w-3.5 h-3.5 animate-spin" />
            ) : (
              <Type className="w-3.5 h-3.5" />
            )}
          </button>
        )}
      </div>

      {/* Transcribed text */}
      {transcribeText !== null && (
        <div className={cn(
          "text-xs text-foreground/70 px-3 py-1 max-w-[240px]",
          isSelf ? "text-right" : "text-left"
        )}>
          {transcribeText}
        </div>
      )}
    </div>
  )
}

function VoiceIcon({ step }: { step: number }) {
  // WeChat style icon: Dot + 3 Arcs radiating to the RIGHT.
  // For sender, the parent container rotates it 180 deg.
  
  return (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
       {/* Core Dot - Always visible */}
       <circle cx="4" cy="12" r="2" fill="currentColor" />

       {/* Arc 1 */}
       <path 
         d="M9 8C10.5 9.5 10.5 14.5 9 16" 
         stroke="currentColor" 
         strokeWidth="2" 
         strokeLinecap="round" 
         strokeLinejoin="round"
         className="transition-opacity duration-150"
         style={{ opacity: step >= 1 ? 1 : 0.3 }}
       />

       {/* Arc 2 */}
       <path 
         d="M13 5C16 8 16 16 13 19" 
         stroke="currentColor" 
         strokeWidth="2" 
         strokeLinecap="round" 
         strokeLinejoin="round"
         className="transition-opacity duration-150"
         style={{ opacity: step >= 2 ? 1 : 0.3 }}
       />

       {/* Arc 3 */}
       <path 
         d="M17 2C22 7 22 17 17 22" 
         stroke="currentColor" 
         strokeWidth="2" 
         strokeLinecap="round" 
         strokeLinejoin="round"
         className="transition-opacity duration-150"
         style={{ opacity: step >= 3 ? 1 : 0.3 }}
       />
    </svg>
  )
}
