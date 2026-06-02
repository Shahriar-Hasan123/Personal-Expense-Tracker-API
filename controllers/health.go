package controllers

// HealthController handles the health check endpoint.
type HealthController struct {
	BaseController
}

// Get handles GET /api/v1/health
// @Title Health Check
// @Summary Check if server is running
// @Description Returns server running status
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{} "Server is running"
// @Router /health [get]
func (c *HealthController) Get() {
	c.respondSuccess("Server is running", nil)
}