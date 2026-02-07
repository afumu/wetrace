package strategy

import (
	"testing"
)

func TestV4_Identify(t *testing.T) {
	s := NewV4()

	tests := []struct {
		filename      string
		expectedType  GroupType
		expectedIndex string
		expectMatch   bool
	}{
		{"message_0.db", Message, "0", true},
		{"message.db", Message, "", true},
		{"message_12.db", Message, "12", true},
		{"contact.db", Contact, "", true},
		{"hardlink.db", Image, "", true},
		{"media_1.db", Voice, "1", true},
		{"Random.txt", Unknown, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			meta, match := s.Identify(tt.filename)
			if match != tt.expectMatch {
				t.Errorf("match: expected %v, got %v", tt.expectMatch, match)
			}
			if match {
				if meta.Type != tt.expectedType {
					t.Errorf("type: expected %v, got %v", tt.expectedType, meta.Type)
				}
				if meta.Index != tt.expectedIndex {
					t.Errorf("index: expected '%v', got '%v'", tt.expectedIndex, meta.Index)
				}
			}
		})
	}
}
