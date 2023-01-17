package etcd

const ResolverName = "etcd"

func New() (*Resolver, error) {

	return &Resolver{}, nil
}

type Resolver struct {
}

func (r *Resolver) Resolve(string) (string, error) {
	return "etcd", nil
}

func (r *Resolver) Info() string {
	return ""
}

func (r *Resolver) Name() string {
	return ResolverName
}

func (r *Resolver) Prefix() string {
	return ResolverName + ":"
}
