package xconfig

const (
	EnvLocal       = "local"
	EnvProduction  = "production" // this env only use in metric-related projects
	EnvTestnetDev  = "testnet-dev"
	EnvTestnetProd = "testnet-prod"
)

type Basic struct {
	Env     string `default:"local"`
	Debug   bool   `default:"false"`
	Release string `default:"local-debug"` // github build number, injected in image by github action
	// slack config is optional, if exists, it will send all warn/error log to slack
	SlackToken   string
	SlackChannel string `default:"C076H0HBZLZ"` // default is channel testnet-dev
}
