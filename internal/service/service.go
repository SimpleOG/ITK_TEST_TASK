package service

import (
	"tryingMicro/OrderAccepter/internal/repository"
	"tryingMicro/OrderAccepter/internal/service/wallet"
	"tryingMicro/OrderAccepter/package/logger"
)

type Services struct {
	Wallet wallet.WalletService
}

func NewServices(repo repository.Repository, log logger.Logger) *Services {
	return &Services{
		Wallet: wallet.New(repo, log),
	}
}
