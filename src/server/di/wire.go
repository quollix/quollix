//go:build wireinject

package di

import (
	"github.com/google/wire"
)

func WireDependencies() (*Dependencies, error) {
	wire.Build(SharedSet, DependenciesSet)
	return nil, nil
}
