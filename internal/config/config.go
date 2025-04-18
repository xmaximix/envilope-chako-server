package config

import "time"

type Config struct {
	Server   ServerConfig `mapstructure:"server"`
	Database DBConfig     `mapstructure:"db"`
	Auth     AuthConfig   `mapstructure:"auth"`
}

type ServerConfig struct {
	Port         string        `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

type DBConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	SSLMode  string `mapstructure:"sslmode"`
}

type AuthConfig struct {
	JWTSecret       string        `mapstructure:"auth.jwt_secret"`
	AccessTokenTTL  time.Duration `mapstructure:"auth.access_ttl"`
	RefreshTokenTTL time.Duration `mapstructure:"auth.refresh_ttl"`
	SMTPHost        string        `mapstructure:"auth.smtp_host"`
	SMTPPort        int           `mapstructure:"auth.smtp_port"`
	SMTPUser        string        `mapstructure:"auth.smtp_user"`
	SMTPPass        string        `mapstructure:"auth.smtp_pass"`
	SMTPFrom        string        `mapstructure:"auth.smtp_from"`
}
