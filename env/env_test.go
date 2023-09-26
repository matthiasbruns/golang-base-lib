package env

import (
	"os"
	"testing"
)

func TestGetEnvWithFallback(t *testing.T) {
	type args struct {
		env      string
		fallback string
	}
	tests := []struct {
		name   string
		args   args
		want   string
		envVal string
	}{
		{
			name: "should return the provided env value and not the fallback",
			args: args{
				env:      "TEST_ENV",
				fallback: "nah",
			},
			want:   "env value",
			envVal: "env value",
		},
		{
			name: "should return the fallback then the env value is empty",
			args: args{
				env:      "TEST_ENV",
				fallback: "nah",
			},
			want:   "nah",
			envVal: "",
		},
	}
	for _, tt := range tests {
		_ = os.Setenv("TEST_ENV", tt.envVal)

		t.Run(
			tt.name, func(t *testing.T) {
				if got := GetEnvWithFallback(tt.args.env, tt.args.fallback); got != tt.want {
					t.Errorf("GetEnvWithFallback() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func TestIsProd(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{name: "should return true when STAGE is equals prod", want: true},
		{name: "should return false when STAGE is not equals prod", want: false},
	}
	for _, tt := range tests {
		if tt.want {
			_ = os.Setenv("STAGE", "prod")
		} else {
			_ = os.Setenv("STAGE", "not prod")
		}

		t.Run(
			tt.name, func(t *testing.T) {
				if got := IsProd(); got != tt.want {
					t.Errorf("IsProd() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func TestIsLocal(t *testing.T) {
	tests := []struct {
		name   string
		want   bool
		envVal string
	}{
		{name: "should return true when STAGE is equals local", envVal: "local", want: true},
		{name: "should return true when STAGE is empty", envVal: "", want: true},
		{name: "should return false when STAGE is not empty and not local", envVal: "prod", want: false},
	}
	for _, tt := range tests {
		_ = os.Setenv("STAGE", tt.envVal)

		t.Run(
			tt.name, func(t *testing.T) {
				if got := IsLocal(); got != tt.want {
					t.Errorf("IsLocal() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}
