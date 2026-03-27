# Harvest: grpc-contract-guardian

> 日付: 2026-03-26
> パイプライン: Launch → Scaffold → Build → Harden → Ship → **Harvest**

## テンプレートの実戦評価

### 使えたもの
- [x] Makefile (build/test/lint/format/check) — 毎Issue で使用
- [x] lint設定 (.golangci.yml) — CIで有効に動作
- [x] CI YAML (.github/workflows/ci.yml) — 全PRで自動実行
- [x] CLAUDE.md (44行ポインタ設計) — 効果的だった
- [x] ADR テンプレート — 001-go-cli.md で技術選定を記録
- [x] 品質チェックリスト 5項目 — Ship時に全項目確認
- [x] E2E テスト (bats) — CLIの動作確認に有用

### 使えなかったもの
| ファイル | 理由 |
|---------|------|
| PostToolUse Hooks | golangci-lint未インストール。post-lint.shにインストールチェックは入っていたが、hookの前にstartup.shを実行する運用が定着していない |
| lefthook | lefthookコマンド自体が未インストール。同上 |
| startup.sh | 一度も実行しなかった。「実行する動機」が弱い |
| session-end.sh | 同上 |
| CONTEXT.json | 手動更新せず。git logで代替 |
| progress.json | 使わなかった |
| quality-override.md | quality-checklist.mdだけで十分だった |
| hooks-structure.md | 参照されなかった |

### テンプレートへの改善提案
| 提案 | 対象ファイル | 内容 |
|------|------------|------|
| startup.sh 自動実行 | idea-work スキル | Step 3の冒頭で `bash .claude/startup.sh` を必ず実行する手順を追加 |
| 不要ファイル削減 | layer-2/cli-tool | quality-override.md を最小限（--help確認+exit code）に |
| ADR割り切り記録 | idea-work スキル | MVP簡易実装を選ぶ際のADR記載を必須化 |

### 次のPJへの申し送り
- startup.sh を初回に必ず実行すること（ツール自動インストールのため）
- Hooks が動いているか、最初のファイル編集後に確認すること
- CONTEXT.json は使わないなら生成しない選択肢もあり（git logで十分な場合）
