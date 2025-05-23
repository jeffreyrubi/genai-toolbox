// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package databricks

import (
	"strings"
)

// ExtractWarehouseID extracts the warehouse ID from an HTTP path
// Examples:
//   - "/sql/1.0/warehouses/84c487324df116f0" -> "84c487324df116f0"
//   - "84c487324df116f0" -> "84c487324df116f0"
//   - "/" -> ""
func ExtractWarehouseID(httpPath string) string {
	if httpPath == "" {
		return ""
	}
	
	// Extract just the warehouse ID from the full path if needed
	if len(httpPath) > 0 && httpPath[0] == '/' {
		parts := []string{}
		for _, part := range strings.Split(httpPath, "/") {
			if part != "" {
				parts = append(parts, part)
			}
		}
		if len(parts) > 0 {
			return parts[len(parts)-1]
		}
		return ""
	}
	
	// Already just the ID
	return httpPath
}
