package http

import (
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GatewayProxy forwards REST requests to module gRPC services.
// For MVP, it maintains a simple routing table of module_name -> grpc address.
type GatewayProxy struct {
	mu    sync.RWMutex
	conns map[string]*grpc.ClientConn // module_name -> connection
	log   *zap.Logger
}

func NewGatewayProxy(log *zap.Logger) *GatewayProxy {
	return &GatewayProxy{
		conns: make(map[string]*grpc.ClientConn),
		log:   log,
	}
}

// RegisterModule adds or updates a module's gRPC connection.
func (gw *GatewayProxy) RegisterModule(name, address string) error {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	gw.mu.Lock()
	if old, ok := gw.conns[name]; ok {
		old.Close()
	}
	gw.conns[name] = conn
	gw.mu.Unlock()
	gw.log.Info("module registered in gateway", zap.String("module", name), zap.String("address", address))
	return nil
}

// UnregisterModule removes a module's gRPC connection.
func (gw *GatewayProxy) UnregisterModule(name string) {
	gw.mu.Lock()
	if conn, ok := gw.conns[name]; ok {
		conn.Close()
		delete(gw.conns, name)
	}
	gw.mu.Unlock()
}

// GetConnection returns the gRPC client connection for a module.
func (gw *GatewayProxy) GetConnection(name string) (*grpc.ClientConn, bool) {
	gw.mu.RLock()
	defer gw.mu.RUnlock()
	conn, ok := gw.conns[name]
	return conn, ok
}

// ProxyHandler returns a Gin handler that routes requests to the appropriate module.
// Pattern: /api/{module}/* -> module gRPC service
func (gw *GatewayProxy) ProxyHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract module name from path: /api/{module}/...
		parts := strings.SplitN(strings.TrimPrefix(c.Request.URL.Path, "/api/"), "/", 2)
		if len(parts) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "module not specified"})
			return
		}
		moduleName := parts[0]

		_, ok := gw.GetConnection(moduleName)
		if !ok {
			c.JSON(http.StatusBadGateway, gin.H{"error": "module not available", "module": moduleName})
			return
		}

		// TODO: Implement full gRPC-JSON transcoding in later phases.
		// For now, return a placeholder indicating the module is reachable.
		c.JSON(http.StatusOK, gin.H{
			"message": "gateway proxy: module reachable",
			"module":  moduleName,
			"path":    c.Request.URL.Path,
		})
	}
}

// Close cleans up all gRPC connections.
func (gw *GatewayProxy) Close() {
	gw.mu.Lock()
	defer gw.mu.Unlock()
	for name, conn := range gw.conns {
		conn.Close()
		delete(gw.conns, name)
	}
}
