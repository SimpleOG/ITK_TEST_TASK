package wallet_test

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"tryingMicro/OrderAccepter/internal/repository"
	"tryingMicro/OrderAccepter/internal/service/wallet"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) WithTx(ctx context.Context, fn func(repository.Querier) error) error {
	args := m.Called(ctx, fn)
	return args.Error(0)
}

func (m *MockRepository) GetWallet(ctx context.Context, id uuid.UUID) (repository.Wallet, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(repository.Wallet), args.Error(1)
}

func (m *MockRepository) GetWalletForUpdate(ctx context.Context, id uuid.UUID) (repository.Wallet, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(repository.Wallet), args.Error(1)
}

func (m *MockRepository) UpdateWalletBalance(ctx context.Context, arg repository.UpdateWalletBalanceParams) (repository.Wallet, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(repository.Wallet), args.Error(1)
}

func (m *MockRepository) CreateWallet(ctx context.Context, id uuid.UUID) (repository.Wallet, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(repository.Wallet), args.Error(1)
}

func withTxOK(m *MockRepository) {
	m.On("WithTx", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(repository.Querier) error)
			fn(m)
		}).Return(nil)
}

func withTxErr(m *MockRepository, err error) {
	m.On("WithTx", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(repository.Querier) error)
			fn(m)
		}).Return(err)
}

