package bussiness

import (
	"tonguediag/bussiness/diag"
	"tonguediag/utils"

	"github.com/gin-gonic/gin"
)

//Init bussiness module init
func Init(config *utils.Config, r *gin.Engine) {
	diag.Init(config, r)
}
