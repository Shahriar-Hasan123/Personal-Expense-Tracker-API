package controllers

import (
	beego "github.com/beego/beego/v2/server/web"
)

// BaseController embeds beego.Controller and provides shared response helpers.
type BaseController struct {
	beego.Controller
}

// respondJSON sends a uniform JSON response.
func (c *BaseController) respondJSON(status int, success bool, message string, data interface{}) {
	c.Ctx.Output.SetStatus(status)
	body := map[string]interface{}{
		"success": success,
		"message": message,
	}
	if data != nil {
		body["data"] = data
	}
	c.Data["json"] = body
	c.ServeJSON()
}

// respondSuccess sends a 200 OK success response.
func (c *BaseController) respondSuccess(message string, data interface{}) {
	c.respondJSON(200, true, message, data)
}

// respondCreated sends a 201 Created success response.
func (c *BaseController) respondCreated(message string, data interface{}) {
	c.respondJSON(201, true, message, data)
}

// respondBadRequest sends a 400 Bad Request error response.
func (c *BaseController) respondBadRequest(message string) {
	c.respondJSON(400, false, message, nil)
}

// respondUnauthorized sends a 401 Unauthorized error response.
func (c *BaseController) respondUnauthorized(message string) {
	c.respondJSON(401, false, message, nil)
}

// respondConflict sends a 409 Conflict error response.
func (c *BaseController) respondConflict(message string) {
	c.respondJSON(409, false, message, nil)
}

// respondInternalError sends a 500 Internal Server Error response.
func (c *BaseController) respondInternalError(message string) {
	c.respondJSON(500, false, message, nil)
}