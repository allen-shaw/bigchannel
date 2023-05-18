package resolver

import (
	"fmt"

	"google.golang.org/grpc/resolver"
)

var (
	Scheme      = "bigchannel"
	ServiceName = "bigchannel-broker"
)

type bcResolverBuilder struct{}

func newBCRB() *bcResolverBuilder {
	fmt.Println("register bc resolver builder")
	return &bcResolverBuilder{}
}

func (*bcResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	fmt.Printf("bc resolver builder Build, target: %+v, cc %+v\n", target, cc)
	r := &bcResolver{
		target: target,
		cc:     cc,
		addrsStore: map[string][]string{
			ServiceName: {"127.0.0.1:50049", "127.0.0.1:50050", "127.0.0.1:50051"},
		},
	}
	r.ResolveNow(resolver.ResolveNowOptions{})
	return r, nil
}
func (*bcResolverBuilder) Scheme() string { return Scheme }

type bcResolver struct {
	target     resolver.Target
	cc         resolver.ClientConn
	addrsStore map[string][]string
}

func (r *bcResolver) ResolveNow(o resolver.ResolveNowOptions) {
	fmt.Println("ResolveNow", o)
	// 直接从map中取出对于的addrList
	addrStrs := r.addrsStore[r.target.Endpoint()]
	addrs := make([]resolver.Address, len(addrStrs))
	for i, s := range addrStrs {
		addrs[i] = resolver.Address{Addr: s}
	}
	r.cc.UpdateState(resolver.State{Addresses: addrs})
}

func (*bcResolver) Close() {
	fmt.Println("bc resovler close!")
}

func init() {
	// Register the example ResolverBuilder. This is usually done in a package's
	// init() function.
	resolver.Register(newBCRB())
}
