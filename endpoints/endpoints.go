package endpoints

import "net/http"

// RequestHandler - declares a type for endpoint functions
type RequestHandler = func(w http.ResponseWriter, r *http.Request)