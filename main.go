package main

import (
	"go-webapi-fw/common/mqutils"
	"go-webapi-fw/web/iris_srv"
)

func main() {

	mqutils.BindConsumer()
	iris_srv.Start()
}
