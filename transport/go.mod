module github.com/ease-lab/xdt/transport

go 1.16

replace (
	github.com/ease-lab/xdt/utils => ../utils
)

require (
	github.com/ease-lab/xdt/utils v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.8.1
)
