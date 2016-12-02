![gopher](http://golang-jp.org/doc/gopher/talks.png)

# echoフレームワークにREST APIを組み込む

## 目的

GO言語のechoフレームワークを遣い、シンプルなREST APIを実装する。
リクエストを受けると、AmazonRDSに接続して結果をJSONで返す。

## 実行環境

```
Amazon Linux AMI
t2.micro
1 コア vCPU (最大 3.3 GHz)、1 GiB メモリ RAM、8 GB ストレージ

RDS MySQL
```

## Go1.6の設定


```bash
$ vi $HOME/.bash_profile
--------------------------------------------------
     :     追記
export GOVERSION=1.6
export GOROOT=/usr/local/go
export GOPATH=$HOME/go/$GOVERSION
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin
--------------------------------------------------
```

```bash
$ source $HOME/.bash_profile
```

```bash
$ go version
go version go1.6 linux/amd64
```

```bash
$ mkdir -p $GOPATH/src
```

## Glideのインストール

```bash
$ curl https://glide.sh/get | sh
$ glide -v
glide version v0.12.3
```

## サンプルプロジェクト作成

Githubからソースコードをクローンします

```bash
$ cd $GOPATH/src
$ git clone https://github.com/hayao56/echoapi.git
$ cd echoapi
```

## 依存パッケージのインストール

```bash
$ glide install
```

## echoサーバー起動

```bash
$ go run server.go
  http server started on :1234
```

## サーバーアクセス

```bash
$ curl http://localhost:1234
```

