package wallet

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	walletService "tryingMicro/OrderAccepter/internal/service/wallet"
	"tryingMicro/OrderAccepter/package/logger"
)

type WalletController interface {
	ProcessOperation(c *gin.Context)
	GetBalance(c *gin.Context)
}

type walletController struct {
	service walletService.WalletService
	log     logger.Logger
}

func New(service walletService.WalletService, log logger.Logger) WalletController {
	return &walletController{
		service: service,
		log:     log,
	}
}

type operationRequest struct {
	OperationType string  `json:"operationType" binding:"required"`
	Amount        float64 `json:"amount"        binding:"required,gt=0"`
}

func (wc *walletController) ProcessOperation(c *gin.Context) {
	walletID, err := uuid.Parse(c.Param("walletId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wallet id"})
		return
	}

	var req operationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := wc.service.ProcessOperation(c.Request.Context(), walletID, req.OperationType, req.Amount)
	if err != nil {
		switch {
		case errors.Is(err, walletService.ErrWalletNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, walletService.ErrInsufficientFunds),
			errors.Is(err, walletService.ErrInvalidOperation):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			wc.log.Error("ProcessOperation", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

func (wc *walletController) GetBalance(c *gin.Context) {
	walletID, err := uuid.Parse(c.Param("walletId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wallet id"})
		return
	}

	result, err := wc.service.GetBalance(c.Request.Context(), walletID)
	if err != nil {
		if errors.Is(err, walletService.ErrWalletNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		wc.log.Error("GetBalance", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, result)
}
