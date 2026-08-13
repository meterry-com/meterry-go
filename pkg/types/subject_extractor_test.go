package types

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestSubjectExtractorJSONRoundTrip(t *testing.T) {
	subject := SubjectExtractor{
		DefaultType: "user",
		TypePath:    "$.subject.type",
		DefaultID:   "anonymous",
		IDPath:      "$.subject.id",
		Required:    true,
	}
	raw, err := json.Marshal(subject)
	if err != nil {
		t.Fatal(err)
	}
	var decoded SubjectExtractor
	if err = json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded, subject) {
		t.Fatalf("decoded = %#v, want %#v", decoded, subject)
	}
}
