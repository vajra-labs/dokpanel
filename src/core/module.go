package core

import (
	apidocs "goploy/src/core/apidoc"
	"goploy/src/core/logger"

	"go.uber.org/fx"
)

var Module = fx.Module(
	"core",
	logger.Module,
	apidocs.Module,
	fx.Provide(provideFiber),
)
