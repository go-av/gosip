package dialog

type DialogState uint8

const (
	Proceeding DialogState = iota // 等待处理
	Trying                        // 收到 100
	Ringing                       // 铃声 180
	Accepted                      // 接收 200
	Error                         // 错误
)

func (d DialogState) String() string {
	switch d {
	case Proceeding:
		return "Proceeding"
	case Trying:
		return "Trying"
	case Ringing:
		return "Ringing"
	case Accepted:
		return "Accepted"
	case Error:
		return "Error"
	}
	return "Error"
}
