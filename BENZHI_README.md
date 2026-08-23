# BENZHI 评测说明 · task177-isomix

同位素混合来源约束求解服务（纯后端 Go，SQLite 持久化）。

## 构建与自检

```bash
# 标准命令（须真实成功）
CGO_ENABLED=0 GOTOOLCHAIN=local go build ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go vet   ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go test  ./...
GOTOOLCHAIN=local go run ./cmd/isomix --smoke-test

# Docker 双架构构建（linux/amd64、linux/arm64）
bash build_benzhi_docker.sh task177-isomix linux/amd64
bash build_benzhi_docker.sh task177-isomix linux/arm64
```

## --smoke-test 契约

`go run ./cmd/isomix --smoke-test` 不启动长驻服务，而是：

1. 真实创建 3 个端元（mantle/crust/organic，同位素 d18O、d2H）并置为可用；
2. 创建样品 river-water，检查为可求解；
3. 提交求解 → 可行，输出 3 端元比例可行区间（质量守恒 Σ=1）；
4. 同一样品重复提交 → 输入哈希幂等，返回同一求解；
5. 加排除约束（organic）→ 可行域收缩，organic 比例坍缩为 [0,0]；
6. 加矛盾比例约束（mantle:crust=4）→ 不可行，返回最小冲突核心（含该约束）；
7. 发布可行解报告 → 修订端元 → 旧报告判定已替代并标记 superseded；
8. 注入一条 queued 求解（模拟崩溃）→ 关闭数据库 → 重开同一数据库：
   数据齐全、幂等命中持久化解、queued 记录经输入哈希恢复重算。
9. 全部断言通过后以退出码 0 结束，输出 `SMOKE TEST PASSED`。

## API

全部路由前缀 `/api`，JSON 请求/响应，错误体 `{code, message}`。
主要入口见 README「API」表；端到端示例：

```bash
# 登记端元
curl -s -X POST localhost:8080/api/endmembers \
  -d '{"name":"mantle","components":{"d18O":{"lo":5.2,"hi":5.8},"d2H":{"lo":-90,"hi":-70}}}'
# 校验并置可用、登记样品、提交求解
curl -s -X POST localhost:8080/api/endmembers/{id}/validate
curl -s -X POST localhost:8080/api/endmembers/{id}/available
curl -s -X POST localhost:8080/api/samples \
  -d '{"name":"river-water","measurements":{"d18O":{"lo":6.0,"hi":8.0},"d2H":{"lo":-85,"hi":-50}}}'
curl -s -X POST localhost:8080/api/samples/{id}/check
curl -s -X POST localhost:8080/api/solutions -d '{"sample_id":"{id}"}'
```

## Docker 双架构

`Dockerfile`（= `benzhi.Dockerfile`）单阶段构建：
`golang:1.26.3-bookworm`，`CGO_ENABLED=0 go build ./cmd/isomix`，
`ENTRYPOINT ["/app/isomix"]`、`CMD ["--smoke-test"]`。
`docker run --rm <镜像> --smoke-test` 为容器自检的唯一判据；
健康基线已通过 `docker_baseline_validation.py --verify-and-record` 双架构验证
（证明文件 `.private/docker_baseline_validation.json`）。
