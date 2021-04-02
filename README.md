arciv
====

Sorry, the README is written in japanese only now...

___このREADMEは書きかけです。___

___arciv は現在ベータ版です。破壊的変更が加わり、互換性を維持しないことで今後のバージョンのバイナリを利用するとリポジトリが読めなくなる可能性があります。___

arcivは大容量のバイナリファイル向けバージョン管理システムです。
複数の写真や動画を始めとするファイルを長期間に渡って保存する場合に、バージョン管理をしながら別のディスクやAWS S3に保存することが出来ます。
特に AWS S3 Glacier に保存することを前提としており、ファイルの実態が変化しなければディレクトリ構成やファイル名の変更を行っても多重に課金されることを避けることが出来ます。

## Description

arcivは指定したディレクトリ (以下リポジトリルートと呼ぶ) とその配下のファイルをまとめて１つのリポジトリとして扱います。
リポジトリ内のファイルの変更履歴を管理し、別のディスクや AWS S3 上にバックアップすることが出来ます。

## Demo

```sh
# バージョン管理・バックアップしたいディレクトリがあるとします。
$ cd ~/important-data
$ echo "IMPORTANT STRING" > message.txt

# まずリポジトリを初期化します。
$ arciv init

# バックアップ用の他リポジトリ (今回はディレクトリ) を指定します。
$ arciv repository add type:file name:local-disk path:/media/disk0/backup-important-data

# ファイル構成を記録してバックアップします。
$ arciv store --repository local-disk

# バックアップが行われたことを確認します。
$ arciv log --repository local-disk

# リポジトリに変更を加えます。
$ rm message.txt

# リポジトリに変更が加わったことを確認します。
$ arciv status

# リポジトリに加えた変更を、バックアップから復元することで修正します。
# (先程の store 時の commit-id が 60dddd-0000000000000000000000000000000000000000000000000000000000000000 であったとする。)
$ arciv restore --repository local-disk --commit 60ddd --force
$ ls
```

## VS. 

Version Control System として普及する git と比較して次のような特徴があります。

### Good points

#### 構造が単純

arciv が commit 時に生成するリポジトリの状態を記録した list は、概ね sha256sum で作成されるリストと同じ構造を持ちます。
万一リポジトリに問題が起きても、単純な構造を活かし、手動またはシェルスクリプトを用いて復元できる可能性が高いです。

#### .git以下に複製をもたない

git は分散型のバージョン管理システムであることから、ローカルでバージョン管理を完結する必要があり、リポジトリで管理されるファイルの複製が .git ディレクトリ以下に存在します。
数百GBや数TB単位の容量を持つバイナリファイル群を管理する場合、git であると少なくともローカルで、元のバイナリファイル群の2倍以上のディスク容量を確保する必要があります。

arciv は、gitのように頻繁に書き戻すことを前提とせず、長期的にファイルを保管する際のバックアップや誤って削除したデータの復元のために用いるシステムであるため、ローカルでバージョン管理を完結することを目指しません。
arciv は、自身または他のリポジトリを明示的に指定して `store`サブコマンドを実行したときにはじめて、変更があった記録すべきファイルの実体を転送し保存します。
ローカルに常に複製がないのは用途によっては短所ですが、arcivがターゲットとする環境では長所となります。

#### アップロードしたファイルが上書き/変更されない

arciv は、git と異なり、commit で保存されるリポジトリの状態を記録した list と、バックアップすべきファイルの実体である blob を分けて保管します。
これにより特に AWS S3 において、ファイルの目的に合わせて異なるストレージクラスを利用できます。
arciv のバックアップを用いると頻繁に参照される可能性のある listは すぐにアクセスできる S3 標準クラスへ、参照される機会の少ない blob はストレージ単価の安い Glacier Deep Archive 層に保存することになっています。
大容量のファイルを長期間に渡り安価に保管するのに適したデータ構造を持つのがarcivです。

また、Glacier Deep Archive 層を用いるにあたり、blob はファイルの中身 (sha256 hash) のみによって指定され、過去に作成された blob は必ず上書きされないことになっています。
これにより、リポジトリ内でファイルを移動したとしても Glacier Deep Archive 層のファイルは変更されず、最低利用期間に関する重複した課金が発生せずに済みます。

### Bad points

#### 圧縮しない

git とは異なり、arciv はテキストファイルとバイナリファイルを区別せずに、全て非圧縮で保存します。
これは arciv がバイナリファイルをターゲットとしているためで、現在動画や画像などで用いられているバイナリのファイル形式は十分に圧縮されており、再度圧縮してもあまり効果がないことが理由です。
よって、テキストファイルを中心とするファイル群を管理する場合は、圧縮して記録する git のほうが向いているでしょう。

#### ローカルに複製を保存しない

