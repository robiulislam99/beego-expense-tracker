// Package controllers handles all incoming HTTP requests and sends responses.
// base.go provides shared response helper methods used by all controllers.
package controllers

import beego "github.com/beego/beego/v2/server/web"

// BaseController embeds beego.Controller and provides
// shared response helper methods for all other controllers.
type BaseController struct {
	beego.Controller
}

// ResponseData represents the standard API response structure
// used across all endpoints in the application.
type ResponseData struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// SendSuccess sends a successful JSON response with HTTP 200 status.
// It includes an optional data payload.
func (c *BaseController) SendSuccess(message string, data interface{}) {
	c.Ctx.Output.SetStatus(200)
	c.Data["json"] = ResponseData{
		Success: true,
		Message: message,
		Data:    data,
	}
	c.ServeJSON()
}

// SendCreated sends a successful JSON response with HTTP 201 status.
// Used when a new resource has been created successfully.
func (c *BaseController) SendCreated(message string, data interface{}) {
	c.Ctx.Output.SetStatus(201)
	c.Data["json"] = ResponseData{
		Success: true,
		Message: message,
		Data:    data,
	}
	c.ServeJSON()
}

// SendError sends an error JSON response with the given HTTP status code.
// Used for all failure responses across the API.
func (c *BaseController) SendError(statusCode int, message string) {
	c.Ctx.Output.SetStatus(statusCode)
	c.Data["json"] = ResponseData{
		Success: false,
		Message: message,
	}
	c.ServeJSON()
}