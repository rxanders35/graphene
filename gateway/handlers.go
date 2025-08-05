package gateway

import "github.com/gin-gonic/gin"

type GatewayHandler struct{} //does this need to exist to hold a dependency?

// this is the API surface of the entire app
func (g *GatewayHandler) Write(c *gin.Context) {}
func (g *GatewayHandler) Read(c *gin.Context)  {}
