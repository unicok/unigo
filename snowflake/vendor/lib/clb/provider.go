package clb

type AddressProvider interface {
	GetAddress() (Address, error)
}

type StaticAddressProvider struct {
	Address Address
}

func (s *StaticAddressProvider) GetAddress() (Address, error) {
	return s.Address, nil
}

func NewAddressProvider(address string) *SRVAddressProvider {
	return &SRVAddressProvider{
		Lb:      New(),
		Address: address,
	}
}

type SRVAddressProvider struct {
	Lb      LoadBalancer
	Address string
}

func (s *SRVAddressProvider) GetAddress() (Address, error) {
	return s.Lb.GetAddress(s.Address)
}
