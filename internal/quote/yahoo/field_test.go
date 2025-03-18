package yahoo_test

import (
	"encoding/json"
	"testing"

	. "github.com/achannarasappa/ticker/v4/internal/quote/yahoo"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type TestStruct struct {
	Field ResponseFieldString `json:"field"`
}

var _ = Describe("ResponseFieldString", func() {
	Describe("JSON unmarshaling", func() {
		It("should properly unmarshal string values", func() {
			jsonStr := `{"field": "123.45"}`
			var result TestStruct
			err := json.Unmarshal([]byte(jsonStr), &result)

			Expect(err).ToNot(HaveOccurred())
			Expect(result.Field.Raw).To(Equal("123.45"))
			Expect(result.Field.Fmt).To(Equal("123.45"))
		})

		It("should properly unmarshal object values", func() {
			jsonObj := `{"field": {"raw": "100-120", "fmt": "100 to 120"}}`
			var result TestStruct
			err := json.Unmarshal([]byte(jsonObj), &result)

			Expect(err).ToNot(HaveOccurred())
			Expect(result.Field.Raw).To(Equal("100-120"))
			Expect(result.Field.Fmt).To(Equal("100 to 120"))
		})

		It("should handle empty data", func() {
			var result TestStruct
			err := json.Unmarshal([]byte(""), &result)

			Expect(err).To(HaveOccurred())
		})

		It("should handle invalid string JSON", func() {
			var result TestStruct
			err := json.Unmarshal([]byte(`{"field": invalid"`), &result)

			Expect(err).To(HaveOccurred())
		})

		It("should handle empty string", func() {
			jsonStr := `{"field": ""}`
			var result TestStruct
			err := json.Unmarshal([]byte(jsonStr), &result)

			Expect(err).ToNot(HaveOccurred())
			Expect(result.Field.Raw).To(Equal(""))
			Expect(result.Field.Fmt).To(Equal(""))
		})

		It("should handle null values", func() {
			jsonStr := `{"field": null}`
			var result TestStruct
			err := json.Unmarshal([]byte(jsonStr), &result)

			Expect(err).ToNot(HaveOccurred())
			Expect(result.Field.Raw).To(Equal(""))
			Expect(result.Field.Fmt).To(Equal(""))
		})
	})
})

func TestResponseFieldStringUnmarshalling(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		wantRaw  string
		wantFmt  string
		wantErr  bool
	}{
		{
			name:    "string format",
			json:    `{"field": "42.5-43.2"}`,
			wantRaw: "42.5-43.2",
			wantFmt: "42.5-43.2",
			wantErr: false,
		},
		{
			name:    "object format",
			json:    `{"field": {"raw": "42.5-43.2", "fmt": "42.5 to 43.2"}}`,
			wantRaw: "42.5-43.2",
			wantFmt: "42.5 to 43.2",
			wantErr: false,
		},
		{
			name:    "empty string",
			json:    `{"field": ""}`,
			wantRaw: "",
			wantFmt: "",
			wantErr: false,
		},
		{
			name:    "null value",
			json:    `{"field": null}`,
			wantRaw: "",
			wantFmt: "",
			wantErr: false,
		},
		{
			name:    "malformed object",
			json:    `{"field": {"raw": "42.5-43.2"}}`,
			wantRaw: "42.5-43.2",
			wantFmt: "",
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			json:    `{"field": {`,
			wantRaw: "",
			wantFmt: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result TestStruct
			err := json.Unmarshal([]byte(tt.json), &result)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr {
				if result.Field.Raw != tt.wantRaw {
					t.Errorf("Raw = %v, want %v", result.Field.Raw, tt.wantRaw)
				}
				
				if result.Field.Fmt != tt.wantFmt {
					t.Errorf("Fmt = %v, want %v", result.Field.Fmt, tt.wantFmt)
				}
			}
		})
	}
}
