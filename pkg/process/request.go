package process

type PHPHandler struct {
	Request  string
	Response chan string
	Error    chan error
}

func NewPHPHandler(req string) *PHPHandler {
	return &PHPHandler{
		Request:  req,
		Response: make(chan string, 1),
		Error:    make(chan error, 1),
	}
}
