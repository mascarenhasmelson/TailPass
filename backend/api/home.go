package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"port/utils"
	"time"
)

func GetIspInfo(ctx context.Context) (utils.Home, utils.Error) {
	var out utils.Home
	var e utils.Error

	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}

	// Prepare request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://ip-api.com/json", nil)
	if err != nil {
		e.Message = fmt.Sprintf("failed to build request: %v", err)
		out.Interstatus = "down"
		return out, e
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		e.Message = fmt.Sprintf("request error: %v", err)
		out.Interstatus = "down"
		return out, e
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		e.Message = fmt.Sprintf("unexpected status: %s", resp.Status)
		out.Interstatus = "down"
		return out, e
	}

	// Response model from ip-api.com
	var result struct {
		Status string `json:"status"`
		Query  string `json:"query"` // Public IP
		ISP    string `json:"isp"`   // ISP name
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		e.Message = fmt.Sprintf("decode error: %v", err)
		out.Interstatus = "down"
		return out, e
	}

	if result.Status != "success" {
		e.Message = fmt.Sprintf("api error: %s", result.Status)
		out.Interstatus = "down"
		return out, e
	}

	out.Status = "success"
	out.PublicIP = result.Query
	out.ISPInfo = result.ISP
	out.Interstatus = "up"

	return out, e
}

func HandleGetISPInfo(ctx context.Context, w http.ResponseWriter) {

	out, errObj := GetIspInfo(ctx)
	w.Header().Set("Content-Type", "application/json")
	if errObj.Message != "" {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errObj)
		return
	}
	json.NewEncoder(w).Encode(out)
}
