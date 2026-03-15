# 测试目录说明

## 目录结构

```
tests/
├── unit/          # 单元测试（黑盒测试，仅测试公开API）✅
│   ├── config/                 # 配置模块单元测试
│   ├── model/                  # 数据模型单元测试
│   └── response/               # 响应工具单元测试
├── integration/   # 集成测试（测试模块间交互）✅
│   ├── database_test.go        # 数据库集成测试
│   └── middleware_test.go      # 中间件集成测试
├── benchmark/     # 性能基准测试（待添加）
├── testdata/      # 测试数据（待添加）
├── README.md      # 本文档
└── TEST_SUMMARY.md # 测试总结报告
```

## 测试类型

### 1. 单元测试 (unit/)

**目的**: 测试单个模块的公开API
**范围**: 不依赖外部资源（数据库、网络等）
**运行**: `go test ./tests/unit/...`

**示例**:
- 配置加载测试
- 数据模型测试
- 响应格式测试

### 2. 集成测试 (integration/)

**目的**: 测试多个模块协同工作
**范围**: 可能依赖数据库、外部服务
**运行**: `go test ./tests/integration/... -tags=integration`

**示例**:
- 数据库连接和操作
- 完整的HTTP请求处理流程
- 中间件和路由集成

### 3. 性能测试 (benchmark/)

**目的**: 测试关键操作的性能
**范围**: CPU、内存、并发性能
**运行**: `go test ./tests/benchmark/... -bench=. -benchmem`

**示例**:
- 数据库查询性能
- 并发请求处理
- 大文件上传性能

---

## 测试约定

### 文件命名

```
单元测试:   *_test.go
集成测试:   *_integration_test.go
性能测试:   *_benchmark_test.go
```

### 包命名

```go
// 单元测试 - 使用 _test 包（黑盒测试）
package config_test

// 集成测试 - 使用 integration 包
package integration

// 性能测试 - 使用 benchmark 包
package benchmark
```

---

## 运行测试

### 运行所有测试

```bash
# 所有测试（包括源代码目录中的测试）
go test ./...

# 只运行 tests 目录的测试
go test ./tests/...
```

### 按类型运行

```bash
# 单元测试
go test ./tests/unit/... -v

# 集成测试（需要数据库）
go test ./tests/integration/... -v -tags=integration

# 性能测试
go test ./tests/benchmark/... -bench=. -benchmem
```

### 覆盖率报告

```bash
# 生成覆盖率报告
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# 按模块查看覆盖率
go test ./tests/unit/... -cover
go test ./tests/integration/... -cover
```

### 并行测试

```bash
# 并行运行测试（默认）
go test ./tests/... -parallel 4

# 禁用并行
go test ./tests/... -parallel 1
```

---

## 测试最佳实践

### 1. 测试隔离

- 每个测试应该独立运行
- 不依赖其他测试的执行顺序
- 测试后清理资源

### 2. 使用表驱动测试

```go
func TestSomething(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"case1", "input1", "output1"},
        {"case2", "input2", "output2"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Something(tt.input)
            if result != tt.expected {
                t.Errorf("got %s, want %s", result, tt.expected)
            }
        })
    }
}
```

### 3. 使用测试辅助函数

```go
func setupTestDB(t *testing.T) *gorm.DB {
    t.Helper()
    // 设置测试数据库
    db := ...
    t.Cleanup(func() {
        // 清理
    })
    return db
}
```

### 4. Mock 外部依赖

```go
// 定义接口
type UserRepository interface {
    Create(user *User) error
}

// 测试时使用 Mock
type mockUserRepo struct {}
func (m *mockUserRepo) Create(user *User) error {
    return nil
}
```

---

## CI/CD 集成

### GitHub Actions

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v4
      with:
        go-version: '1.25'

    - name: Unit Tests
      run: go test ./tests/unit/... -v

    - name: Integration Tests
      run: go test ./tests/integration/... -v -tags=integration

    - name: Coverage
      run: |
        go test ./... -coverprofile=coverage.out
        go tool cover -func=coverage.out
```

---

## 测试数据

测试数据放在 `tests/testdata/` 目录：

```
tests/testdata/
├── config/
│   └── test_config.yaml
├── fixtures/
│   └── users.json
└── mock_data/
    └── api_responses.json
```

---

## 注意事项

1. **不要在测试中使用生产数据库**
   - 使用测试数据库或内存数据库
   - 集成测试后清理数据

2. **测试应该快速**
   - 单元测试应在毫秒级完成
   - 集成测试可以稍慢但不超过几秒

3. **测试应该稳定**
   - 避免依赖时间、随机数
   - 使用固定的测试数据

4. **保持测试代码质量**
   - 测试代码也需要重构
   - 添加必要的注释
   - 避免重复代码

---

## 相关文档

- [Go Testing 官方文档](https://golang.org/pkg/testing/)
- [Table Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
- [Go Test Comments](https://github.com/golang/go/wiki/TestComments)
