# grpc-contract-guardian

## コマンド
- ビルド: `make build`
- テスト: `make test`
- lint: `make lint`
- フォーマット: `make format`
- 全チェック: `make check`
- E2E: `bats test/cli.bats`

## ワークフロー
1. research.md を作成（調査結果の記録）
2. plan.md を作成（実装計画。人間承認まで実装禁止）
3. 承認後に実装開始。plan.md のtodoを進捗管理に使用

## 構造
- cmd/guardian/main.go — エントリポイント（cobra root command）
- internal/analyzer/ — proto解析、AST構築
- internal/graph/ — サービス依存グラフ構築・出力
- internal/buf/ — buf breaking結果のパース・構造化
- internal/reporter/ — テキスト/GitHubレポート生成

## ルール
- ADR: docs/adr/ 参照。新規決定はADRを書いてから実装
- テスト: 機能追加時は必ずテストを同時に書く
- lint設定の変更禁止（ADR必須）
## 禁止事項
- any型 / 不要なinterface{} → 具体型またはジェネリクス
- fmt.Println のコミット（log パッケージを使用）
- TODO コメントのコミット（Issue化すること）
- .env・credentials のコミット
- lint設定の無効化

## Hooks
- 設定: .claude/settings.json
- 構造定義: docs/hooks-structure.md

## 状態管理
- progress.json: セッション履歴（JSON形式）
- 開始: `bash .claude/startup.sh`
- 終了: `bash .claude/session-end.sh "要約" "次アクション"`
