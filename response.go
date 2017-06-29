package just

type IResponse interface {
	GetData() []byte
	GetStatus() int
}

type Response struct {
	Status int
	Bytes  []byte
}

func (r *Response) GetStatus() int {
	return r.Status
}

func (r *Response) GetData() []byte {
	return r.Bytes
}
