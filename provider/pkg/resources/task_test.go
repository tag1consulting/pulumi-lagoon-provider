package resources

import "testing"

func TestTaskArgumentsDiffer(t *testing.T) {
	tests := []struct {
		name string
		a    *[]TaskArgumentInput
		b    *[]TaskArgumentInput
		want bool
	}{
		{
			name: "both nil",
			a:    nil,
			b:    nil,
			want: false,
		},
		{
			name: "both empty slices",
			a:    &[]TaskArgumentInput{},
			b:    &[]TaskArgumentInput{},
			want: false,
		},
		{
			name: "nil vs empty slice",
			a:    nil,
			b:    &[]TaskArgumentInput{},
			want: false,
		},
		{
			name: "empty slice vs nil",
			a:    &[]TaskArgumentInput{},
			b:    nil,
			want: false,
		},
		{
			name: "different lengths",
			a: &[]TaskArgumentInput{
				{Name: "arg1", DisplayName: "Arg 1", Type: "string"},
			},
			b:    &[]TaskArgumentInput{},
			want: true,
		},
		{
			name: "same length same content",
			a: &[]TaskArgumentInput{
				{Name: "arg1", DisplayName: "Arg 1", Type: "string"},
			},
			b: &[]TaskArgumentInput{
				{Name: "arg1", DisplayName: "Arg 1", Type: "string"},
			},
			want: false,
		},
		{
			name: "different Name",
			a: &[]TaskArgumentInput{
				{Name: "arg1", DisplayName: "Arg", Type: "string"},
			},
			b: &[]TaskArgumentInput{
				{Name: "arg2", DisplayName: "Arg", Type: "string"},
			},
			want: true,
		},
		{
			name: "different DisplayName",
			a: &[]TaskArgumentInput{
				{Name: "arg1", DisplayName: "Arg One", Type: "string"},
			},
			b: &[]TaskArgumentInput{
				{Name: "arg1", DisplayName: "Arg Two", Type: "string"},
			},
			want: true,
		},
		{
			name: "Type case insensitive - STRING vs string",
			a: &[]TaskArgumentInput{
				{Name: "arg1", DisplayName: "Arg", Type: "STRING"},
			},
			b: &[]TaskArgumentInput{
				{Name: "arg1", DisplayName: "Arg", Type: "string"},
			},
			want: false,
		},
		{
			name: "different Type values - string vs number",
			a: &[]TaskArgumentInput{
				{Name: "arg1", DisplayName: "Arg", Type: "string"},
			},
			b: &[]TaskArgumentInput{
				{Name: "arg1", DisplayName: "Arg", Type: "number"},
			},
			want: true,
		},
		{
			name: "multiple elements all same",
			a: &[]TaskArgumentInput{
				{Name: "a", DisplayName: "A", Type: "string"},
				{Name: "b", DisplayName: "B", Type: "number"},
			},
			b: &[]TaskArgumentInput{
				{Name: "a", DisplayName: "A", Type: "string"},
				{Name: "b", DisplayName: "B", Type: "number"},
			},
			want: false,
		},
		{
			name: "multiple elements second differs",
			a: &[]TaskArgumentInput{
				{Name: "a", DisplayName: "A", Type: "string"},
				{Name: "b", DisplayName: "B", Type: "number"},
			},
			b: &[]TaskArgumentInput{
				{Name: "a", DisplayName: "A", Type: "string"},
				{Name: "c", DisplayName: "C", Type: "number"},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := taskArgumentsDiffer(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("taskArgumentsDiffer() = %v, want %v", got, tt.want)
			}
		})
	}
}
