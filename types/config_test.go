package types

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestConfig_Parse(t *testing.T) {
	yamlFile, _ := ioutil.TempFile("/tmp", "*.yaml")
	jsonFile, _ := ioutil.TempFile("/tmp", "*.json")
	txtFile, _ := ioutil.TempFile("/tmp", "*.txt")

	defer func() {
		_ = os.Remove(yamlFile.Name())
		_ = os.Remove(jsonFile.Name())
		_ = os.Remove(txtFile.Name())
	}()

	_, err := jsonFile.Write([]byte(`{}`))
	require.NoError(t, err)

	t.Run("no file", func(t *testing.T) {
		cfg := &Cfg{
			path: "no-file",
		}
		require.Error(t, cfg.Parse())
	})

	t.Run("wrong file format", func(t *testing.T) {
		cfg := &Cfg{
			path: txtFile.Name(),
		}
		err := cfg.Parse()
		require.EqualError(t, err, fmt.Sprintf("unknown configuration format `%s`", txtFile.Name()))
	})

	t.Run("json config", func(t *testing.T) {
		cfg := &Cfg{
			path: jsonFile.Name(),
		}
		require.NoError(t, cfg.Parse())
	})
	t.Run("yaml config", func(t *testing.T) {
		cfg := &Cfg{
			path: yamlFile.Name(),
		}
		require.NoError(t, cfg.Parse())
	})
}

func TestConfig_Validate(t *testing.T) {
	t.Run("empty config", func(t *testing.T) {
		cfg := &Cfg{
			path: "",
		}
		err := cfg.Validate()
		require.EqualError(t, err, "no hosts for monitoring")
	})

	t.Run("host without URL", func(t *testing.T) {
		cfg := &Cfg{
			Hosts: []Host{
				{
					Name: "no-host",
				},
			},
		}
		err := cfg.Validate()
		require.EqualError(t, err, "no url for no-host")
	})

	t.Run("intervals", func(t *testing.T) {
		t.Run("retry", func(t *testing.T) {
			cfg := &Cfg{
				Retry:        "",
				Timeout:      "1s",
				InitialDelay: "0s",
				Success: &Success{
					Code: []int{200},
				},
				Hosts: []Host{
					{
						Name: "test",
						URL:  "test",
					},
				},
			}

			cfg.Retry = "123s"
			require.NoError(t, cfg.Validate())
			require.Equal(t, "123s", cfg.Hosts[0].Retry)
			require.Equal(t, time.Second*123, cfg.Hosts[0].RetryInterval)
		})
		t.Run("timeout", func(t *testing.T) {
			cfg := &Cfg{
				Retry:        "1s",
				InitialDelay: "0s",
				Success: &Success{
					Code: []int{200},
				},
				Hosts: []Host{
					{
						Name: "test",
						URL:  "test",
					},
				},
			}

			cfg.Timeout = "123s"
			require.NoError(t, cfg.Validate())
			require.Equal(t, "123s", cfg.Hosts[0].Timeout)
			require.Equal(t, time.Second*123, cfg.Hosts[0].TimeoutInterval)
		})
		t.Run("initial delay", func(t *testing.T) {
			cfg := &Cfg{
				Retry:        "1s",
				Timeout:      "1s",
				InitialDelay: "0s",
				Success: &Success{
					Code: []int{200},
				},
				Hosts: []Host{
					{
						Name:         "test",
						URL:          "test",
						InitialDelay: "error",
					},
				},
			}

			require.Error(t, cfg.Validate())
			cfg.Hosts[0].InitialDelay = ""
			require.NoError(t, cfg.Validate())

			require.Equal(t, "0s", cfg.Hosts[0].InitialDelay)
			require.Equal(t, time.Second*0, cfg.Hosts[0].InitialDelayInterval)
		})
	})

	t.Run("success code", func(t *testing.T) {
		cfg := &Cfg{
			Retry:        "1s",
			Timeout:      "1s",
			InitialDelay: "0s",
			Success: &Success{
				Code: []int{200},
			},
			Hosts: []Host{
				{
					URL: "ok",
				},
			},
		}
		require.NoError(t, cfg.Validate())
		require.Equal(t, []int{200}, cfg.Hosts[0].Success.Code)
	})

	t.Run("headers", func(t *testing.T) {
		cfg := &Cfg{
			Retry:        "1s",
			Timeout:      "1s",
			InitialDelay: "0s",
			Success: &Success{
				Code: []int{200},
			},
			Headers: map[string]string{
				"key-1": "value-1",
				"key-2": "value-2",
				"key-3": "value-3",
			},
			Hosts: []Host{
				{
					URL: "ok",
					Headers: map[string]string{
						"key-2": "value-2-2",
					},
				},
			},
		}

		require.NoError(t, cfg.Validate())

		require.Len(t, cfg.Hosts[0].Headers, 3)

		require.Equal(t, "value-1", cfg.Hosts[0].Headers["key-1"])
		require.Equal(t, "value-2-2", cfg.Hosts[0].Headers["key-2"])
		require.Equal(t, "value-3", cfg.Hosts[0].Headers["key-3"])
	})
}
