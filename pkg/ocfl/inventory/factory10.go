package inventory

var factory10 = NewFactory10()

func NewFactory10() *Factory10 {
	return &Factory10{
		FactoryBase: NewFactoryBase(),
	}
}

type Factory10 struct {
	*FactoryBase
}

var _ Factory = (*Factory10)(nil)
