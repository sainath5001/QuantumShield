package apierror

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestErrorEnvelopeJSON(t *testing.T) {
	err := BadRequest(CodeInvalidRequest, "bad input").WithDetails(map[string]interface{}{
		"field": "address",
	})
	raw, marshalErr := json.Marshal(err.ToResponse())
	if marshalErr != nil {
		t.Fatal(marshalErr)
	}

	var got map[string]interface{}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	errorObj, ok := got["error"].(map[string]interface{})
	if !ok {
		t.Fatalf("missing error object: %s", raw)
	}
	if errorObj["code"] != string(CodeInvalidRequest) {
		t.Fatalf("code = %v", errorObj["code"])
	}
	if errorObj["message"] != "bad input" {
		t.Fatalf("message = %v", errorObj["message"])
	}
}

func TestHelperStatusCodes(t *testing.T) {
	if NotFound("x").HTTPStatus != http.StatusNotFound {
		t.Fatal("NotFound status")
	}
	if Internal("x").HTTPStatus != http.StatusInternalServerError {
		t.Fatal("Internal status")
	}
	if Timeout("x").HTTPStatus != http.StatusGatewayTimeout {
		t.Fatal("Timeout status")
	}
}
