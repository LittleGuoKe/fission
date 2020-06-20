#!/usr/bin/env sh
set -eux
set -o pipefail
# -u 表示遇到不存在的变量的时候，报错
# -e 只要脚本有错误，就终止执行
# -x 表示执行命令之前，先输出执行的命令
# -o pipefail 只要一个子命令失败，整个管道命令就失败，脚本就会终止执行


# 配置go mod
export GO111MODULE=on
export GOPROXY=https://goproxy.io
export KUBECONFIG=$(pwd)/kubectl-config

# 下载依赖
#go mod download

# 运行代码
go run -gcflags=-trimpath=$GOPATH -asmflags=-trimpath=$GOPATH \
  cmd/fission-bundle/main.go --mqt --routerUrl http://func.fission.huxiang.pro