# watermint switchbox

[![Build](https://github.com/watermint/toolbox/actions/workflows/build.yml/badge.svg)](https://github.com/watermint/toolbox/actions/workflows/build.yml)
[![Test](https://github.com/watermint/toolbox/actions/workflows/test.yml/badge.svg)](https://github.com/watermint/toolbox/actions/workflows/test.yml)
[![CodeQL](https://github.com/watermint/toolbox/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/watermint/toolbox/actions/workflows/codeql-analysis.yml)
[![Codecov](https://codecov.io/gh/watermint/toolbox/branch/main/graph/badge.svg?token=CrE8reSVvE)](https://codecov.io/gh/watermint/toolbox)

![watermint toolbox](resources/images/watermint-toolbox-256x256.png)

Dropbox、Dropbox for teams、Google、GitHubなどのウェブサービス用の多目的ユーティリティコマンドラインツール。

# ライセンスと免責条項

watermint toolboxはMITライセンスのもと配布されています.
詳細はファイル LICENSE.mdまたは LICENSE.txt ご参照ください.

以下にご留意ください:
> ソフトウェアは「現状のまま」で、明示であるか暗黙であるかを問わず、何らの保証もなく提供されます。ここでいう保証とは、商品性、特定の目的への適合性、および権利非侵害についての保証も含みますが、それに限定されるものではありません。

# ビルド済み実行ファイル

コンパイル済みバイナリは [最新のリリース](https://github.com/watermint/toolbox/releases/latest) からダウンロードいただけます. ソースコードからビルドする場合には [BUILD.md](BUILD.md) を参照してください.

## macOS/LinuxでHomebrewを使ってインストールする。

まずHomebrewをインストールします. 手順は [オフィシャルサイト](https://brew.sh/)を参照してください. 次のコマンドを実行してwatermint toolboxをインストールします.
```
brew tap watermint/toolbox
brew install toolbox
```

# 製品ライフサイクル

## メンテナンス ポリシー

この製品自体は実験的なものであり、サービスの品質を維持するためのメンテナンスの対象ではありません。プロジェクトは、重大なバグやセキュリティ上の問題を最善の努力で修正するよう努めます。しかし、それは保証されているわけではありません。

この製品は、特定のメジャーリリースのパッチリリースをリリースしません。本製品は、修正が認められた場合、次のリリースとして修正を適用します。

## 仕様変更

このプロジェクトの成果物は、スタンドアロンの実行可能プログラムです。プログラムのバージョンを明示的にアップグレードしない限り、仕様変更は適用されません。

新バージョンのリリースにおける変更は、以下の方針で行われます。

コマンドパス、引数、戻り値などは、可能な限り互換性を保つようにアップグレードされますが、廃止または変更される可能性があります。
一般的な方針は以下の通り。

* 引数の追加やメッセージの変更など、既存の動作を壊さない変更は予告なく実施されます。
* 使用頻度が低いと思われるコマンドは、予告なく廃止または移動されます。
* その他のコマンドの変更は、30～180日以上前に発表されます。

仕様の変更は[お知らせ](https://github.com/watermint/toolbox/discussions/categories/announcements)で発表されます。仕様変更予定一覧は[仕様変更](https://toolbox.watermint.org/ja/guides/spec-change.html)をご参照ください。

# セキュリティとプライバシー

## 情報は収集しません 

watermint toolboxは、第三者のサーバーに情報を収集することはありません. 

watermint toolboxは、Dropbox のようなサービスとご自身のアカウントでやりとりするためのものです. 第三者のアカウントは関与していません. コマンドは、PCのストレージにAPIトークン、ログ、ファイル、またはレポートを保存します.

## 機密データ

APIトークンなどの機密データのほとんどは、難読化されてアクセス制限された状態でPCのストレージに保存されています. しかし、それらのデータを秘密にするのはあなたの責任です.
特に、ツールボックスのワークパスの下にある`secrets`フォルダ(デフォルトでは`C:\Users\<ユーザー名>\.toolbox`、または`$HOME/.toolbox`フォルダ以下)は共有しないでください。

# 利用方法

`tbx` にはたくさんの機能があります. オプションなしで実行をするとサポートされているコマンドやオプションの一覧が表示されます.
つぎのように引数なしで実行すると利用可能なコマンド・オプションがご確認いただけます.

```
% ./sbx

watermint switchbox xx.x.xxx
============================

© 2024-2024 Takayuki Okazaki
オープンソースライセンスのもと配布されています. 詳細は`license`コマンドでご覧ください.

Dropbox用ツールとDropbox for teams

使い方:
=======

./sbx  コマンド

利用可能なコマンド:
===================

| コマンド | 説明                       | 備考 |
|----------|----------------------------|------|
| license  | ライセンス情報を表示します |      |
| version  | バージョン情報             |      |

```

# コマンド

## ユーティリティー

| コマンド                               | 説明                       |
|----------------------------------------|----------------------------|
| [license](docs/ja/commands/license.md) | ライセンス情報を表示します |
| [version](docs/ja/commands/version.md) | バージョン情報             |

