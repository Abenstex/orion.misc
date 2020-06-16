module orion.misc

go 1.13

require (
	github.com/eclipse/paho.mqtt.golang v1.2.0
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/gorilla/mux v1.7.4 // indirect
	github.com/lib/pq v1.7.0 // indirect
	github.com/rhinoman/couchdb-go v0.0.0-20180321180027-310a5a9beb66 // indirect
	github.com/sirupsen/logrus v1.6.0 // indirect
	github.com/spf13/viper v1.7.0
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	laniakea v0.0.0
	orion.commons v0.0.0
)

replace (
	laniakea => ../laniakea
	orion.commons => ../orion.commons
)
