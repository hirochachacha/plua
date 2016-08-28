package object

import "github.com/hirochachacha/blua/internal/limits"

const (
	MaxInteger = Integer(limits.MaxInt)
	MinInteger = Integer(limits.MinInt)
	MaxNumber  = Number(limits.MaxFloat64)
)