good に書いた「.git 以下に複製を持たない」ことの裏返しですが、arciv はバージョン管理システム兼バックアップツールであり、他のリポジトリ (disk, AWS S3など) にも保存することを前提としています。
ローカルだけでバージョン管理を行いたい場合は、分散型のバージョン管理システムである git のほうが向いているでしょう。

## Requirement

## Usage

### 認証

AWS S3 上にバックアップを作成したい場合は、S3 バケットの作成、IAM ロールの作成、認証情報の登録が必要になります。

### 初期化 (init)

```sh
$ cd /path/to/repository/dir
$ arciv init
# カレントディレクトリをリポジトリとして扱うことを宣言し、管理情報を初期化します。
# 具体的には /path/to/repository/dir/.arcivディレクトリを作成し、配下に必要なディレクトリとファイルを生成します。
```

### 他リポジトリの登録/閲覧/削除 (repository / repository add / repository remove)
```sh
$ arciv repository add type:file name:backup-repo path:/path/to/backup/repository
# 又は
$ arciv repository add type:s3 name:backup-repo region:us-east-1 bucket:your-bucket-name
# バックアップする場所を登録します。
# arciveのリポジトリには2種類のtypeがあります。
# 1つは同じコンピュータの別のディレクトリをリポジトリとしてpathで指し示すtype:fileです。
# もう1つは、AWS S3上のバケットをリポジトリとしてregionとbucketで指し示すtype:s3です。
# リポジトリを追加する際には次に示すメタ情報を`メタ情報名:メタ情報文字列`のようにコロンでつなぎ、これらを半角スペースをあけて並べることで指定しなければなりません。
# type:fileには type, name, path を指定する必要があります
# type:s3には type, name, region, bucket を指定する必要があります。

$ arciv repository
# 登録したリポジトリを確認します。
# name:self はデフォルトで登録されている自分自身のリポジトリであり、削除はできません。

# リポジトリの登録を削除する場合は次のようにして行います
# 以下のコマンドを実行すると自身のリポジトリからyour-repository-nameリポジトリへの参照は削除しますが、リポジトリの実態(ディレクトリやAWS S3バケットの中のファイルや.arcivディレクトリの中身)は削除されません。
$ arciv repository remove your-repository-name
```

### バックアップの実行 (store)

```sh
# リポジトリを登録した他リポジトリ (ここではyour-repository-nameリポジトリ) にバックアップします。
$ arciv store --repository your-repository-name

# バックアップが完了すると、その時点でのリポジトリの中身を記録したcommitが作成されます。
# commitは自身のリポジトリとバックアップ先のリポジトリに作成されますが、commitが指し示すファイルの実体はバックアップ先のリポジトリのみに保存されます。
(`arciv store`コマンドを実行した際に自身のリポジトリのファイルが変更,削除されるわけではありません)
# 補足: commitが記録するリポジトリの中身とは、各ファイルごとのリポジトリルートからの相対パスとsha256です。
# 補足: commitが指し示すファイルの実体とは、sha256とそれに対応する元ファイルのバイト列です。
```

補足,注意: ___AWS S3 にアクセスすると課金が発生します。___ 特に AWS S3 Glacier Deep Archive を利用するため、すぐにファイルを消しても最低利用期間分の課金が発生することに注意してください。

### バックアップ結果の閲覧 (log)

```sh
# commitを見てみましょう
# 自身のリポジトリのcommit一覧は次のように確認します。
$ arciv log
# 他のリポジトリのcommit一覧も次のように確認できます。
$ arciv log --repository your-repository-name

# 各commitの中身は次のように確認します。
# ここで指定するcommit-idは先程の`arciv log`で確認された各行の0-9,a-f, - で構成された73文字の文字列のことです。
$ arciv log --commit <commit-id>
```

#### 補足,コラム: commit-idの省略

ここで指定するcommit-idは先程の`arciv log`で確認された各行の0-9,a-f, - で構成された73文字の文字列のことです。
73文字全体を指定してもいいですし、先頭かハイフンより後ろから始まる数文字のみを指定しても構いません。
ただし、省略した指定がリポジトリ内の他のcommit-idをも指し示すことがありえるときは使えません。

たとえば、`01234567-abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789` というcommit-idであれば、次のような指定が考えられます。

- `01234567-abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789`
- `01234567-abc`
- `01234567`
- `abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789`
- `012345`
- `012`
- `abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789`
- `abcdef0123456789abcdef0123456789abcdef0123`
- `abcd`
- `a`

省略した指定がリポジトリ内の他のcommit-idをも指し示すときとは、次のような場合に発生します。
例えば`arciv log --repository hogerepo`の結果が
```
01234567-abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789
01288867-ab00000000000000000000000000000000000000000000000000000000000000
```
のとき、次のような指定は有効ですが、
- `01234567-abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789`
- `0123`
- `01234567-ab`
- `abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789`
- `abc`

