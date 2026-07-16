package httpapi

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReadJSONRejectsUnknownFields(t *testing.T) {
	req := httptest.NewRequest("POST", "/", strings.NewReader(`{"event_id":"e1","user_id":"u1","x":1}`))
	var got bookingReq
	if err := readJSON(req, &got); err == nil {
		t.Fatal("expected unknown field error")
	}
}
