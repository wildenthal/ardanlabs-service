package config

import (
	"fmt"
	"os"
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
				otlpHostKey:        "OTEL_EXPORTER_OTLP_ENDPOINT",
				apiHostKey:         "API_HOST",
				debugHostKey:       "DEBUG_HOST",
			},
			want: &Config{
				Build:           "test-build",
				Desc:            "",
				APIHost:         "API_HOST",
				DebugHost:       "DEBUG_HOST",
				OTLPHost:        "OTEL_EXPORTER_OTLP_ENDPOINT",
				ReadTimeout:     137 * time.Second,
				WriteTimeout:    137 * time.Second,
				IdleTimeout:     137 * time.Second,
				ShutdownTimeout: 137 * time.Second,
			},
		},
		{
			name: "OK with default values",
			env: map[string]string{
				otlpHostKey: "OTEL_EXPORTER_OTLP_ENDPOINT",
			},
			want: &Config{
				Build:           "",
				Desc:            "",
				APIHost:         defaultApiHost,
				DebugHost:       defaultDebugHost,
				OTLPHost:        "OTEL_EXPORTER_OTLP_ENDPOINT",
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
			wantErr: fmt.Errorf(loadEnvVarError, readTimeoutKey, fmt.Errorf(`time: invalid duration "invalid"`)),
		},
		{
			name: "Invalid write timeout",
			env: map[string]string{
				writeTimeoutKey: "invalid",
			},
			wantErr: fmt.Errorf(loadEnvVarError, writeTimeoutKey, fmt.Errorf(`time: invalid duration "invalid"`)),
		},
		{
			name: "Invalid idle timeout",
			env: map[string]string{
				idleTimeoutKey: "invalid",
			},
			wantErr: fmt.Errorf(loadEnvVarError, idleTimeoutKey, fmt.Errorf(`time: invalid duration "invalid"`)),
		},
		{
			name: "Invalid shutdown timeout",
			env: map[string]string{
				shutdownTimeoutKey: "invalid",
			},
			wantErr: fmt.Errorf(loadEnvVarError, shutdownTimeoutKey, fmt.Errorf(`time: invalid duration "invalid"`)),
		},
		{
			name:    "Missing OTLP host",
			wantErr: fmt.Errorf(missingEnvVarError, otlpHostKey),
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

			if (tcase.wantErr == nil) != (err == nil) {
				t.Errorf("want error: %v, got: %v", tcase.wantErr, err)
			} else if tcase.wantErr != nil && err != nil {
				if tcase.wantErr.Error() != err.Error() {
					t.Errorf("want error message: %v, got: %v", tcase.wantErr, err)
				}
			}

			if (tcase.want == nil) != (got == nil) {
				t.Errorf("want: %v, got: %v", tcase.want, got)
			} else if tcase.want != nil && got != nil {
				if *tcase.want != *got {
					t.Errorf("want build: %v, got: %v", tcase.want.Build, got.Build)
				}
			}
		})
	}
}
