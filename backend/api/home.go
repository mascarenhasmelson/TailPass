package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/mascarenhasmelson/TailPass/utils"
)

func GetIspInfo(ctx context.Context) (utils.IPInfoRaw, utils.Error) {
	var errResp utils.Error
	var result utils.IPInfoRaw
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://ipinfo.io/json", nil)
	if err != nil {
		errResp.Message = "Failed to create request"
		return result, errResp
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		errResp.Message = "Failed to fetch IP info"
		return result, errResp
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		errResp.Message = "Failed to decode JSON"
		return result, errResp
	}

	return utils.IPInfoRaw{
		IP:  result.IP,
		Org: result.Org,
	}, utils.Error{}

}
func HandleGetISPInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	out, errObj := GetIspInfo(ctx)

	w.Header().Set("Content-Type", "application/json")

	if errObj.Message != "" {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errObj)
		return
	}

	json.NewEncoder(w).Encode(out)
}

// func HandleGetISPInfo(ctx context.Context, w http.ResponseWriter) {

// 	out, errObj := GetIspInfo(ctx)
// 	w.Header().Set("Content-Type", "application/json")
// 	if errObj.Message != "" {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		json.NewEncoder(w).Encode(errObj)
// 		return
// 	}
// 	json.NewEncoder(w).Encode(out)
// }
