package conf

import (
	"testing"
)

func TestLoggingConfigureDefault(t *testing.T) {
	s := DefaultSettings()
	ConfigureLogging(s)
}

func TestLoggingConfigureJSON(t *testing.T) {
	s := DefaultSettings()
	s.Logging.Format = "json"
	s.Logging.Level = "DEBUG"
	ConfigureLogging(s)
}

func TestLoggingConfigureWarnLevel(t *testing.T) {
	s := DefaultSettings()
	s.Logging.Level = "WARN"
	ConfigureLogging(s)
}

func TestLoggingConfigureErrorLevel(t *testing.T) {
	s := DefaultSettings()
	s.Logging.Level = "ERROR"
	ConfigureLogging(s)
}

func TestLoggingConfigureInvalidLevel(t *testing.T) {
	s := DefaultSettings()
	s.Logging.Level = "INVALID"
	ConfigureLogging(s)
}