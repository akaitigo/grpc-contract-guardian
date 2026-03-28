# ADR-002: MVPでは正規表現パーサーを使用し、v2でprotoccompile移行

## ステータス
承認

## コンテキスト
PRD では buf の内部ライブラリ `protocompile` を利用した proto 解析を想定していたが、
MVP 段階では以下の理由から正規表現ベースのパーサーを採用した。

- protocompile は import 解決やソースファイルシステムの構成が必要で、MVP のスコープを超える
- 現時点の対象は単一パッケージ内の `.proto` ファイルであり、正規表現で十分対応可能
- 外部依存を最小化し、シングルバイナリの軽量性を維持したい

## 決定
MVP（v1.x）では `internal/analyzer` パッケージで正規表現ベースのパーサーを使用する。
v2 で protocompile への移行を検討する。

## 選択肢
| 選択肢 | 長所 | 短所 |
|--------|------|------|
| A: 正規表現パーサー（採用） | 依存ゼロ、実装が速い、デバッグ容易 | ネストmessage・oneof・map等の高度な構文対応が限定的 |
| B: protocompile | 完全なAST取得、全構文対応 | import解決の設定が複雑、依存が大きい |
| C: protoc --descriptor_set_out | 公式ツール利用 | protoc のインストール必須、バイナリ解析が必要 |

## 移行計画
1. v1.x: 正規表現パーサーで基本的な service/message/field/rpc を解析
2. v2.0: protocompile に移行し、以下を対応
   - ネスト message の完全対応
   - oneof / map フィールド
   - proto2 構文
   - クロスパッケージ import 解決

## リスク
- 正規表現パーサーは edge case（複雑なコメント、ネスト構造）で誤認する可能性がある
- v2 移行時にパーサーのインターフェースを維持し、呼び出し側の変更を最小化する必要がある
