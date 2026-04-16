package dto

import "testing"

func TestNormalizeUploadPublicPath(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "empty",
			in:   "",
			want: "",
		},
		{
			name: "relative models path",
			in:   "models/a.obj",
			want: "/uploads/models/a.obj",
		},
		{
			name: "already prefixed",
			in:   "/uploads/models/a.obj",
			want: "/uploads/models/a.obj",
		},
		{
			name: "uploads prefix without slash",
			in:   "uploads/models/a.obj",
			want: "/uploads/models/a.obj",
		},
		{
			name: "absolute local path kept",
			in:   "/data/models/a.obj",
			want: "/data/models/a.obj",
		},
		{
			name: "http url kept",
			in:   "https://cdn.example.com/models/a.obj",
			want: "https://cdn.example.com/models/a.obj",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := NormalizeUploadPublicPath(tc.in)
			if got != tc.want {
				t.Fatalf("NormalizeUploadPublicPath(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
