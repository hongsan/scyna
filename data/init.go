package scyna_data

import (
	"flag"
	"scyna"
)

func Init() {
	secret_ := flag.String("scyna.data", "123456", "Authenticate")
	managerUrl := flag.String("managerUrl", "http://127.0.0.1:8081", "Manager Url")
	flag.Parse()

	scyna.RemoteInit(scyna.RemoteConfig{
		ManagerUrl: *managerUrl,
		Name:       "scyna.data",
		Secret:     *secret_,
	})
	scyna.UseDirectLog(5)
}

func Release() {
	scyna.Release()
}
