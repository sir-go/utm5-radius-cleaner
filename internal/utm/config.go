package utm

type Config struct {
	Url      string `toml:"api_url"`
	Prefix   string `toml:"prefix"`
	Username string `toml:"username"`
	Password string `toml:"password"`
}
