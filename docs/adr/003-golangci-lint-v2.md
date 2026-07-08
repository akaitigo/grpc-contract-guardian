# ADR-003: golangci-lint v2 への移行

## ステータス
承認

## コンテキスト
golangci-lint v1 系は v1.64.x をもって開発終了し、以降の Go バージョン対応・リンター更新は v2 系でのみ提供される。
v2 は設定ファイル形式が非互換（`version: "2"` 必須、`linters-settings` → `linters.settings`、
formatters の分離、`run.timeout` 廃止等）であり、CI 側も `golangci-lint-action@v6` は v2 バイナリに非対応のため
action@v7 + バージョン明示が必要になる。

2026-04 に v2 移行が着手されたが（別クローンでの手作業編集）、未完のまま放置されていた。
手作業版には `linters-settings` トップレベル残存・`colored-line-number` 形式名・廃止済み `stylecheck` 設定という
v2 非互換の誤りが含まれていた。

## 決定
- golangci-lint **v2.12.2** に更新し、設定は公式 `golangci-lint migrate` による機械変換を採用する
- CI は `golangci/golangci-lint-action@v7` + `version: v2.12.2` の明示指定とする
- 実効リンター集合は v1 時代と同等を維持する
  - `typecheck` は v2 でリンターから除外（コンパイルエラーとして常時報告）
  - `gosimple` / `stylecheck` は `staticcheck` に統合（`staticcheck.checks: all` で維持）
  - `gofumpt` / `goimports` は `formatters` セクションに分離
  - `errcheck` / `govet` / `ineffassign` / `unused` は v2 デフォルト有効のため enable 列挙から省略（設定は維持）

## 選択肢
| 選択肢 | 長所 | 短所 |
|--------|------|------|
| A: 公式 migrate で機械変換（採用） | 変換の正確性が保証される、v1 の実効挙動を維持 | v1 の暗黙デフォルト（exclusions presets）が明示化され設定が長くなる |
| B: 手作業版（別クローンの編集）を採用 | 4月の作業を活用 | v2 非互換の誤りが3箇所あり、config verify を通らない |
| C: v1.64.8 のまま維持 | 変更ゼロ | 開発終了系列のため新 Go バージョンで壊れる時限爆弾になる |

## 検証
- `golangci-lint config verify` パス
- `golangci-lint run ./...` で 0 issues（v2.12.2）
- `make check`（lint → test → build）パス

## リスク
- v2 の exclusions presets（`comments` / `common-false-positives` / `legacy` / `std-error-handling`）は
  v1 の暗黙デフォルト除外を明示化したもの。将来 preset を外す場合は検出件数が増える
- `run.timeout` は v2 でデフォルト無効。CI でハングした場合はジョブ側のタイムアウトで検出する
