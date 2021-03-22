package magnet

func (m *Magnet) Call(fn interface{}) (*CallResults, error) {
	return m.NewCaller(fn).Call()
}
