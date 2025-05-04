package world

import (
	"math/rand/v2"

	"github.com/yuta-otsubo/isucon-sutra/bench/internal/random"
)

type Context struct {
	world  *World
	rand   *rand.Rand
	client Client
}

func NewContext(world *World, client Client) *Context {
	return &Context{
		world: world,
		// TODO: rand どうする？
		rand:   rand.New(random.NewLockedSource(rand.NewPCG(rand.Uint64(), rand.Uint64()))),
		client: client,
	}
}
