package mixcoin

import (
	"github.com/stretchr/testify/mock"
)

type MockPool struct {
	mock.Mock
}

func (p *MockPool) ReceivingKeys() []string {
	args := p.Mock.Called()
	return args.Get(0).([]string)
}

func (p *MockPool) Scan(k []string) []PoolItem {
	args := p.Mock.Called(k)
	return args.Get(0).([]PoolItem)
}

func (p *MockPool) Filter(f func(PoolItem) bool) {
	p.Mock.Called(f)
}

func (p *MockPool) Get(t PoolType) (PoolItem, error) {
	args := p.Mock.Called(t)
	return args.Get(0).(PoolItem), args.Error(1)
}

func (p *MockPool) Put(t PoolType, i PoolItem) {
	p.Mock.Called(t, i)
}

func NewMockPool() *MockPool {
	return &MockPool{}
}
