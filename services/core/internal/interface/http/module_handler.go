package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/application/command"
	"github.com/HuynhHoangPhuc/myrmex/services/core/internal/application/query"
)

type ModuleHandler struct {
	registerHandler *command.RegisterModuleHandler
	listHandler     *query.ListModulesHandler
	gateway         *GatewayProxy
}

func NewModuleHandler(
	register *command.RegisterModuleHandler,
	list *query.ListModulesHandler,
	gateway *GatewayProxy,
) *ModuleHandler {
	return &ModuleHandler{
		registerHandler: register,
		listHandler:     list,
		gateway:         gateway,
	}
}

type registerModuleRequest struct {
	Name        string `json:"name" binding:"required"`
	Version     string `json:"version" binding:"required"`
	GRPCAddress string `json:"grpc_address" binding:"required"`
}

func (h *ModuleHandler) RegisterModule(c *gin.Context) {
	var req registerModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mod, err := h.registerHandler.Handle(c.Request.Context(), command.RegisterModuleCommand{
		Name: req.Name, Version: req.Version, GRPCAddress: req.GRPCAddress,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Register in gateway proxy
	if h.gateway != nil {
		_ = h.gateway.RegisterModule(mod.Name, mod.GRPCAddress)
	}

	c.JSON(http.StatusCreated, gin.H{
		"name": mod.Name, "version": mod.Version, "grpc_address": mod.GRPCAddress,
	})
}

func (h *ModuleHandler) ListModules(c *gin.Context) {
	modules, err := h.listHandler.Handle(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := make([]gin.H, len(modules))
	for i, m := range modules {
		result[i] = gin.H{
			"name": m.Name, "version": m.Version,
			"grpc_address": m.GRPCAddress, "health_status": string(m.HealthStatus),
		}
	}
	c.JSON(http.StatusOK, gin.H{"modules": result})
}

func (h *ModuleHandler) UnregisterModule(c *gin.Context) {
	name := c.Param("name")
	// TODO: call unregister command handler
	if h.gateway != nil {
		h.gateway.UnregisterModule(name)
	}
	c.JSON(http.StatusNoContent, nil)
}
