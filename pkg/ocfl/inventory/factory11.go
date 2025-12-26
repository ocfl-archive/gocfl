package inventory

func NewFactory11() *Factory11 {
	return &Factory11{
		FactoryBase: NewFactoryBase(),
	}
}

type Factory11 struct {
	*FactoryBase
}

var _ Factory = (*Factory11)(nil)
