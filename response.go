package just

type IResponse interface {
	GetData() []byte
	GetStatus() int
	GetHeaders() map[string]string
}

type Response struct {
	Status  int
	Bytes   []byte
	Headers map[string]string
}

func (r *Response) GetStatus() int {
	return r.Status
}

func (r *Response) GetData() []byte {
	return r.Bytes
}

func (r *Response) GetHeaders() map[string]string {
	return r.Headers
}
