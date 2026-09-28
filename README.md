# Round-Robin League Manager

一个完全本地的小型联赛赛程与积分管理工具。使用 Go 标准库实现，无任何在线服务依赖。
程序从标准输入逐行读取 JSON 请求，向标准输出逐行返回 JSON 结果；赛程与赛果以
`league.json` 保存在 `--state` 指定的数据目录中，正常退出后重开查询结果一致。

## 运行

```sh
go build ./cmd/league
./league --state ./data < requests.jsonl
```

或直接：

```sh
go run ./cmd/league --state ./data < examples/session1.jsonl
```

每行一个请求 JSON，每行一个响应 JSON。`--state` 为必填参数；目录不存在会自动创建。

## 请求格式

| action | 其他字段 | 说明 |
| --- | --- | --- |
| `init` | `teams`: 至少 2 个非空且唯一的队伍编号 | 创建联赛与赛程。已有联赛时拒绝，不覆盖 |
| `schedule` | 无 | 查询赛程，按轮次、比赛编号排序 |
| `register` | `match_id`, `home_goals`, `away_goals` | 登记赛果，仅接受未登记比赛 |
| `correct` | `match_id`, `home_goals`, `away_goals` | 更正赛果，仅接受已登记比赛 |
| `revoke` | `match_id` | 撤销赛果，恢复未进行状态 |
| `standings` | `through_round`（可选，正整数） | 查询积分榜；省略时统计全部轮次 |

`home_goals`/`away_goals` 必须是非负整数。未知比赛、非法比分、不符合当前状态的
操作都会得到 `{"ok":false,"error":"..."}`，且不改变任何已有数据。

响应字段：

- `init`/`schedule`：`teams`（按字典序排序）、`total_rounds`；`schedule` 另有
  `matches`，每场含 `id`、`round`、`home`、`away`、`played` 及已登记时的
  `home_goals`/`away_goals`。
- `standings`：`standings` 数组，字段为 `team`、`played`、`wins`、`draws`、
  `losses`、`goals_for`、`goals_against`、`goal_diff`、`points`。

## 赛程规则

- 单循环：每对队伍恰好交锋一次；每队每轮最多一场。
- 偶数队：`队数 - 1` 轮；奇数队：`队数` 轮，每轮一队轮空，轮空不生成比赛。
- 初始化时队伍按编号字典序排序后使用循环移位法排程，因此相同名单无论输入顺序如何，
  生成的赛程与比赛编号完全一致（不保证主客场均衡）。
- 比赛编号形如 `M01-1`：前两位为轮次，`-` 后为该轮内按对阵双方字典序排列的序号。
- 初始化后队伍名单和赛程固定，不允许重复初始化覆盖。

## 积分与排序规则

- 只统计已登记比赛：胜 3 分、平局各 1 分、负 0 分。
- 轮空与未进行比赛不计场次、不得分。
- 积分榜包含全部队伍（含全零行），排序依次为：积分降序、净胜球降序、进球数降序、
  队伍编号字典序升序。
- `through_round` 只纳入该轮及之前**已登记**的赛果。允许先登记较后轮次的比赛，
  但不会据此推定此前轮次已完成。截止轮次超出赛程总轮次时直接拒绝，不会静默截断。

## 示例

`examples/session1.jsonl` 演示了初始化、查询赛程、登记、跨轮积分榜、更正、撤销的
完整流程：

```sh
go run ./cmd/league --state ./data < examples/session1.jsonl
```

## 测试

```sh
go test ./...
```

测试覆盖：奇/偶数队赛程性质（对阵不重复、轮次数、轮空）、输入顺序无关性、
平局并列排序、登记/更正/撤销状态约束、跨轮查询、超范围截止轮次拒绝，以及
保存/重开后数据一致。
