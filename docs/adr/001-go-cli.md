# ADR-001: Go + CLI アーキテクチャの採用

## ステータス
承認

## コンテキスト
gRPC/Proto定義の後方互換性チェックと影響範囲可視化ツールの技術選定が必要。
CI/CDパイプラインに組み込むため、軽量で高速な実行環境が求められる。

## 決定
Go言語でCLIツールとして実装する。CLIフレームワークにcobraを採用。

## 選択肢
| 選択肢 | 長所 | 短所 |
|--------|------|------|
| A: Go CLI | シングルバイナリ配布、bufとの親和性（同じGo）、高速起動、cobra/viper エコシステム | GUIなし |
| B: TypeScript CLI (Node.js) | Web UIとの共通化が容易、npm配布 | 起動遅い、ランタイム依存、CI環境でNode.js必須 |
| C: Rust CLI | 最速実行、メモリ安全 | protobufエコシステムがGoほど充実していない、学習コスト |

## 結果
- シングルバイナリをCI環境にダウンロードするだけで使える（ランタイム不要）
- bufの内部ライブラリ（protocompile）を直接利用でき、proto解析の互換性が高い
- cobra による `guardian check` / `guardian graph` のサブコマンド設計が自然に実現できる
