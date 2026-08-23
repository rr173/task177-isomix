# task177-isomix · 同位素混合来源约束求解服务

纯后端 Go 服务：地球化学研究员登记候选端元组成区间、样品同位素测量与
排除/比例约束，服务用线性规划求解质量守恒下的**来源比例可行域**；不可行时
返回**最小不可满足约束集**（冲突核心）。接受的求解可发布为可引用报告，
端元/约束修订只触发新求解，已发布报告冻结快照、只可替代不可改写。

## 技术栈

- Go 1.26.3（`GOTOOLCHAIN=local`，CGO_ENABLED=0）
- SQLite（纯 Go 驱动 `modernc.org/sqlite`，WAL 模式，离线可构建）
- 组件版本锁见 `component-versions.json`
- 核心算法：自研两阶段单纯形法（`internal/lp`），Bland 规则防退化、
  容差 1e-9、NaN/Inf 一律判数值不稳定并附证据

## 快速开始

```bash
# 端到端自检（真实建库、求解、发布报告、重启恢复验证）
GOTOOLCHAIN=local go run ./cmd/isomix --smoke-test

# 启动 HTTP 服务（默认内存库）
GOTOOLCHAIN=local go run ./cmd/isomix --addr :8080 --db ./isomix.db

# 构建 / 静态检查 / 测试
CGO_ENABLED=0 GOTOOLCHAIN=local go build ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go vet   ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go test  ./...
```

## 业务闭环

1. 登记端元（组成区间）→ 校验 → 置为可用；
2. 登记样品测量（区间 + 协方差摘要）→ 检查不确定度；
3. 登记约束（排除 / 比例 / 上限）并启用；
4. 提交求解：构建质量守恒 + 同位素区间 + 用户约束的线性系统，
   求各端元比例可行区间；不可行时删除式求最小冲突核心；
5. 接受结果 → 发布报告（冻结端元/约束版本快照）；
6. 端元/约束修订 → 旧报告标记已替代，新提交产生新求解。

## 状态机

- 端元：`draft → validated → available`，`available → excluded`；修订回 `draft`
- 样品：`pending → solvable | overspread`，`solvable → solved`；任意态 → `sealed`
- 约束：`enabled ⇄ relaxed | conflict`，任意态 → `revoked`
- 求解：`queued → running → feasible | infeasible | numerically_unstable`
- 报告：`draft → published → superseded`

## 持久化与恢复

- SQLite 保存端元区间、样品测量、协方差摘要、约束、求解矩阵摘要、
  可行域与报告快照；`--smoke-test` 关闭重开同一数据库验证恢复。
- 输入哈希（端元/约束版本 + 样品测量的规范序列化 SHA-256）幂等：
  相同输入只返回已有结果；重启后未完成求解（queued/running）经
  持久化输入哈希校验后重算恢复。

## API（前缀 /api）

| 能力 | 入口 | 生产实现 |
|---|---|---|
| 端元登记 | POST /api/endmembers | httpapi/endmember_api.go → endmember.Service → store.EndmemberStore |
| 端元列表/查询 | GET /api/endmembers[/{id}] | 同上 |
| 端元修订/校验/可用/排除 | PUT /api/endmembers/{id}，POST .../validate \| available \| exclude | 同上 |
| 样品登记/列表/查询 | POST|GET /api/samples[/{id}] | httpapi/sample_api.go → measure.Service → store.SampleStore |
| 样品修订/检查/封存 | PUT ...，POST .../check \| seal | 同上 |
| 约束登记/列表/查询 | POST|GET /api/constraints[/{id}] | httpapi/constraint_api.go → constraint.Service |
| 约束启用/松弛/撤销 | POST .../enable \| relax \| revoke | 同上 |
| 求解提交/列表/查询 | POST|GET /api/solutions[/{id}] | httpapi/solution_api.go → solve.Service（LP） |
| 可行域查询 | GET /api/solutions/{id}/feasible-region | solve → lp.Bounds |
| 冲突核心查询 | GET /api/solutions/{id}/conflict-core | solve → minimalConflictCore（MUS） |
| 报告创建/列表/查询 | POST|GET /api/reports[/{id}] | httpapi/report_api.go → report.Service |
| 报告发布/过期检查/差异 | POST .../publish，GET .../supersede-check \| diff | 同上 |
| 健康/自检/统计 | GET /api/health[/selfcheck]，GET /api/stats | httpapi/health_api.go |

## 目录结构

```
cmd/isomix/main.go        # 入口：--addr --db --smoke-test
internal/
  model/                  # 实体、状态机、领域错误
  hashutil/               # 输入哈希（幂等与恢复锚点）
  endmember/              # 端元管理 + 组成区间校验
  measure/                # 样品测量 + 协方差正定校验
  constraint/             # 约束管理 + 编译为线性不等式
  lp/                     # 两阶段单纯形求解器（可行域/活跃约束）
  solve/                  # 求解编排 + 冲突核心（MUS）+ 重启恢复
  report/                 # 报告冻结 / 发布 / 替代 / 差异
  store/                  # SQLite 迁移与各实体存取
  httpapi/                # REST JSON API
```
