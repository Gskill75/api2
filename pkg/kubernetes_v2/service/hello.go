package service

type HelloService struct{}

func NewHelloService() *HelloService {
	return &HelloService{}
}

func (s *HelloService) GetHelloWorld() string {
	// Ici tu pourrais mettre une logique métier réelle v2…
	return "hello world from kubernetes v2!"
}
