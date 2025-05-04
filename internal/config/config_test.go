package config

import (
	"errors"
	"os"
	"reflect"
	"testing"
	"time"
)

func Test_LoadConfig(t *testing.T) {
	for _, tcase := range []struct {
		name    string
		build   string
		env     map[string]string
		want    *Config
		wantErr error
	}{
		{
			name:  "OK",
			build: "test-build",
			env: map[string]string{
				readTimeoutKey:     "137s",
				writeTimeoutKey:    "137s",
				idleTimeoutKey:     "137s",
				shutdownTimeoutKey: "137s",
				apiHostKey:         "API_HOST",
				debugHostKey:       "DEBUG_HOST",
			},
			want: &Config{
				Build:           "test-build",
				Desc:            "",
				APIHost:         "API_HOST",
				DebugHost:       "DEBUG_HOST",
				ReadTimeout:     137 * time.Second,
				WriteTimeout:    137 * time.Second,
				IdleTimeout:     137 * time.Second,
				ShutdownTimeout: 137 * time.Second,
			},
		},
		{
			name: "OK with default values",
			want: &Config{
				Build:           "",
				Desc:            "",
				APIHost:         defaultApiHost,
				DebugHost:       defaultDebugHost,
				ReadTimeout:     defaultReadTimeout,
				WriteTimeout:    defaultWriteTimeout,
				IdleTimeout:     defaultIdleTimeout,
				ShutdownTimeout: defaultShutdownTimeout,
			},
		},
		{
			name: "Invalid read timeout",
			env: map[string]string{
				readTimeoutKey: "invalid",
			},
			wantErr: errInvalidDuration,
		},
		{
			name: "Invalid write timeout",
			env: map[string]string{
				writeTimeoutKey: "invalid",
			},
			wantErr: errInvalidDuration,
		},
		{
			name: "Invalid idle timeout",
			env: map[string]string{
				idleTimeoutKey: "invalid",
			},
			wantErr: errInvalidDuration,
		},
		{
			name: "Invalid shutdown timeout",
			env: map[string]string{
				shutdownTimeoutKey: "invalid",
			},
			wantErr: errInvalidDuration,
		},
	} {
		t.Run(tcase.name, func(t *testing.T) {
			for key, value := range tcase.env {
				if err := os.Setenv(key, value); err != nil {
					t.Fatalf("failed to set mock env variable: %v", err)
				}
				defer func(k string) {
					if err := os.Unsetenv(k); err != nil {
						t.Fatalf("failed to unset mock env variable: %v", err)
					}
				}(key)
			}
			got, err := LoadConfig(tcase.build)

			if !errors.Is(err, tcase.wantErr) {
				t.Errorf("unexpected error: want=%v, got=%v", tcase.wantErr, err)
			}

			if tcase.want != nil && !reflect.DeepEqual(tcase.want, got) {
				t.Errorf("unexpected config: want=%+v, got=%+v", tcase.want, got)
			}
		})
	}
}
