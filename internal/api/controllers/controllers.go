package controllers

import (
	"tryingMicro/OrderAccepter/package/logger"
)

type Controllers struct {
}

func NewControllers(logger logger.Logger) *Controllers {
	return &Controllers{}
}
