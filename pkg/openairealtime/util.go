package openairealtime

import (
	"encoding/json"
	"fmt"
)

func prettyPrint(data map[string]interface{}) string {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error formatting JSON: %v", err)
	}
	return string(bytes)
}
