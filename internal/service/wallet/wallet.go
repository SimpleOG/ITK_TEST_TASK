package wallet

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
	"tryingMicro/OrderAccepter/internal/repository"
	"tryingMicro/OrderAccepter/package/logger"
)

const (
	OperationDeposit  = "DEPOSIT"
	OperationWithdraw = "WITHDRAW"
)

type WalletService interface {
	ProcessOperation(ctx context.Context, walletID uuid.UUID, opType string, amount float64) (repository.Wallet, error)
	GetBalance(ctx context.Context, walletID uuid.UUID) (repository.Wallet, error)
	CreateWallet(ctx context.Context) (repository.Wallet, error)
}

type walletService struct {
	repo   repository.Repository
	logger logger.Logger
	locker *walletLocker
}

func New(repo repository.Repository, log logger.Logger) WalletService {
	return &walletService{
		repo:   repo,
		logger: log,
		locker: newWalletLocker(),
	}
}

func (s *walletService) ProcessOperation(ctx context.Context, walletID uuid.UUID, opType string, amount float64) (repository.Wallet, error) {
	unlock := s.locker.Lock(walletID)
	defer unlock()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var result repository.Wallet

	err := s.repo.WithTx(ctx, func(q repository.Querier) error {
		w, err := q.GetWalletForUpdate(ctx, walletID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				s.logger.Warn("wallet not found", zap.String("walletId", walletID.String()))
				return ErrWalletNotFound
			}
			s.logger.Error("failed to get wallet for update", zap.String("walletId", walletID.String()), zap.Error(err))
			return err
		}

		var newBalance float64
		switch opType {
		case OperationDeposit:
			newBalance = w.Balance + amount
		case OperationWithdraw:
			if w.Balance < amount {
				s.logger.Warn("insufficient funds", zap.String("walletId", walletID.String()), zap.Float64("balance", w.Balance), zap.Float64("amount", amount))
				return ErrInsufficientFunds
			}
			newBalance = w.Balance - amount
		default:
			return ErrInvalidOperation
		}

		result, err = q.UpdateWalletBalance(ctx, repository.UpdateWalletBalanceParams{
			ID:      walletID,
			Balance: newBalance,
		})
		if err != nil {
			s.logger.Error("failed to update wallet balance", zap.String("walletId", walletID.String()), zap.Error(err))
		}
		return err
	})

	if err == nil {
		s.logger.Info("wallet operation completed", zap.String("walletId", walletID.String()), zap.String("operation", opType), zap.Float64("amount", amount))
	}

	return result, err
}

func (s *walletService) GetBalance(ctx context.Context, walletID uuid.UUID) (repository.Wallet, error) {
	w, err := s.repo.GetWallet(ctx, walletID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.logger.Warn("wallet not found", zap.String("walletId", walletID.String()))
			return repository.Wallet{}, ErrWalletNotFound
		}
		s.logger.Error("failed to get wallet", zap.String("walletId", walletID.String()), zap.Error(err))
		return repository.Wallet{}, err
	}
	return w, nil
}
func (s *walletService) CreateWallet(ctx context.Context) (repository.Wallet, error) {
	id := uuid.New()
	w, err := s.repo.CreateWallet(ctx, id)
	if err != nil {
		s.logger.Error("failed to create wallet", zap.String("walletId", id.String()), zap.Error(err))
		return repository.Wallet{}, err
	}
	s.logger.Info("wallet created", zap.String("walletId", id.String()))
	return w, nil
}
