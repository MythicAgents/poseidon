package InteractiveTask

import "testing"

func TestFileEditorRangeDoesNotExpandTerminalValues(t *testing.T) {
	for value := 0; value < int(interactiveEnd); value++ {
		if !IsValid(value) {
			t.Fatalf("existing value %d should be valid", value)
		}
	}
	for value := int(interactiveEnd); value < int(FileEditorRequest); value++ {
		if IsValid(value) {
			t.Fatalf("reserved value %d should be invalid", value)
		}
	}
	for _, value := range []MessageType{FileEditorRequest, FileEditorResponse, FileEditorError} {
		if !IsValid(int(value)) {
			t.Fatalf("file editor value %d should be valid", value)
		}
	}
	if IsValid(103) {
		t.Fatal("unused file editor value should be invalid")
	}
}
