package wallet_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"tryingMicro/OrderAccepter/internal/api/controllers/wallet"
	"tryingMicro/OrderAccepter/internal/repository"
	walletSvc "tryingMicro/OrderAccepter/internal/service/wallet"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

type MockWalletService struct {
	mock.Mock
}

func (m *MockWalletService) ProcessOperation(ctx context.Context, walletID uuid.UUID, opType string, amount float64) (repository.Wallet, error) {
	args := m.Called(ctx, walletID, opType, amount)
	return args.Get(0).(repository.Wallet), args.Error(1)
}

func (m *MockWalletService) GetBalance(ctx context.Context, walletID uuid.UUID) (repository.Wallet, error) {
	args := m.Called(ctx, walletID)
	return args.Get(0).(repository.Wallet), args.Error(1)
}

func setupRouter(svc walletSvc.WalletService) *gin.Engine {
	r := gin.New()
	ctrl := wallet.New(svc, zap.NewNop())
	r.POST("/wallets/:walletId/operation", ctrl.ProcessOperation)
	r.GET("/wallets/:walletId", ctrl.GetBalance)
	return r
}

func makeWallet(balance float64) repository.Wallet {
	return repository.Wallet{
		ID:        uuid.New(),
		Balance:   balance,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func decodeBody(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var body map[string]interface{}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	return body
}

func TestProcessOperation_Deposit_Success(t *testing.T) {
	w := makeWallet(150)
	mockSvc := new(MockWalletService)
	mockSvc.On("ProcessOperation", mock.Anything, w.ID, walletSvc.OperationDeposit, 50.0).
		Return(w, nil)

	body := `{"operationType":"DEPOSIT","amount":50}`
	req := httptest.NewRequest(http.MethodPost, "/wallets/"+w.ID.String()+"/operation", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	setupRouter(mockSvc).ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	resp := decodeBody(t, rec)
	assert.InDelta(t, 150.0, resp["balance"], 0.001)
	mockSvc.AssertExpectations(t)
}

func TestProcessOperation_Withdraw_Success(t *testing.T) {
	w := makeWallet(70)
	mockSvc := new(MockWalletService)
	mockSvc.On("ProcessOperation", mock.Anything, w.ID, walletSvc.OperationWithdraw, 30.0).
		Return(w, nil)

	body := `{"operationType":"WITHDRAW","amount":30}`
	req := httptest.NewRequest(http.MethodPost, "/wallets/"+w.ID.String()+"/operation", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	setupRouter(mockSvc).ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	resp := decodeBody(t, rec)
	assert.InDelta(t, 70.0, resp["balance"], 0.001)
	mockSvc.AssertExpectations(t)
}

func TestProcessOperation_InvalidWalletID(t *testing.T) {
	mockSvc := new(MockWalletService)

	body := `{"operationType":"DEPOSIT","amount":50}`
	req := httptest.NewRequest(http.MethodPost, "/wallets/not-a-uuid/operation", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	setupRouter(mockSvc).ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	mockSvc.AssertNotCalled(t, "ProcessOperation")
}

func TestProcessOperation_InvalidBody_MissingFields(t *testing.T) {
	mockSvc := new(MockWalletService)

	req := httptest.NewRequest(http.MethodPost, "/wallets/"+uuid.New().String()+"/operation", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	setupRouter(mockSvc).ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	mockSvc.AssertNotCalled(t, "ProcessOperation")
}

func TestProcessOperation_InvalidBody_NegativeAmount(t *testing.T) {
	mockSvc := new(MockWalletService)

	body := `{"operationType":"DEPOSIT","amount":-10}`
	req := httptest.NewRequest(http.MethodPost, "/wallets/"+uuid.New().String()+"/operation", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	setupRouter(mockSvc).ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	mockSvc.AssertNotCalled(t, "ProcessOperation")
}

func TestProcessOperation_WalletNotFound(t *testing.T) {
	walletID := uuid.New()
	mockSvc := new(MockWalletService)
	mockSvc.On("ProcessOperation", mock.Anything, walletID, walletSvc.OperationDeposit, 50.0).
		Return(repository.Wallet{}, walletSvc.ErrWalletNotFound)

	body := `{"operationType":"DEPOSIT","amount":50}`
	req := httptest.NewRequest(http.MethodPost, "/wallets/"+walletID.String()+"/operation", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	setupRouter(mockSvc).ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
	resp := decodeBody(t, rec)
	assert.Equal(t, walletSvc.ErrWalletNotFound.Error(), resp["error"])
	mockSvc.AssertExpectations(t)
}

func TestProcessOperation_InsufficientFunds(t *testing.T) {
	walletID := uuid.New()
	mockSvc := new(MockWalletService)
	mockSvc.On("ProcessOperation", mock.Anything, walletID, walletSvc.OperationWithdraw, 100.0).
		Return(repository.Wallet{}, walletSvc.ErrInsufficientFunds)

	body := `{"operationType":"WITHDRAW","amount":100}`
	req := httptest.NewRequest(http.MethodPost, "/wallets/"+walletID.String()+"/operation", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	setupRouter(mockSvc).ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	resp := decodeBody(t, rec)
	assert.Equal(t, walletSvc.ErrInsufficientFunds.Error(), resp["error"])
	mockSvc.AssertExpectations(t)
}

func TestProcessOperation_InvalidOperation(t *testing.T) {
	walletID := uuid.New()
	mockSvc := new(MockWalletService)
	mockSvc.On("ProcessOperation", mock.Anything, walletID, "REFUND", 50.0).
		Return(repository.Wallet{}, walletSvc.ErrInvalidOperation)

	body := `{"operationType":"REFUND","amount":50}`
	req := httptest.NewRequest(http.MethodPost, "/wallets/"+walletID.String()+"/operation", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	setupRouter(mockSvc).ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	resp := decodeBody(t, rec)
	assert.Equal(t, walletSvc.ErrInvalidOperation.Error(), resp["error"])
	mockSvc.AssertExpectations(t)
}

func TestProcessOperation_ServiceError(t *testing.T) {
	walletID := uuid.New()
	mockSvc := new(MockWalletService)
	mockSvc.On("ProcessOperation", mock.Anything, walletID, walletSvc.OperationDeposit, 50.0).
		Return(repository.Wallet{}, errors.New("db error"))

	body := `{"operationType":"DEPOSIT","amount":50}`
	req := httptest.NewRequest(http.MethodPost, "/wallets/"+walletID.String()+"/operation", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	setupRouter(mockSvc).ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	resp := decodeBody(t, rec)
	assert.Equal(t, "internal server error", resp["error"])
	mockSvc.AssertExpectations(t)
}

func TestGetBalance_Success(t *testing.T) {
	w := makeWallet(200)
	mockSvc := new(MockWalletService)
	mockSvc.On("GetBalance", mock.Anything, w.ID).Return(w, nil)

	req := httptest.NewRequest(http.MethodGet, "/wallets/"+w.ID.String(), nil)
	rec := httptest.NewRecorder()

	setupRouter(mockSvc).ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	resp := decodeBody(t, rec)
	assert.Equal(t, w.ID.String(), resp["id"])
	assert.InDelta(t, 200.0, resp["balance"], 0.001)
	mockSvc.AssertExpectations(t)
}

func TestGetBalance_InvalidWalletID(t *testing.T) {
	mockSvc := new(MockWalletService)

	req := httptest.NewRequest(http.MethodGet, "/wallets/not-a-uuid", nil)
	rec := httptest.NewRecorder()

	setupRouter(mockSvc).ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	mockSvc.AssertNotCalled(t, "GetBalance")
}

func TestGetBalance_WalletNotFound(t *testing.T) {
	walletID := uuid.New()
	mockSvc := new(MockWalletService)
	mockSvc.On("GetBalance", mock.Anything, walletID).
		Return(repository.Wallet{}, walletSvc.ErrWalletNotFound)

	req := httptest.NewRequest(http.MethodGet, "/wallets/"+walletID.String(), nil)
	rec := httptest.NewRecorder()

	setupRouter(mockSvc).ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
	resp := decodeBody(t, rec)
	assert.Equal(t, walletSvc.ErrWalletNotFound.Error(), resp["error"])
	mockSvc.AssertExpectations(t)
}

func TestGetBalance_ServiceError(t *testing.T) {
	walletID := uuid.New()
	mockSvc := new(MockWalletService)
	mockSvc.On("GetBalance", mock.Anything, walletID).
		Return(repository.Wallet{}, errors.New("db error"))

	req := httptest.NewRequest(http.MethodGet, "/wallets/"+walletID.String(), nil)
	rec := httptest.NewRecorder()

	setupRouter(mockSvc).ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	resp := decodeBody(t, rec)
	assert.Equal(t, "internal server error", resp["error"])
	mockSvc.AssertExpectations(t)
}
