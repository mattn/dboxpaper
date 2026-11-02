package main

import (
	"encoding/json"
	"testing"
)

func TestAPIArg(t *testing.T) {
	in := map[string]interface{}{"path": "/日本語/😀.paper"}
	got, err := apiArg(in)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"path":"/\u65e5\u672c\u8a9e/\ud83d\ude00.paper"}`
	if got != want {
		t.Fatalf("want %s but got %s", want, got)
	}
	var v map[string]interface{}
	if err := json.Unmarshal([]byte(got), &v); err != nil || v["path"] != in["path"] {
		t.Fatalf("round trip failed: %v %v", v, err)
	}
}
