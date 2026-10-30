# tolo-relation-stub

```bash
go run ./cmd/server --data ./testdata/sample.json
```

データは`--data`フラグもしくは`RELATION_DATA`環境変数でJSONファイルパスを渡す　　

フラグが優先される  

`--data` も `RELATION_DATA` も指定がない場合は起動時にエラー

```bash
RELATION_DATA=./testdata/sample.json RELATION_ADDR=:9000 ./server
```

## Docker

```bash
docker run --rm -p 8080:8080 \
  -v "$(pwd)/testdata:/data:ro" \
  -e RELATION_DATA=/data/sample.json \
  ghcr.io/pj-hoakari/tolo-relation-stub:latest
```

対応プラットフォーム: `linux/amd64`, `linux/arm64`

イメージ内のデフォルト環境変数

| 変数 | デフォルト |
| --- | --- |
| `RELATION_ADDR` | `:8080` |
| `GIN_MODE` | `release` |
| `RELATION_DATA` | （未設定。実行時に必ず指定する） |

### Docker Compose

```bash
docker compose -f compose.sample.yaml up
```

## 関係性モデルの制約

- ユーザは必ず 1 つのテナントに属し、複数のテナントには属せない
- ユーザがあるイベントに対してロールを持つ場合、そのイベントの所属テナントに対してもロールを持っている必要がある（テナントロールが上位の制約）
- テナントロールとイベントロールの種類（`owner` / `staff`）は独立（例: テナント staff かつイベント owner が成立）
- イベントは必ず 1 つのテナントに属する（テナント : イベント = 1 : N、0 件も可）

## API

### `GET /tenants/:tenantId/users/:userId/memberships`

tolo-tenant-management の所属参照 API（`GET /tenants/{tenant_id}/users/{user_id}/memberships`）と同じ形で応答する

```json
{
  "memberships": [
    {
      "user_id": "<userId>",
      "tenant_id": "<tenantId>",
      "tenant_role": "owner",
      "event_roles": [
        { "event_id": "<eventId>", "role": "staff" }
      ]
    }
  ]
}
```

- `memberships` の要素は 0 件または 1 件。ユーザがそのテナントにロールを持たない場合は `[]`
- `event_roles` は当該テナント配下のイベントのうち、ユーザがロールを持つもののみ列挙（無ければ `[]`）

| Case | Status |
| --- | --- |
| `:tenantId` が存在しない | `404`（`{"error": "tenant not found"}`） |
| テナントは存在し、ユーザにテナントロールが無い | `200`（`memberships: []`） |
| テナントロールあり | `200`（`memberships` に 1 件） |

### `GET /healthz`

`200 {"status":"ok"}` を返すヘルスチェック

## シード JSON フォーマット

```json
{
  "tenants":       [{ "id": "...", "name": "..." }],
  "events":        [{ "id": "...", "tenantId": "...", "name": "..." }],
  "tenantMembers": [{ "tenantId": "...", "userId": "...", "role": "owner|staff" }],
  "eventMembers":  [{ "eventId":  "...", "userId": "...", "role": "owner|staff" }]
}
```

ロード時に以下を検証し、違反があれば起動時にエラー終了

- `events[].tenantId` が `tenants[].id` に存在すること
- `tenantMembers[].tenantId` / `eventMembers[].eventId` がそれぞれ実在すること
- `role` が `owner` / `staff` のいずれかであること
- ID の重複が無いこと
- 同一ユーザが `tenantMembers` に複数現れないこと（1 ユーザ = 1 テナント）
- `eventMembers[].userId` が、対応するイベントの所属テナントの `tenantMembers` に存在すること

ID値は単純な文字列として扱うため，制限等はなし

## サンプルシードと動作例

`testdata/sample.json` に同梱

### Tenant

| Label | ID | Name |
| --- | --- | --- |
| T1 | `019e6cc4-…-753e-…-d398` | Alpha Inc |
| T2 | `019e6cc4-…-76b1-…-f2c9` | Beta Org |

### Event

| Label | ID | Tenant | Name |
| --- | --- | --- | --- |
| E1 | `…-76be-…-6ef7` | T1 | Summer Fest |
| E2 | `…-76c4-…-0b70` | T1 | Winter Meetup |
| E3 | `…-76ca-…-1e06` | T2 | Beta Launch |

### ユーザと関係性

| Label | ID | 所属テナント | テナントロール | Eventロール |
| --- | --- | --- | --- | --- |
| U1 | `…-76cf-…-035e` | T1 | owner | E1: staff |
| U2 | `…-76d4-…-b9df` | T1 | staff | E1: owner, E2: staff |
| U3 | `…-76da-…-917c` | T2 | owner | E3: staff |
| U4 | `019e6cdd-…-2868` | T2 | staff | E3: owner |

### テナント・ユーザごとのレスポンス組み合わせ

| Tenant | User | Status | `memberships[0].tenant_role` | `memberships[0].event_roles` |
| --- | --- | --- | --- | --- |
| T1 | U1 | 200 | `"owner"` | `[{E1, "staff"}]` |
| T1 | U2 | 200 | `"staff"` | `[{E1, "owner"}, {E2, "staff"}]` |
| T1 | U3 | 200 | -（`memberships: []`） | - |
| T1 | U4 | 200 | -（`memberships: []`） | - |
| T1 | 未知ユーザ | 200 | -（`memberships: []`） | - |
| T2 | U1 | 200 | -（`memberships: []`） | - |
| T2 | U2 | 200 | -（`memberships: []`） | - |
| T2 | U3 | 200 | `"owner"` | `[{E3, "staff"}]` |
| T2 | U4 | 200 | `"staff"` | `[{E3, "owner"}]` |
| 未知テナント | 任意 | 404 | - | - |
