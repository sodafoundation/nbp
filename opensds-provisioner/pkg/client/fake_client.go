package client

type FakeClient struct {
}

var _ WarpOpensdsClient = &FakeClient{}

func NewFakeClient(endpoint string) WarpOpensdsClient {
	return &FakeClient{}
}

func (c *FakeClient) Provision(opts map[string]string) (string, error) {
	return "volume-opendsds-nbp-privisioner", nil
}

func (c *FakeClient) Delete(volumeId string) error {
	return nil
}
