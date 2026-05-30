package controllers

// HealthController handles the health check endpoint.
type HealthController struct {
	BaseController
}

// Get handles GET /api/v1/health
func (c *HealthController) Get() {
	c.respondSuccess("Server is running", nil)
}