# PRD: grpc-contract-guardian

## 概要

gRPC/Proto定義の後方互換性チェック、変更影響範囲の可視化、自動レビューコメント生成を行うCI統合CLIツール。

## 背景・課題

- マイクロサービス間のgRPC契約変更は、影響範囲が見えにくく破壊的変更がレビューをすり抜ける
- bufは lint/breaking に特化しているが、「どのサービスが影響を受けるか」の可視化機能がない
- PRレビュー時にbreaking changeの影響範囲を手動で調査するのはコストが高い

## ターゲットユーザー

- gRPC/Protobufを使うマイクロサービス開発チーム
- CI/CDパイプラインにProto互換性チェックを組み込みたいPlatform Engineer

## ゴール

1. proto定義の変更によるbreaking changeを検出し、影響範囲をサービス依存グラフで可視化する
2. GitHub PRにレビューコメントとして自動投稿し、レビュー負荷を削減する
3. bufのbreaking check結果を人間可読な形式に整形する

## 非ゴール（v1スコープ外）

- protoc プラグインとしての動作
- gRPC通信の実行時モニタリング
- 自動修正・マイグレーションコード生成
- GitLab/Bitbucket対応（GitHub PRのみ）

## MVP スコープ（2週間）

### Milestone 1: Proto解析エンジン（3日）

- .protoファイルのパースとAST構築
- サービス → メッセージ → フィールドの依存関係グラフ構築
- DOT形式およびテキスト形式での依存グラフ出力

### Milestone 2: Breaking Change検出（4日）

- `buf breaking --against` の出力をパース・構造化
- 変更のカテゴリ分類（フィールド削除、型変更、サービスメソッド変更等）
- 依存グラフと組み合わせた影響範囲の算出

### Milestone 3: レポーター + CLI（3日）

- テキスト形式レポート（ターミナル出力）
- GitHub PR コメント投稿（`gh api` 経由）
- Cobra CLIインターフェース
  - `guardian check --against main --format text|github`
  - `guardian graph --output dot|text`

## 受け入れ条件

- [ ] `guardian check --against main` でbreaking changeとその影響範囲がテキスト出力される
- [ ] `guardian check --against main --format github --pr 123` でPRコメントが投稿される
- [ ] `guardian graph` でサービス依存グラフがテキストまたはDOT形式で出力される
- [ ] テストケースに対してbreaking change検出率が100%（buf breakingの出力をパースするため）
- [ ] CIでの実行時間が30秒以内（中規模プロジェクト: proto 50ファイル）

## 技術選定

| 項目 | 選定 | 理由 |
|------|------|------|
| 言語 | Go | シングルバイナリ配布、CLIエコシステム（cobra）、buf自体がGoで書かれている |
| Proto解析 | protocompile (buf内部ライブラリ) | bufと同じ解析エンジンで互換性が高い |
| CLI | cobra | Go CLIのデファクト。サブコマンド・フラグ管理が容易 |
| GitHub連携 | gh CLI / GitHub API | PRコメント投稿。OAuthトークンはgh auth経由 |

## 競合比較

| 機能 | buf | Spectral | guardian |
|------|-----|----------|---------|
| lint | Yes | Yes | No（bufに委譲） |
| breaking check | Yes | No | Yes（bufの結果を活用） |
| 影響範囲可視化 | No | No | **Yes** |
| PRコメント自動投稿 | No | No | **Yes** |
| 依存グラフ出力 | No | No | **Yes** |

## 成功指標

- GitHub Starが公開1ヶ月で10以上
- open-posプロジェクトのCI に組み込まれ、実運用で使われている
- breaking changeの見逃しがゼロ（bufベースなので理論上達成可能）
