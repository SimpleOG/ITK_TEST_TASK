package controllers

import (
	"tryingMicro/OrderAccepter/internal/api/controllers/wallet"
	walletService "tryingMicro/OrderAccepter/internal/service/wallet"
	"tryingMicro/OrderAccepter/package/logger"
)

type Controllers struct {
	Wallet wallet.WalletController
}

func NewControllers(walletService walletService.WalletService, log logger.Logger) *Controllers {
	return &Controllers{
		Wallet: wallet.New(walletService, log),
	}
}
