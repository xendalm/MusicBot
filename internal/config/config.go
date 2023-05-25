package config

type PostgresConfig struct {
	PostgresUser     string `env:"POSTGRES_USER"`
	PostgresPassword string `env:"POSTGRES_PASSWORD"`
	PostgresDb       string `env:"POSTGRES_DB"`
	PostgresPort     int    `env:"POSTGRES_PORT"`
	PostgresDbHost   string `env:"POSTGRES_DB_HOST"`
}

type BotConfig struct {
	Token        string `env:"DISCORD_TOKEN"`
	NodeName     string `env:"NODE_NAME"`
	NodeHost     string `env:"NODE_HOST"`
	NodePort     string `env:"NODE_PORT"`
	NodePassword string `env:"NODE_PASSWORD"`
	NodeSecure   bool   `env:"NODE_SECURE"`
}
