#!/usr/bin/env bats

setup() {
    TOOL="./bin/guardian"
    TEST_TEMP="$(mktemp -d)"
}

teardown() {
    rm -rf "$TEST_TEMP"
}

@test "version コマンドでバージョンが表示される" {
    run "$TOOL" version
    [ "$status" -eq 0 ]
    [[ "$output" == *"guardian"* ]]
}

@test "--help でusageが表示される" {
    run "$TOOL" --help
    [ "$status" -eq 0 ]
    [[ "$output" == *"check"* ]]
    [[ "$output" == *"graph"* ]]
    [[ "$output" == *"version"* ]]
}

@test "check --help でフラグ一覧が表示される" {
    run "$TOOL" check --help
    [ "$status" -eq 0 ]
    [[ "$output" == *"--against"* ]]
    [[ "$output" == *"--format"* ]]
    [[ "$output" == *"--pr"* ]]
    [[ "$output" == *"--proto-root"* ]]
}

@test "graph --help でフラグ一覧が表示される" {
    run "$TOOL" graph --help
    [ "$status" -eq 0 ]
    [[ "$output" == *"--output"* ]]
    [[ "$output" == *"--proto-root"* ]]
}

@test "graph --output text でtestdataのグラフが出力される" {
    run "$TOOL" graph --proto-root testdata --output text
    [ "$status" -eq 0 ]
    [[ "$output" == *"UserService"* ]]
}

@test "graph --output dot でDOT形式が出力される" {
    run "$TOOL" graph --proto-root testdata --output dot
    [ "$status" -eq 0 ]
    [[ "$output" == *"digraph"* ]]
    [[ "$output" == *"UserService"* ]]
}

@test "graph --output invalid でエラーが返る" {
    run "$TOOL" graph --proto-root testdata --output invalid
    [ "$status" -ne 0 ]
}

@test "存在しないサブコマンドでエラー" {
    run "$TOOL" nonexistent
    [ "$status" -ne 0 ]
}
