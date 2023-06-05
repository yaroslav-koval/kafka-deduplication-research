//nolint:errchkjson, gomnd, wsl
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"kafka-polygon/pkg/cerror/v2"
	"kafka-polygon/pkg/log"
)

func main() {
	log.SetGlobalLogLevel("info")
	commonError()
	validationError()
	httpError()
	postgresError()
}

var _bgCtx = context.Background()

// use when error is created as a result of some business logic
// or after receiving unexpected error from external package (e.g. json.Unmarshal).
func commonError() {
	fmt.Println("============ Common error ============")
	// error in logs
	err := cerror.NewF(_bgCtx, cerror.KindConflict, "some conflict").LogError()
	// {
	// 	"level":"error",
	// 	"error":"some conflict",
	// 	"errorKind":"conflict",
	// 	"errorCode":409,
	// 	"operations":[
	// 	   "main.commonError:23",
	// 	   "main.main:16",
	// 	   "runtime.main:250",
	// 	   "runtime.goexit:1594"
	// 	],
	// 	"time":"2023-02-13T11:49:53+02:00"
	//  }

	// error as response
	b, _ := json.Marshal(cerror.BuildErrorHTTPResponse(err))

	fmt.Println(string(b))
	// {
	// 	"error":{
	// 	   "message":"some conflict",
	// 	   "kind":"conflict"
	// 	}
	// }
}

// use for errors that represent business validations.
// Especially when each field has it's own error message.
func validationError() {
	fmt.Println("============ Validation error ============")
	// error in logs
	err := cerror.NewValidationError(_bgCtx, map[string]string{"name": "min length"}).LogError()
	// {
	// 	"level":"error",
	// 	"error":"name:min length",
	// 	"errorKind":"validation",
	// 	"errorCode":422,
	// 	"payload":{
	// 	   "name":"min length"
	// 	},
	// 	"operations":[
	// 	   "main.validationError:54",
	// 	   "main.main:16",
	// 	   "runtime.main:250",
	// 	   "runtime.goexit:1594"
	// 	],
	// 	"time":"2023-02-13T11:55:54+02:00"
	//  }

	// error as response
	b, _ := json.Marshal(cerror.BuildErrorHTTPResponse(err))

	fmt.Println(string(b))
	// {
	// 	"error":{
	// 	   "message":"some fields are invalid",
	// 	   "kind":"validation",
	// 	   "errors":{
	// 		  "name":{
	// 			 "message":"min length"
	// 		  }
	// 	   }
	// 	}
	//  }
}

// use to represent errors after http requests (4xx, 5xx statuses)
func httpError() {
	fmt.Println("============ HTTP error ============")
	respCode := 422
	respBody := map[string]interface{}{"request": "invalid"}

	// error in logs
	err := cerror.NewF(
		_bgCtx,
		cerror.KindExternalAPI, // or cerror.KindInternalAPI for communication with our own services
		"bad http response",
	).
		// code here shouldn't be mapped from respCode.
		// Instead it should be difined based on a concrete business logic.
		// 409 was chosen just as an example.
		// WithHTTPCode is optional, it will be used in logs and will be a preferable
		// value for a http response status.
		WithHTTPCode(409).
		WithAttributesHTTP(cerror.ErrAttributesValuesHTTP{Code: respCode}). // here we use the code from the response
		WithPayload(respBody).
		LogError()
	// {
	// 	"level":"error",
	// 	"error":"bad http response",
	// 	"errorKind":"external_api",
	// 	"errorCode":409,
	// 	"errorAttributes":{
	// 	   "type":"api",
	// 	   "subtype":"http",
	// 	   "code":422
	// 	},
	// 	"payload":{
	// 	   "request":"invalid"
	// 	},
	// 	"operations":[
	// 	   "main.httpError:101",
	// 	   "main.main:17",
	// 	   "runtime.main:250",
	// 	   "runtime.goexit:1594"
	// 	],
	// 	"time":"2023-02-13T13:19:00+02:00"
	//  }

	// error as response
	b, _ := json.Marshal(cerror.BuildErrorHTTPResponse(err))

	fmt.Println(string(b))
	// {
	// 	"error":{
	// 	   "message":"bad http response",
	// 	   "kind":"external_api",
	// 	   "attributes":{
	// 		  "type":"api",
	// 		  "subtype":"http",
	// 		  "code":422
	// 	   }
	// 	}
	//  }
}

// use to represent errors during interaction with non-http resources(e.g. db, msg brokers, etc.).
// postgres is used as example.
func postgresError() {
	fmt.Println("============ Postgres error ============")
	originalError := fmt.Errorf("42501: insufficient privilege")

	// error in logs
	err := cerror.New(
		_bgCtx,
		cerror.KindInternalResource, // or cerror.KindExternal resource for communication with third party resources
		originalError,
	).
		WithAttributesPostgres(cerror.ErrAttributesValuesPostgres{Err: originalError}).
		LogError()
	// {
	// 	"level":"error",
	// 	"error":"42501: insufficient privilege",
	// 	"errorKind":"internal_resource",
	// 	"errorCode":500,
	// 	"errorAttributes":{
	// 	   "type":"db",
	// 	   "subtype":"postgres",
	// 	   "codeStr":"permission"
	// 	},
	// 	"operations":[
	// 	   "main.postgresError:162",
	// 	   "main.main:18",
	// 	   "runtime.main:250",
	// 	   "runtime.goexit:1594"
	// 	],
	// 	"time":"2023-02-13T13:38:49+02:00"
	//  }

	// error as response
	b, _ := json.Marshal(cerror.BuildErrorHTTPResponse(err))

	fmt.Println(string(b))
	// {
	// 	"error":{
	// 	   "message":"42501: insufficient privilege",
	// 	   "kind":"internal_resource",
	// 	   "attributes":{
	// 		  "type":"db",
	// 		  "subtype":"postgres",
	// 		  "codeStr":"permission"
	// 	   }
	// 	}
	//  }
}
