package responsehandler

type Response struct {
	Status int
	Msg    string
	Data   interface{}
}

func Build(status int, msg string, data interface{}) (int, Response) {
	return status, Response{Status: status, Msg: msg, Data: data}
}
