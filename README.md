# UPF Tester - UPF 包围测试工具

一个功能强大的 UPF (User Plane Function) 测试工具，支持通过流程配置进行信令上下线控制以及数据面测试。

## ✨ 核心功能

### 🔌 信令控制平面
- ✅ PFCP Association Setup/Release
- ✅ Session Establishment (会话建立)
- ✅ Session Modification (会话修改)
- ✅ Session Deletion (会话删除)
- ✅ 完整的会话生命周期管理
- ✅ 会话上下文跟踪 (SEID, TEID, UE IP)

### 📡 数据平面测试
- ✅ ICMP Echo 测试 (连通性验证)
- ✅ GTP-U 封装的上行数据发送
- ✅ 下行数据接收和验证
- ✅ 可配置的测试参数 (时长、包数量、间隔)
- ✅ 实时测试统计和结果报告

### 🎯 测试流程编排
- ✅ 基于 YAML 的灵活配置
- ✅ 支持多种测试步骤类型
- ✅ 可配置的延迟和等待
- ✅ 会话与数据流的自动关联
- ✅ 完整的错误处理和日志

## 🚀 快速开始

### 前置要求
- Go 1.18+
- UPF 实例 (如 free5GC UPF)
- 网络连通性 (N4 和 N3 接口)

### 编译
```bash
cd /localdisk/upf-tester/cmd
go build -o upf-tester main.go
```

### 配置
编辑 `config/config.yaml`:
```yaml
basic:
  localN4Ip: "192.168.12.200"  # SMF N4 接口 IP
  upfN4Ip: "192.168.12.210"    # UPF N4 接口 IP
dataPlane:
  gnbIp: "192.168.12.203"      # 模拟 gNB IP
  n3Ip: "192.168.12.213"       # UPF N3 接口 IP
  n6Ip: "192.168.12.216"       # UPF N6 接口 IP
  dnIp: "192.168.12.206"       # DN (数据网络) IP
resources:
  queueSize: 10000
  startUeIp: "10.250.0.1"
  startSeId: 1
  startTeTd: 1
```

### 运行
```bash
cd /localdisk/upf-tester/cmd
./upf-tester
```

## 📋 测试用例配置

### 完整测试流程示例
`testcases/complete_test_case/complete_test_case.yaml`:
```yaml
testSteps:
  # 1. 建立会话
  - step: 1
    type: "session_establishment_request"
    action: "send"
    path: "01_session_establishment_request.yaml"

  - step: 2
    type: "session_establishment_response"
    action: "recv"
  
  # 2. 数据平面测试
  - step: 3
    type: "data_plane_test"
    action: "icmp"
    path: "05_data_plane_test.yaml"

  # 3. 修改会话
  - step: 4
    type: "session_modification_request"
    action: "send"
    path: "03_session_modification_request.yaml"

  - step: 5
    type: "session_modification_response"
    action: "recv"

  # 4. 删除会话
  - step: 6
    type: "session_deletion_request"
    action: "send"
    path: "06_session_deletion_request.yaml"

  - step: 7
    type: "session_deletion_response"
    action: "recv"
```

### 数据平面测试配置
`testcases/complete_test_case/yaml/05_data_plane_test.yaml`:
```yaml
testType: "icmp"
duration: 10        # 测试时长（秒）
packetCount: 20     # 发送包数量
interval: 500       # 发送间隔（毫秒）
payloadSize: 64     # 负载大小（字节）
```

## 🏗️ 架构设计

### 核心组件

#### 1. 信令控制层 (`internal/handler`)
- `pfcphandler.go` - PFCP 消息分发器
- `assochandler.go` - Association 处理
- `testcasehandler.go` - 测试用例执行器
- `session_context.go` - 会话上下文管理

#### 2. 编码层 (`encoding/pfcp`)
- `establishmentrequest.go` - Session Establishment 编码
- `modificationrequest.go` - Session Modification 编码
- `deletionrequest.go` - Session Deletion 编码
- `types.go` - PFCP 数据结构

#### 3. 数据平面层 (`internal/dataplane`)
- `test.go` - 数据平面测试框架
- `sender.go` - 数据包发送器
- `receiver.go` - 数据包接收器
- `gtp.go` - GTP-U 封装
- `icmp.go` - ICMP 消息构造

#### 4. 工具层 (`internal/util`)
- `seid.go` - SEID 分配器
- `seqnumber.go` - 序列号管理
- `teid.go` - TEID 资源管理

### 会话与数据流关联

每个会话建立后自动分配：
- **SEID** - 会话标识符
- **TEID** - 数据面隧道标识符
- **UE IP** - 用户设备 IP 地址

数据平面测试自动使用当前会话的这些标识符，实现会话与数据流的一一对应。

## 📊 测试步骤类型

| 类型 | Action | 说明 |
|------|--------|------|
| `session_establishment_request` | send | 发送会话建立请求 |
| `session_establishment_response` | recv | 接收会话建立响应 |
| `session_modification_request` | send | 发送会话修改请求 |
| `session_modification_response` | recv | 接收会话修改响应 |
| `session_deletion_request` | send | 发送会话删除请求 |
| `session_deletion_response` | recv | 接收会话删除响应 |
| `data_plane_test` | icmp | ICMP 连通性测试 |
| `sleep` | wait | 等待指定秒数 |

## 🎯 使用场景

### 场景 1: 基础功能验证
验证 UPF 的基本 PFCP 信令和数据转发功能。

### 场景 2: 会话生命周期测试
测试完整的会话建立、修改、删除流程。

### 场景 3: 数据平面验证
在会话建立后进行数据平面连通性测试。

### 场景 4: 并发会话测试
通过配置多个会话，测试 UPF 的并发处理能力。

## 📝 日志输出

程序运行时会输出详细的日志：
```
2026/01/16 04:19:21 Sending session establishment request, SEID: 0x0000000000000001
2026/01/16 04:19:21 Session established successfully, SMF SEID: 0x0000000000000001, UPF SEID: 0x0000000000000002
2026/01/16 04:19:23 Starting ICMP test: UE IP=10.250.0.1, TEID=1, Duration=10s
2026/01/16 04:19:33 ICMP test completed (timeout)
2026/01/16 04:19:33 ICMP Test Result: Sent=20, Received=0, Lost=20, Loss Rate=100.00%
2026/01/16 04:19:35 Sending session deletion request, UPF SEID: 0x0000000000000002
2026/01/16 04:19:35 Session deleted successfully, SEID: 0x0000000000000001
```

## 🔧 扩展开发

### 添加新的测试类型
1. 在 `testcasehandler.go` 的 switch 语句中添加新的 case
2. 实现相应的测试逻辑
3. 更新测试用例 YAML 配置

### 添加新的数据平面测试
1. 在 `internal/dataplane/test.go` 中实现新的测试类型
2. 实现 `DataPlaneTest` 接口
3. 在 `testcasehandler.go` 中集成

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License

## 🙏 致谢

- [go-pfcp](https://github.com/wmnsk/go-pfcp) - PFCP 协议库
- [free5GC](https://github.com/free5gc/free5gc) - 5G 核心网参考实现