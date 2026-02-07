import { mediaApi } from "@/api/media"

interface VideoMessageProps {
  md5?: string
}

export function VideoMessage({ md5 }: VideoMessageProps) {
  if (!md5) {
    return <div className="text-sm text-muted-foreground">[视频无法加载]</div>
  }

  const videoUrl = mediaApi.getVideoUrl(md5)

  return (
    <div className="relative max-w-[240px] flex flex-col gap-1">
      <video 
        controls 
        className="rounded-lg w-full h-auto bg-black"
        preload="metadata"
      >
        <source src={videoUrl} />
        <div className="p-4 text-center text-white text-xs">
          <p>您的浏览器不支持此视频播放</p>
          <p className="opacity-70 mt-1">(可能是 HEVC/H.265 编码)</p>
        </div>
      </video>
      <a 
        href={videoUrl} 
        target="_blank" 
        rel="noopener noreferrer"
        className="text-xs text-blue-500 hover:underline self-end"
      >
      </a>
    </div>
  )
}
