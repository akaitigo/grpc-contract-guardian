# grpc-contract-guardian

gRPC/Proto定義の後方互換性チェック + 影響範囲可視化CLI。

## Features

- **Breaking Change検出**: `buf breaking` の結果をパースし、人間可読な形式で整形
- **影響範囲可視化**: サービス依存グラフから、breaking changeがどのサービスに波及するかを表示
- **PRコメント自動投稿**: GitHub PRにbreaking changeレポートを自動投稿

## Quick Start

```bash
# インストール
go install github.com/akaitigo/grpc-contract-guardian/cmd/guardian@latest

# Breaking changeチェック + 影響範囲表示
guardian check --against main

# GitHub PRにコメント投稿
guardian check --against main --format github --pr 123

# サービス依存グラフの表示
guardian graph --output text
guardian graph --output dot | dot -Tpng -o graph.png
```

## Development

```bash
# ビルド
make build

# テスト
make test

# lint
make lint

# 全チェック（format + tidy + lint + test + build）
make check
```

## Architecture

```
cmd/guardian/          # CLIエントリポイント（cobra）
internal/
  analyzer/            # proto解析、AST構築
  graph/               # サービス依存グラフ構築・出力
  buf/                 # buf breaking結果のパース・構造化
  reporter/            # テキスト/GitHubレポート生成
```

## License

MIT
