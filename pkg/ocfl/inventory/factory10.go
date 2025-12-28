package inventory

var f10 = NewFactory10()

func NewFactory10() Factory {
	return &factory10{
		Factory: NewFactoryBase(),
	}
}

type factory10 struct {
	Factory
}

var _ Factory = (*factory10)(nil)
