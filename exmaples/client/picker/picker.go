package picker

import (
	"fmt"
	"log"
	"sync"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

func init() {
	balancer.Register(newBuilder())
}

const Scheme = "bigchannel-lb"

type bcPikerBuilder struct {
}

func newBcPikerBuilder() *bcPikerBuilder {
	fmt.Println("new bc picker builder")
	return &bcPikerBuilder{}
}

func newBuilder() balancer.Builder {
	return base.NewBalancerBuilder(Scheme, newBcPikerBuilder(), base.Config{HealthCheck: true})
}

func (p *bcPikerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {

	fmt.Println("build bc picker...", info)

	// 没有可用的连接
	if len(info.ReadySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}

	scs := make([]balancer.SubConn, 0, len(info.ReadySCs))

	for subConn := range info.ReadySCs {
		scs = append(scs, subConn)
	}

	return &bcPiker{
		scs: scs,
	}
}

type bcPiker struct {
	scs []balancer.SubConn
	mu  sync.Mutex
	id  int
}

func (p *bcPiker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	log.Println("bcPicker.Pick called...", info)
	p.mu.Lock()
	p.id++
	p.id = p.id % len(p.scs)
	sc := p.scs[p.id]
	p.mu.Unlock()

	fmt.Println("picked", sc)
	return balancer.PickResult{SubConn: sc}, nil
}
