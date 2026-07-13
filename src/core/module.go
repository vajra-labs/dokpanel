package core

import (
	"goploy/src/core/logger"

	"go.uber.org/fx"
)

var Module = fx.Module(
	"core",
	logger.Module,
	fx.Provide(provideFiber),
)
