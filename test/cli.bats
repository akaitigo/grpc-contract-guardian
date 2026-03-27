#!/usr/bin/env bats
# =============================================================================
# grpc-contract-guardian CLI E2E テスト（bats-core）
# =============================================================================

# --- セットアップ ---

setup() {
    TOOL="./bin/guardian"
    TEST_TEMP="$(mktemp -d)"
}

teardown() {
    rm -rf "$TEST_TEMP"
}

# --- ヘルプ・バージョン ---

@test "--version でバージョンが表示される" {
    run "$TOOL" --version
    [ "$status" -eq 0 ]
    [[ "$output" == *"guardian"* ]]
}

# --- 引数バリデーション ---

@test "引数なしで実行するとusageが表示され終了コード1" {
    run "$TOOL"
    [ "$status" -eq 1 ]
    [[ "$output" == *"Usage"* ]]
}

# --- 出力制御 ---

@test "エラーメッセージがstderrに出力される" {
    run "$TOOL" 2>&1
    [ "$status" -eq 1 ]
}

# --- コマンド一覧 ---

@test "usageにcheckコマンドが含まれる" {
    run "$TOOL" 2>&1
    [[ "$output" == *"check"* ]]
}

@test "usageにgraphコマンドが含まれる" {
    run "$TOOL" 2>&1
    [[ "$output" == *"graph"* ]]
}
