package config

type OAuthConfig struct {
	GitHub       OAuthProvider `koanf:"github"`
	Google       OAuthProvider `koanf:"google"`
	Yandex       OAuthProvider `koanf:"yandex"`
	CallbackBase string       `koanf:"callback_base"`
}

type OAuthProvider struct {
	ClientID     string `koanf:"client_id"`
	ClientSecret string `koanf:"client_secret"`
}
