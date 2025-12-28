package inventory

var f11 = NewFactory11()

func NewFactory11() Factory {
	return &factory11{
		Factory: NewFactoryBase(),
	}
}

type factory11 struct {
	Factory
}

var _ Factory = (*factory11)(nil)
