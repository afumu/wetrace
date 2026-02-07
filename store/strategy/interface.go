package strategy

// GroupType 代表数据库文件的逻辑分类
type GroupType int

const (
	Unknown GroupType = iota
	Message
	Contact
	Image
	Video
	File
	Voice
	Session
)

func (g GroupType) String() string {
	switch g {
	case Message:
		return "Message"
	case Contact:
		return "Contact"
	case Image:
		return "Image"
	case Video:
		return "Video"
	case File:
		return "File"
	case Voice:
		return "Voice"
	case Session:
		return "Session"
	default:
		return "Unknown"
	}
}

// FileMeta 包含从文件名中提取的信息
type FileMeta struct {
	Type  GroupType
	Path  string
	Index string // 例如：MSG0.db 为 "0"，MicroMsg.db 为 ""
}

// Strategy 定义了特定平台的行为
type Strategy interface {
	// Identify 根据文件名对文件进行分类。
	// 如果文件被识别，返回元数据和 true。
	Identify(filename string) (FileMeta, bool)
}
