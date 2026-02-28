package controllers

import (
	"tryingMicro/OrderAccepter/internal/api/controllers/wallet"
	"tryingMicro/OrderAccepter/internal/service"
	"tryingMicro/OrderAccepter/package/logger"
)

type Controllers struct {
	Wallet wallet.WalletController
}

func NewControllers(service *service.Services, log logger.Logger) *Controllers {
	return &Controllers{
		Wallet: wallet.New(service.Wallet, log),
	}
}
