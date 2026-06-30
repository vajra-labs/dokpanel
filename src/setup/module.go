package setup

import "go.uber.org/fx"

// Module provides *Runner via fx.
var Module = fx.Module("setup", fx.Provide(newRunner))
