package dialog

type DialogState string

const (
	Proceeding DialogState = "Proceeding"
	Ringing    DialogState = "Ringing"  // 铃声
	Answered   DialogState = "Answered" // 接听
	Missed     DialogState = "Missed"   // 未接
	Hangup     DialogState = "Hangup"   // 挂断
	Error      DialogState = "error"    // 错误
)
