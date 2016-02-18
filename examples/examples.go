//Package examples hosts some common examples
package examples

// OrderCustom custom object provider
type FullOrderCustomProvider struct {
}

func (cp FullOrderCustomProvider) NewOrderCustom() interface{} {
	return &OrderCustom{}
}

func (cp FullOrderCustomProvider) NewPositionCustom() interface{} {
	return nil
}

func (cp FullOrderCustomProvider) NewAddressCustom() interface{} {
	return nil
}

type OrderCustom struct {
	ResponsibleSmurf string
}
