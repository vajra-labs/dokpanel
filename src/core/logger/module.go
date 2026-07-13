package logger

import "go.uber.org/fx"

// Module is the fx module for the logger package.
var Module = fx.Module("logger", fx.Invoke(configLogger))
