package id

import (
	"monkey/logger"
	"monkey/pd/storage"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var (
	ihloger, _ = logger.GetLoggerManager().GetLogger(logger.MainTag)
)

type IdHandler struct {
	sequenceStorage storage.SequenceStorage
}

func NewIdHandler(sequenceStorage storage.SequenceStorage) *IdHandler {
	return &IdHandler{
		sequenceStorage: sequenceStorage,
	}
}

func (ih *IdHandler) NewServerId(c *gin.Context) {
	sequenceType := c.DefaultPostForm("sequenceType", "server")
	resp, err := ih.sequenceStorage.NewSequence(sequenceType, 1)
	if err != nil {
		ihloger.Error("NewServerId failed ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (ih *IdHandler) NewSequence(c *gin.Context) {
	sequenceType := c.DefaultPostForm("sequenceType", "server")
	step, err := strconv.ParseUint(c.DefaultPostForm("step", "100"), 10, 64)
	if err != nil {
		ihloger.Error("NewSequence failed ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "step must be a number"})
		return
	}
	resp, err := ih.sequenceStorage.NewSequence(sequenceType, step)
	if err != nil {
		ihloger.Error("NewSequence failed ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
