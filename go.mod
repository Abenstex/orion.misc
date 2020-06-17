module orion.misc

go 1.13

require (
	github.com/eclipse/paho.mqtt.golang v1.2.0
	github.com/sirupsen/logrus v1.6.0
	//github.com/eclipse/paho.mqtt.golang v1.2.0
	github.com/spf13/viper v1.7.0
	github.com/ztrue/shutdown v0.1.1
	laniakea v0.0.0
	orion.commons v0.0.0-local
)

replace (
	laniakea => ../laniakea
	orion.commons => ../orion.commons
)
