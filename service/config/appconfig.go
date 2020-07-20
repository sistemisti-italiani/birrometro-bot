package config

type APPConfig struct {
	Group int64 `conf:"default:0"` // TODO: supports multiple groups
}
