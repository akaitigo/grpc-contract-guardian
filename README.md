# grpc-contract-guardian

gRPC/Proto定義の後方互換性チェック + 影響範囲可視化CLI。

## Demo

```
$ guardian graph --proto-root testdata --output text

[SERVICE] UserService
  -> example.user.v1.GetUserRequest (input:GetUser)
  -> example.user.v1.GetUserResponse (output:GetUser)
  -> example.user.v1.ListUsersRequest (input:ListUsers)
  -> example.user.v1.ListUsersResponse (output:ListUsers)
```

```
$ guardian check --against main --dry-run

=== Breaking Change Impact Report ===
Total: 2 breaking change(s)

  HIGH:   1
  MEDIUM: 1

--- Details ---

1. [FIELD_REMOVED] user/v1/user.proto:10
   Previously present field "5" with name "email" on message "User" was deleted.
   Affected services: UserService
   Path: example.v1.GetUserResponse -[field:user]-> example.v1.User
```

## Features

- **Breaking Change検出**: `buf breaking` の結果をパースし、人間可読な形式で整形
- **影響範囲可視化**: サービス依存グラフから、breaking changeがどのサービスに波及するかを表示
- **PRコメント自動投稿**: GitHub PRにbreaking changeレポートを自動投稿

## Prerequisites

- **Go** 1.23+
- **buf** CLI ([installation guide](https://buf.build/docs/installation)) - required for breaking change detection
- **gh** CLI ([installation guide](https://cli.github.com/)) - required for `--format github` PR comment posting

> **Note**: guardian parses the text output of `buf breaking`. If buf changes its output format in future versions, parsing may need to be updated. This is tracked as a known risk.

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
