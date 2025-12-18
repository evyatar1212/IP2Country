package models

// IPLocation represents geographic information for an IP address
// In Go, structs are used to define data structures
// JSON tags tell Go how to convert this struct to/from JSON
type IPLocation struct {
	IP      string `json:"-" example:"-"`                      // The IP address (not included in JSON response)
	City    string `json:"city" example:"Mountain View"`       // City name
	Country string `json:"country" example:"United States"`    // Country name
}

// ErrorResponse is the standard error response format
// This is what we return when something goes wrong
type ErrorResponse struct {
	Error string `json:"error" example:"Invalid IP address format"` // Error message
}
