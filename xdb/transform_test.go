package xdb

import (
	"testing"
)

func TestTranslateWithCallback(t *testing.T) {
	tests := []struct {
		name     string
		template string
		callback func(string) string
		expected string
	}{
		{
			name:     "simple @{var} pattern",
			template: "hello @{name}",
			callback: func(param string) string {
				if param == "name" {
					return "world"
				}
				return ""
			},
			expected: "hello world",
		},
		{
			name:     "simple {var} pattern",
			template: "hello {name}",
			callback: func(param string) string {
				if param == "name" {
					return "world"
				}
				return ""
			},
			expected: "hello world",
		},
		{
			name:     "only @ symbol without braces should stay",
			template: "price: @$100",
			callback: func(param string) string {
				return "replaced"
			},
			expected: "price: @$100",
		},
		{
			name:     "@ followed by non-brace should stay",
			template: "Hello @username!",
			callback: func(param string) string {
				return "world"
			},
			expected: "Hello @username!",
		},
		{
			name:     "double @@ symbol handling - AAA@@BBB pattern",
			template: "AAA@@BBB",
			callback: func(param string) string {
				return "replaced"
			},
			expected: "AAA@@BBB",
		},
		{
			name:     "triple @@@ symbol handling",
			template: "XXX@@@YYY",
			callback: func(param string) string {
				return "replaced"
			},
			expected: "XXX@@@YYY",
		},
		{
			name:     "mixed patterns - only braces work",
			template: "Hello @{name}, your score is {score}, and @notvar remains",
			callback: func(param string) string {
				switch param {
				case "name":
					return "Alice"
				case "score":
					return "95"
				default:
					return "unknown"
				}
			},
			expected: "Hello Alice, your score is 95, and @notvar remains",
		},
		{
			name:     "multiple occurrences of same variable",
			template: "@{name} said hello to {name}",
			callback: func(param string) string {
				if param == "name" {
					return "Bob"
				}
				return "unknown"
			},
			expected: "Bob said hello to Bob",
		},
		{
			name:     "complex template with special chars",
			template: "SELECT * FROM users WHERE name=@{user_name} AND age>{min_age}",
			callback: func(param string) string {
				switch param {
				case "user_name":
					return "john"
				case "min_age":
					return "18"
				default:
					return "unknown"
				}
			},
			expected: "SELECT * FROM users WHERE name=john AND age>18",
		},
		{
			name:     "empty variable names",
			template: "before @@{empty} and {empty}@ after",
			callback: func(param string) string {
				if param == "empty" {
					return "filled"
				}
				return ""
			},
			expected: "before @filled and filled@ after",
		},
		{
			name:     "variables with special chars inside",
			template: "Values: @{val_1}, {val-2}, @special3",
			callback: func(param string) string {
				return "replaced"
			},
			expected: "Values: replaced, replaced, @special3",
		},
		{
			name:     "no variables",
			template: "just plain text",
			callback: func(param string) string {
				return "replaced"
			},
			expected: "just plain text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TranslateCallback(tt.template, tt.callback)
			if result != tt.expected {
				t.Errorf("TranslateCallback(%q) = %q, want %q", tt.template, result, tt.expected)
			}
		})
	}
}

func TestTranslate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "basic @{var} substitution",
			template: "Hello @{name}",
			data:     map[string]interface{}{"name": "World"},
			expected: "Hello World",
		},
		{
			name:     "basic {var} substitution",
			template: "Hello {name}",
			data:     map[string]interface{}{"name": "World"},
			expected: "Hello World",
		},
		{
			name:     "non-brace @var should remain",
			template: "Price: @$100, User: @username",
			data:     map[string]interface{}{"username": "test"},
			expected: "Price: @$100, User: @username",
		},
		{
			name:     "mixed patterns",
			template: "Welcome @{name}, your score is {score}",
			data:     map[string]interface{}{"name": "Alice", "score": 95},
			expected: "Welcome Alice, your score is 95",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Translate(tt.template, tt.data)
			if result != tt.expected {
				t.Errorf("Translate(%q, %v) = %q, want %q", tt.template, tt.data, result, tt.expected)
			}
		})
	}
}