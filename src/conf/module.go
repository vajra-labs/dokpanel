package conf

import "go.uber.org/fx"

// Config Module
var Module = fx.Module("config", fx.Provide(provideConfig))