func makeWallet(balance float64) repository.Wallet {
	return repository.Wallet{
		ID:        uuid.New(),
		Balance:   balance,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func TestProcessOperation_Deposit_Success(t *testing.T) {
	existing := makeWallet(100)
	updated := existing
	updated.Balance = 150

	mockRepo := new(MockRepository)
	withTxOK(mockRepo)
	mockRepo.On("GetWalletForUpdate", mock.Anything, existing.ID).
		Return(existing, nil)
	mockRepo.On("UpdateWalletBalance", mock.Anything, repository.UpdateWalletBalanceParams{
		ID: existing.ID, Balance: 150,
	}).Return(updated, nil)

	svc := wallet.New(mockRepo, zap.NewNop())
	result, err := svc.ProcessOperation(context.Background(), existing.ID, wallet.OperationDeposit, 50)

	require.NoError(t, err)
	assert.Equal(t, 150.0, result.Balance)
	mockRepo.AssertExpectations(t)
}

func TestProcessOperation_Withdraw_Success(t *testing.T) {
	existing := makeWallet(100)
	updated := existing
	updated.Balance = 70

	mockRepo := new(MockRepository)
	withTxOK(mockRepo)
	mockRepo.On("GetWalletForUpdate", mock.Anything, existing.ID).
		Return(existing, nil)
	mockRepo.On("UpdateWalletBalance", mock.Anything, repository.UpdateWalletBalanceParams{
		ID: existing.ID, Balance: 70,
	}).Return(updated, nil)

	svc := wallet.New(mockRepo, zap.NewNop())
	result, err := svc.ProcessOperation(context.Background(), existing.ID, wallet.OperationWithdraw, 30)

	require.NoError(t, err)
	assert.Equal(t, 70.0, result.Balance)
	mockRepo.AssertExpectations(t)
}

func TestProcessOperation_WalletNotFound(t *testing.T) {
	walletID := uuid.New()

	mockRepo := new(MockRepository)
	withTxErr(mockRepo, wallet.ErrWalletNotFound)
	mockRepo.On("GetWalletForUpdate", mock.Anything, walletID).
		Return(repository.Wallet{}, pgx.ErrNoRows)

	svc := wallet.New(mockRepo, zap.NewNop())
	_, err := svc.ProcessOperation(context.Background(), walletID, wallet.OperationDeposit, 50)

	require.ErrorIs(t, err, wallet.ErrWalletNotFound)
	mockRepo.AssertExpectations(t)
}

func TestProcessOperation_RepoError(t *testing.T) {
	walletID := uuid.New()
	repoErr := io.ErrUnexpectedEOF

	mockRepo := new(MockRepository)
	withTxErr(mockRepo, repoErr)
	mockRepo.On("GetWalletForUpdate", mock.Anything, walletID).
		Return(repository.Wallet{}, repoErr)

	svc := wallet.New(mockRepo, zap.NewNop())
	_, err := svc.ProcessOperation(context.Background(), walletID, wallet.OperationDeposit, 50)

	require.ErrorIs(t, err, repoErr)
	mockRepo.AssertExpectations(t)
}

func TestProcessOperation_InsufficientFunds(t *testing.T) {
	existing := makeWallet(50)

	mockRepo := new(MockRepository)
	withTxErr(mockRepo, wallet.ErrInsufficientFunds)
	mockRepo.On("GetWalletForUpdate", mock.Anything, existing.ID).
		Return(existing, nil)

	svc := wallet.New(mockRepo, zap.NewNop())
	_, err := svc.ProcessOperation(context.Background(), existing.ID, wallet.OperationWithdraw, 100)

	require.ErrorIs(t, err, wallet.ErrInsufficientFunds)
	mockRepo.AssertExpectations(t)
}

func TestProcessOperation_InvalidOperation(t *testing.T) {
	existing := makeWallet(100)

	mockRepo := new(MockRepository)
	withTxErr(mockRepo, wallet.ErrInvalidOperation)
	mockRepo.On("GetWalletForUpdate", mock.Anything, existing.ID).
		Return(existing, nil)

	svc := wallet.New(mockRepo, zap.NewNop())
	_, err := svc.ProcessOperation(context.Background(), existing.ID, "REFUND", 50)

	require.ErrorIs(t, err, wallet.ErrInvalidOperation)
	mockRepo.AssertExpectations(t)
}

func TestProcessOperation_UpdateError(t *testing.T) {
	existing := makeWallet(100)
	updateErr := errors.New("db update failed")

	mockRepo := new(MockRepository)
	withTxErr(mockRepo, updateErr)
	mockRepo.On("GetWalletForUpdate", mock.Anything, existing.ID).
		Return(existing, nil)
	mockRepo.On("UpdateWalletBalance", mock.Anything, repository.UpdateWalletBalanceParams{
		ID: existing.ID, Balance: 150,
	}).Return(repository.Wallet{}, updateErr)

	svc := wallet.New(mockRepo, zap.NewNop())
	_, err := svc.ProcessOperation(context.Background(), existing.ID, wallet.OperationDeposit, 50)

	require.ErrorIs(t, err, updateErr)
	mockRepo.AssertExpectations(t)
}

func TestGetBalance_Success(t *testing.T) {
	expected := makeWallet(200)

	mockRepo := new(MockRepository)
	mockRepo.On("GetWallet", mock.Anything, expected.ID).Return(expected, nil)

	svc := wallet.New(mockRepo, zap.NewNop())
	result, err := svc.GetBalance(context.Background(), expected.ID)

	require.NoError(t, err)
	assert.Equal(t, expected.ID, result.ID)
	assert.Equal(t, expected.Balance, result.Balance)
	mockRepo.AssertExpectations(t)
}

func TestGetBalance_NotFound(t *testing.T) {
	walletID := uuid.New()

	mockRepo := new(MockRepository)
	mockRepo.On("GetWallet", mock.Anything, walletID).
		Return(repository.Wallet{}, pgx.ErrNoRows)

	svc := wallet.New(mockRepo, zap.NewNop())
	_, err := svc.GetBalance(context.Background(), walletID)

	require.ErrorIs(t, err, wallet.ErrWalletNotFound)
	mockRepo.AssertExpectations(t)
}

func TestGetBalance_RepoError(t *testing.T) {
	walletID := uuid.New()
	repoErr := io.ErrUnexpectedEOF

	mockRepo := new(MockRepository)
	mockRepo.On("GetWallet", mock.Anything, walletID).
		Return(repository.Wallet{}, repoErr)

	svc := wallet.New(mockRepo, zap.NewNop())
	_, err := svc.GetBalance(context.Background(), walletID)

	require.ErrorIs(t, err, repoErr)
	mockRepo.AssertExpectations(t)
}
func TestCreateWallet_Success(t *testing.T) {
	expected := makeWallet(0)

	mockRepo := new(MockRepository)
	mockRepo.On("CreateWallet", mock.Anything, mock.AnythingOfType("uuid.UUID")).
		Return(expected, nil)

	svc := wallet.New(mockRepo, zap.NewNop())
	result, err := svc.CreateWallet(context.Background())

	require.NoError(t, err)
	assert.Equal(t, expected.ID, result.ID)
	assert.Equal(t, 0.0, result.Balance)
	mockRepo.AssertExpectations(t)
}

func TestCreateWallet_RepoError(t *testing.T) {
	repoErr := errors.New("db error")

	mockRepo := new(MockRepository)
	mockRepo.On("CreateWallet", mock.Anything, mock.AnythingOfType("uuid.UUID")).
		Return(repository.Wallet{}, repoErr)

	svc := wallet.New(mockRepo, zap.NewNop())
	_, err := svc.CreateWallet(context.Background())

	require.ErrorIs(t, err, repoErr)
	mockRepo.AssertExpectations(t)
}
