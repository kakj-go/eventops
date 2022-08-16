package main

import "github.com/gin-gonic/gin"

type Router interface {
	Router(*gin.RouterGroup)
}
