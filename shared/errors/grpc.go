package errors

import (
	"encoding/json"

	commonpb "github.com/MamangRust/microservice-point-of-sale-pb/common"
)

func GrpcErrorToJson(err *commonpb.ErrorResponse) string {
	jsonData, _ := json.Marshal(err)
	return string(jsonData)
}
