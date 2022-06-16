package jwt

//	{"secret":"123456789","method":"sha256","expire":"3600","excludes":["/a/**","/demo/**"]}

type Config struct {
	Secret   string   `json:"secret" yaml:"secret"`
	Method   string   `json:"method" yaml:"method"`
	Expire   int      `json:"expire" yaml:"expire"`
	Excludes []string `json:"excludes" yaml:"excludes"`
}