次のような指定はエラーとなります。

- `012`
- `01`
- `0`
- `ab`
- `a`

2つのcommit-idのうちどちらを指しているかわからないからです。


#### リポジトリになにか手を加えた後、バックアップから復元してみましょう。

```sh
# リポジトリ内のファイルに何らかの変更を加えてみましょう
$ vi change-any-files-1
$ rm change-any-files-2

# 再度バックアップしてみましょう
$ arciv store --repository your-repository-name

# 変更を加える前に書き戻してみましょう

## type:file の場合
$ arciv restore --repository your-repository-name --commit commit-id

## type:s3 の場合
# AWS S3 Glacier Deep Archive を利用しているため、すぐにファイルの実体をダウンロードできません。
# 一度AWS S3内でのアーカイブからの復元(ファイルをダウンロードできる状態にしてもらう) をリクエストし、48時間未満での完了を待ってから実体のダウンロードとリポジトリの復元を実行します。このように2段階での復元操作が必要となります。

# まず AWS S3 に アーカイブからの復元をリクエスト
$ arciv restore --request --repository your-repository-name --commit commit-id
# --valid-days --force --fast --dry-run
# コマンドを実行するとrestore-request-idが表示されるので、メモしておきます。

# 48時間程度待ってAWS S3内でのアーカイブからの復元が完了したら、実体のダウンロードとリポジトリの復元を実行
$ arciv restore --run-requested <restore-request-id>
# restore-request-id を忘れてしまったら `$ls .arciv/restore-request` で確認できます。restore-request-idは新しいほど辞書順で並べたときにあとになるようにIDが生成されています。
```

### status

前回のcommitから変更・削除・追加があったファイル名を表示します。

### diff

2つのcommitの間で変更・削除・追加があったファイル名を表示します。

### version

### stash

__非推奨__

低レベルなアクセスです。
リポジトリ内のファイル構成を記録(commitを作成)し、ファイルの実体を.arciv/blobに退避します。
### unstash

__非推奨__

リポジトリ内のファイル構成をcommitから取得し、.arciv/blobを利用してファイルの実体を復元します。

### s3lowaccess

__非推奨__

ファイル、region、bucket を指定して AWS S3と通信します。


## Install

```sh
$ git clone https://github.com/basd4g/arciv.git
$ arciv
$ go build
$ mv arciv /usr/local/bin/
```

## Architecture

### .arciv directory

arcivではバージョン管理に必要な情報や他のリポジトリの情報をリポジトリ直下の`.arciv`ディレクトリに格納しています。
本章では、この`.arciv`ディレクトリ配下に保存されるファイルについて説明します。

- `.arciv/blob/` 他リポジトリからダウンロードしたり、一時的に退避したりしたファイルの実体を保存するディレクトリです。バックアップ先のリポジトリでは原則としてファイルの実体はこのディレクトリの中のみに保存され、リポジトリの中の`.arciv`ディレクトリ以外は空となります。
- `.arciv/list/` 各commit-idをファイル名として、そのcommitに含まれるファイルのリポジトリルートからの相対パスとファイルのsha256を記録したものです。場合によっては過去のcommitとの差分のみを記録していることがあります。
- `.arciv/restore-request/` AWS S3 Glaclier からアーカイブ済みファイルをダウンロードできる状態にするようリクエストしたときに、そのリクエストIDをファイル名としたリクエスト情報を記録するファイルを含むディレクトリです。
各ファイルに含まれるのは`#`で始まるメタ情報の他に、各行が復元をリクエストしたファイルの実体のsha256が記録されています。
- `.arciv/repositories` `arciv repository add`で登録したリポジトリを記録するファイルです。selfは含みません。
- `.arciv/timeline`commit-idのリストを保持するファイルです。
- `.arciv/timestamps`commit作成時に使える--fastオプションを実行するための、各ファイルのタイムスタンプ情報をキャッシュするファイルです。

### aws s3 bucket

type:file でも type:s3 でも基本的には変わらず、ほかリポジトリには .arciv ディレクトリ以下のみに読み書きを行なうようになっています。
aws s3 bucket の場合は、ファイルの実体を保管している .arciv/blob/ 以下のファイルのみ、AWS S3 Glacier Deep Archive 階層に保管されます。


## Contribution

## TODO

- シンボリックリンクの扱いの調査と明確化 dir と file での違い
- restore-request-id を確認する方法
- commit-id と restore-request-id の実施時刻をhuman readable に印字する機能の追加

## 用語集

- Commit
- list
- blob
- 他リポジトリ
- ファイルの実体

## Licence

[MIT](https://github.com/tcnksm/tool/blob/master/LICENCE)

## Author

[basd4g](https://github.com/basd4g)
