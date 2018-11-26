package handler

import (
	"github.com/gin-gonic/gin"
)

type OutModel struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func ResponseJSON(c *gin.Context, code int, msg string, data interface{}) {
	c.Header("Access-Control-Allow-Origin", "*")
	out := &OutModel{
		Code: code,
		Msg:  msg,
		Data: data,
	}

	c.JSON(code, out)
}
