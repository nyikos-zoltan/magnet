package magnet

func (m *Magnet) Prepare() error {
	m.detectCycles()
	for _, v := range m.providerMap {
		if !v.forceRecreate {
			if _, err := v.Build(m); err != nil {
				return err
			}
		}
	}
	return nil
}
