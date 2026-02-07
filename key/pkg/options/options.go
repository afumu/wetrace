package options

type CliOptions struct {
	AutoMode           bool
	ImageKeyMode       bool
	ManualPid          uint32
	NoRestart          bool
	WechatPath         string
	WechatDataPath     string
	DllPath            string
	StartupWaitTimeout int
	WindowWaitTimeout  int
	KeyWaitTimeout     int
	OutputFile         string
	OutputFormat       string // "text" or "json"
	ExtendedJson       bool
	WorkDir            string
	HttpEnabled        bool
	HttpAddr           string
	HideConsole        bool
	Verbose            bool
	Quiet              bool
	NoColor            bool
}
