package tui

import (
	"strings"
	"sync"
)

type StringBuilderPool struct {
	pool *sync.Pool
}

func NewStringBuilderPool() *StringBuilderPool {
	return &StringBuilderPool{
		pool: &sync.Pool{
			New: func() interface{} {
				return &strings.Builder{}
			},
		},
	}
}

func (p *StringBuilderPool) Get() *strings.Builder {
	sb := p.pool.Get().(*strings.Builder)
	sb.Reset()
	return sb
}

func (p *StringBuilderPool) Put(sb *strings.Builder) {
	sb.Reset()
	p.pool.Put(sb)
}
