package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	beegoContext "github.com/beego/beego/v2/server/web/context"
)

func TestHealthCheck(t *testing.T) {
	tests := []struct {
		name        string
		wantStatus  int
		wantSuccess bool
		wantMsg     string
	}{
		{
			name:        "health check returns 200 with correct body",
			wantStatus:  200,
			wantSuccess: true,
			wantMsg:     "Server is running",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
			rr := httptest.NewRecorder()

			ctx := beegoContext.NewContext()
			ctx.Reset(rr, req)

			c := &HealthController{}
			c.Ctx = ctx
			c.Data = map[interface{}]interface{}{}
			c.Get()

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rr.Code, tt.wantStatus)
			}

			var body map[string]interface{}
			if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
				t.Fatalf("failed to decode response: %v — body: %s", err, rr.Body.String())
			}

			success, ok := body["success"].(bool)
			if !ok || success != tt.wantSuccess {
				t.Errorf("success = %v, want %v", body["success"], tt.wantSuccess)
			}
			if body["message"] != tt.wantMsg {
				t.Errorf("message = %q, want %q", body["message"], tt.wantMsg)
			}
		})
	}
}