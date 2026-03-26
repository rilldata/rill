package hubspot

var _ Client = &noop{}

type noop struct{}

func NewNoop() Client {
	return &noop{}
}

func (n *noop) UpsertContact(email string, properties map[string]string) {}
