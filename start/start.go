package start

import (
	"liberty-town/node/app"
	"liberty-town/node/config"
	"liberty-town/node/config/arguments"
	"liberty-town/node/config/globals"
	"liberty-town/node/federations"
	"liberty-town/node/federations/federation"
	"liberty-town/node/federations/federation_network"
	"liberty-town/node/federations/federation_network/federation_network_sync"
	"liberty-town/node/federations/federation_serve"
	"liberty-town/node/gui"
	"liberty-town/node/network"
	"liberty-town/node/settings"
	"liberty-town/node/store"
	"os"
	"os/signal"
	"syscall"
)

func StartMainNow() (err error) {

	if !globals.MainStarted.CompareAndSwap(false, true) {
		return
	}

	//加载数据库
	if err = store.InitDB(); err != nil {
		return
	}
	globals.MainEvents.BroadcastEvent("main", "store initialized")

	//加载设置
	if err = settings.Load(); err != nil {
		return
	}

	if err = federations.InitializeFederations(); err != nil {
		return
	}

	//设置默认值
	if federation_serve.ServeFederation.Load() == nil {
		federations.FederationsDict.Range(func(key string, value *federation.Federation) bool {
			if err = federation_serve.SetServeFederation(value, false); err != nil {
				return true
			}
			return false
		})
	}

	//检查更新
	if config.NODE_CONSENSUS == config.NODE_CONSENSUS_TYPE_FULL {
		federation_network_sync.ContinuouslyDownloadFederationData()
		federation_network_sync.ContinuouslyDownloadFederation()
	}

	//检查新通知
	if config.NODE_CONSENSUS == config.NODE_CONSENSUS_TYPE_APP {
		federation_network.SubscribeToChat()
	}
	federation_network.ConnectFederationSeeds()

	globals.MainEvents.BroadcastEvent("main", "initialized")

	return
}

func InitMain(ready func()) (err error) {

	if err = arguments.InitArguments(os.Args[1:]); err != nil {
		return
	}
	if err = gui.InitGUI(); err != nil {
		return
	}
	if err = config.InitConfig(); err != nil {
		return
	}
	globals.MainEvents.BroadcastEvent("main", "config initialized")

	if err = network.NewNetwork(); err != nil {
		return
	}

	if err = app.Init(); err != nil {
		return
	}
	if err = startMain(); err != nil {
		return
	}

	if ready != nil {
		ready()
	}

	exitSignal := make(chan os.Signal, 10)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
	<-exitSignal

	if err = app.Close(); err != nil {
		return
	}
	signal.Stop(exitSignal)

	return
}
