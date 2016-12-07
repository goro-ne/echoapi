![gopher](http://golang-jp.org/doc/gopher/talks.png)

# echoフレームワークにREST APIを組み込む

## 目的

GO言語のechoフレームワークを使ってシンプルなREST APIを実装する。
GET/POST/PUT/DELETEリクエストを受け取るとAmazonRDSに接続し、マスタテーブルを参照/追加/更新/削除する。

参考にしたサイト
http://qiita.com/keika299/items/62e806ae42828bb3567a

## 実行環境

```
Amazon Linux AMI
t2.micro - $0 / 月750時間
1 コア vCPU, 1 GiB メモリ RAM, 8 GB ストレージ

RDS for MySQL 5.6.34 (Auroraに移行した場合は MySQL 5.6.10aと互換性あり)
db.t2.micro - $0 / 月750時間
1 コア vCPU, 1 GiB メモリ, 10 GB ストレージ, シングルAZ
DBユーザー: 任意
データベースのポート: 3306
```


## RDS for MySQLの設定

![RDS for MySQL](https://cdn2.iconfinder.com/data/icons/amazon-aws-stencils/100/Database_copy_Amazon_RDS_MySQL_DB_Instance-128.png)
```
1. Webのセキュリティグループの作成
  - VPCのセキュリティグループをクリック
  - セキュリティグループの作成、ネームタグ:test-web、VPC:vpc-xxxxxxxx (defautと同じ)
  - インバウンドタブで、タイプ:カスタムTCPルール、プロトコル:TCP、ポート範囲:1234、送信元:0.0.0.0/0を指定する。
  - アウトバウンドタブで、タイプ:すべてのトラフィック、プロトコル:すべて、ポート範囲:すべて、送信元:0.0.0.0/0を指定する。
```
```
2. RDS for MySQLのセキュリティグループの作成 (RDSに接続するインスタンス)
  - VPCのセキュリティグループをクリック
  - セキュリティグループの作成、ネームタグ:test-db、VPC:vpc-xxxxxxxx (defautと同じ)
  - インバウンドタブで、タイプ:MYSQL/Aurora、プロトコル:TCP、ポート範囲:3306、送信先:RDSのセキュリティグループID: sg-xxxxxxxを指定する。
  - アウトバウンドタブで、タイプ:すべてのトラフィック、プロトコル:すべて、ポート範囲:すべて、送信元:0.0.0.0/0を指定する。
```
```
3. RDSのパラメータグループを作成 (文字エンコードをlatin1からutf8にして日本語対応)
  - RDSダッシュボードのパラメータグループを選択する。
  - パラメータグループ作成をクリック。
  - パラメータグループファミリー:「mysql5.6」、グループ名:「utf8」、説明:「UTF8」
  - 「utf8」を選択し、パラメータの編集をクリック。
  - character_set_client、character_set_connection、character_set_database、character_set_results、character_set_serverを「utf8」に設定する。
  - skip-character-set-client-handshakeを「1」に設定する。
  - time_zoneを「Asia/Tokyo」に変更して、変更の保存する。
```
```
4. RDSのインスタンスを作成
  - RDS for MySQL 5.6.34
  - パラメータグループ: utf8
  - host: MySQLドメイン、port: 3306、database: 'test'、user: 'dev'、password: 'password' (config/db.yamlに任意の値を設定)
```

## MySQLクライアントのインストール

```bash
$ sudo -i
# rpm -ihv http://downloads.mysql.com/archives/get/file/MySQL-client-5.6.33-1.el7.x86_64.rpm
# yum search mysql 56
mysql56.x86_64 : MySQL client programs and shared libraries
     :
# yum install mysql56.x86_64
```

## MySQL RDSへの接続、テストテーブル作成

```bash
$ mysql -h エンドポイント -P ポート -u ユーザ名 -p DB名
Enter password: 設定したユーザのパスワード

mysql> show variables like 'char%';
+--------------------------+-------------------------------------------+
| Variable_name            | Value                                     |
+--------------------------+-------------------------------------------+
| character_set_client     | utf8                                      |
| character_set_connection | utf8                                      |
| character_set_database   | utf8                                      |
| character_set_filesystem | binary                                    |
| character_set_results    | utf8                                      |
| character_set_server     | utf8                                      |
| character_set_system     | utf8                                      |
| character_sets_dir       | /rdsdbbin/mysql-5.6.34.R1/share/charsets/ |
+--------------------------+-------------------------------------------+
8 rows in set (0.00 sec)


mysql>
create table test.userinfo(
  id int not null auto_increment,
  email varchar(50),
  first_name varchar(50),
  last_name varchar(50),
  primary key(id)
);

mysql> desc test.userinfo;
+------------+-------------+------+-----+---------+----------------+
| Field      | Type        | Null | Key | Default | Extra          |
+------------+-------------+------+-----+---------+----------------+
| id         | int(11)     | NO   | PRI | NULL    | auto_increment |
| email      | varchar(50) | YES  |     | NULL    |                |
| first_name | varchar(50) | YES  |     | NULL    |                |
| last_name  | varchar(50) | YES  |     | NULL    |                |
+------------+-------------+------+-----+---------+----------------+
4 rows in set (0.00 sec)

mysql> insert into test.userinfo(email,first_name,last_name) values ('musashi-miyamoto@gmail.com','武蔵','宮本');

mysql> select * from test.userinfo;
+----+----------------------------+------------+-----------+
| id | email                      | first_name | last_name |
+----+----------------------------+------------+-----------+
|  1 | musashi-miyamoto@gmail.com | 武蔵       | 宮本      |
+----+----------------------------+------------+-----------+
1 rows in set (0.00 sec)

```


## Go1.6の設定

```
$ sudo -i
# mkdir -p /usr/local/go
# cd /usr/local
# wget https://storage.googleapis.com/golang/go1.6.linux-amd64.tar.gz
# tar -C /usr/local -xzf go1.6.linux-amd64.tar.gz
# chmod -R 777 /usr/local/go
# exit
```

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

$ go get golang.org/x/tools/cmd/goimports
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
(2回目以降は glide update)
```

## echoサーバー起動

```bash
$ go run server.go
  http server started on :1234
```

## サーバーアクセス

公開IPアドレス: EC2のセキュリティグループのインバウンドタブで、ポート:1234を追加

### GETリクエスト

```bash
$ curl http://公開IPアドレス:1234/user/1

{"id":1,"email":"musashi-miyamoto@gmail.com","firstName":"武蔵","lastName":"宮本"}
```


## Crome拡張「Advanced REST client」からデータ登録

![Advanced REST client](https://github.com/jarrodek/ChromeRestClient/blob/develop/app/assets/arc_icon_128.png?raw=true)

https://chrome.google.com/webstore/detail/advanced-rest-client/hgmloofddffdnphfgcellkdfbfbjeloo/related

ここからChrome拡張機能を追加


### POSTリクエスト

```
http://公開IPアドレス:1234/users/

Raw Header
-----------------------------------------------------
Content-Type: application/json
-----------------------------------------------------

Raw payload
-----------------------------------------------------
{
  "email":"kojiro-sasaki@gmail.com",
  "firstName":"小次郎",
  "lastName":"佐々木"
}
-----------------------------------------------------
SENDをクリック
```

```bash
mysql> select * from test.userinfo;
+----+----------------------------+------------+-----------+
| id | email                      | first_name | last_name |
+----+----------------------------+------------+-----------+
|  1 | musashi-miyamoto@gmail.com | 武蔵       | 宮本      |
|  2 | kojiro-sasaki@gmail.com    | 小次郎     | 佐々木    |
+----+----------------------------+------------+-----------+
2 rows in set (0.00 sec)
```


