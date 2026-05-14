# AlfQ · 设备信息收集方案

> 文档版本：v1.0
> 创建日期：2026-05-14
> 适用范围：AlfQ 安卓客户端 + AntClaw 服务端

---

## 一、概述

设备信息收集是移动应用的基础功能，用于：
- 用户行为分析
- 设备兼容性检测
- 安全风控
- 个性化推荐

**核心原则**：
- ✅ 隐私合规（GDPR/隐私政策）
- ✅ 数据最小化（仅收集必要信息）
- ✅ 加密传输（HTTPS）
- ✅ 用户可控制（支持数据删除）

---

## 二、职责划分

### 2.1 客户端职责（主要）

| 职责 | 说明 |
|-----|------|
| **信息采集** | 通过系统 API 获取设备硬件、软件、网络等信息 |
| **数据封装** | 将采集到的信息组织成统一格式 |
| **加密传输** | 通过 HTTPS 发送到服务端 |
| **本地缓存** | 缓存设备信息，避免重复采集 |

### 2.2 服务端职责（次要）

| 职责 | 说明 |
|-----|------|
| **数据接收** | 接收客户端上报的设备信息 |
| **数据校验** | 验证数据合法性和完整性 |
| **数据存储** | 将设备信息持久化到数据库 |
| **数据分析** | 统计分析设备分布、用户行为等 |

---

## 三、数据模型

### 3.1 设备信息字段定义

| 字段名 | 类型 | 来源 | 说明 | 敏感程度 |
|-------|------|-----|------|---------|
| `deviceId` | String | 客户端生成 | 设备唯一标识符（UUID） | 中 |
| `model` | String | Build.MODEL | 设备型号（如 iPhone 15 Pro） | 低 |
| `brand` | String | Build.BRAND | 设备品牌（如 Apple/Samsung） | 低 |
| `osVersion` | String | Build.VERSION.RELEASE | 操作系统版本（如 Android 14） | 低 |
| `osType` | String | 固定值 | 操作系统类型（android/ios） | 低 |
| `appVersion` | String | BuildConfig.VERSION_NAME | 应用版本号（如 1.0.0） | 低 |
| `buildNumber` | String | BuildConfig.VERSION_CODE | 构建号（如 100） | 低 |
| `screenWidth` | Int | DisplayMetrics | 屏幕宽度（像素） | 低 |
| `screenHeight` | Int | DisplayMetrics | 屏幕高度（像素） | 低 |
| `networkType` | String | ConnectivityManager | 网络类型（wifi/cellular/none） | 低 |
| `timezone` | String | TimeZone.getDefault() | 时区（如 Asia/Shanghai） | 低 |
| `locale` | String | Locale.getDefault() | 语言地区（如 zh_CN） | 低 |
| `manufacturer` | String | Build.MANUFACTURER | 设备制造商 | 低 |
| `fingerprint` | String | Build.FINGERPRINT | 设备指纹（用于识别） | 中 |
| `createdAt` | Timestamp | 服务端生成 | 首次上报时间 | 低 |
| `updatedAt` | Timestamp | 服务端生成 | 最后更新时间 | 低 |

### 3.2 数据结构

```protobuf
message DeviceInfo {
  string device_id = 1;           // 设备唯一标识
  string model = 2;               // 设备型号
  string brand = 3;               // 设备品牌
  string os_version = 4;          // 操作系统版本
  string os_type = 5;             // 操作系统类型
  string app_version = 6;         // 应用版本号
  string build_number = 7;        // 构建号
  int32 screen_width = 8;         // 屏幕宽度
  int32 screen_height = 9;        // 屏幕高度
  string network_type = 10;       // 网络类型
  string timezone = 11;           // 时区
  string locale = 12;             // 语言地区
  string manufacturer = 13;       // 制造商
  string fingerprint = 14;        // 设备指纹
}
```

---

## 四、客户端实现方案

### 4.1 收集时机

| 时机 | 说明 |
|-----|------|
| **App 首次启动** | 安装后首次打开时采集并上报 |
| **App 版本更新** | 版本变更时重新采集（可能有新字段） |
| **设备信息变更** | 网络切换、系统升级等事件触发 |
| **定时同步** | 每天/每周同步一次最新信息 |

### 4.2 Android 实现示例

```kotlin
class DeviceInfoCollector(private val context: Context) {
    
    fun collect(): DeviceInfo {
        val metrics = context.resources.displayMetrics
        
        return DeviceInfo(
            deviceId = getDeviceId(),
            model = Build.MODEL,
            brand = Build.BRAND,
            osVersion = Build.VERSION.RELEASE,
            osType = "android",
            appVersion = BuildConfig.VERSION_NAME,
            buildNumber = BuildConfig.VERSION_CODE.toString(),
            screenWidth = metrics.widthPixels,
            screenHeight = metrics.heightPixels,
            networkType = getNetworkType(),
            timezone = TimeZone.getDefault().id,
            locale = Locale.getDefault().toString(),
            manufacturer = Build.MANUFACTURER,
            fingerprint = Build.FINGERPRINT
        )
    }
    
    private fun getDeviceId(): String {
        // 使用 Android ID 或生成 UUID 并持久化
        val prefs = context.getSharedPreferences("device", Context.MODE_PRIVATE)
        var deviceId = prefs.getString("device_id", null)
        if (deviceId.isNullOrEmpty()) {
            deviceId = UUID.randomUUID().toString()
            prefs.edit().putString("device_id", deviceId).apply()
        }
        return deviceId
    }
    
    private fun getNetworkType(): String {
        val connectivityManager = context.getSystemService(Context.CONNECTIVITY_SERVICE) as ConnectivityManager
        val network = connectivityManager.activeNetwork
        val capabilities = connectivityManager.getNetworkCapabilities(network)
        
        return when {
            capabilities?.hasTransport(NetworkCapabilities.TRANSPORT_WIFI) == true -> "wifi"
            capabilities?.hasTransport(NetworkCapabilities.TRANSPORT_CELLULAR) == true -> "cellular"
            else -> "none"
        }
    }
}
```

### 4.3 上报逻辑

```kotlin
class DeviceInfoRepository(private val api: AntclawApi) {
    
    suspend fun reportDeviceInfo(info: DeviceInfo) {
        try {
            val request = ReportDeviceInfoRequest.newBuilder()
                .setDeviceId(info.deviceId)
                .setModel(info.model)
                .setBrand(info.brand)
                .setOsVersion(info.osVersion)
                .setOsType(info.osType)
                .setAppVersion(info.appVersion)
                .setBuildNumber(info.buildNumber)
                .setScreenWidth(info.screenWidth)
                .setScreenHeight(info.screenHeight)
                .setNetworkType(info.networkType)
                .setTimezone(info.timezone)
                .setLocale(info.locale)
                .setManufacturer(info.manufacturer)
                .setFingerprint(info.fingerprint)
                .build()
            
            api.reportDeviceInfo(request)
        } catch (e: Exception) {
            // 上报失败，保存到本地待重试
            saveToCache(info)
        }
    }
    
    private fun saveToCache(info: DeviceInfo) {
        // 保存到 SharedPreferences 或 Room
        // 后续在网络恢复时重试
    }
}
```

---

## 五、服务端实现方案

### 5.1 API 接口设计

```protobuf
service DeviceService {
  // 上报设备信息
  rpc ReportDeviceInfo(ReportDeviceInfoRequest) returns (ReportDeviceInfoResponse);
  
  // 获取设备信息
  rpc GetDeviceInfo(GetDeviceInfoRequest) returns (GetDeviceInfoResponse);
  
  // 删除设备信息
  rpc DeleteDeviceInfo(DeleteDeviceInfoRequest) returns (DeleteDeviceInfoResponse);
}

message ReportDeviceInfoRequest {
  DeviceInfo device_info = 1;
  string user_id = 2;           // 用户ID（已登录时）
  string session_token = 3;     // 会话令牌
}

message ReportDeviceInfoResponse {
  bool success = 1;
  string message = 2;
}
```

### 5.2 服务端处理流程

```go
func (h *DeviceHandler) ReportDeviceInfo(ctx context.Context, req *pb.ReportDeviceInfoRequest) (*pb.ReportDeviceInfoResponse, error) {
    // 1. 验证请求参数
    if req.DeviceInfo == nil || req.DeviceInfo.DeviceId == "" {
        return nil, errors.New("device info is required")
    }
    
    // 2. 验证用户身份（如已登录）
    if req.UserId != "" {
        _, err := h.authService.ValidateToken(ctx, req.SessionToken)
        if err != nil {
            return nil, err
        }
    }
    
    // 3. 转换为领域模型
    device := &domain.Device{
        DeviceID:     req.DeviceInfo.DeviceId,
        Model:        req.DeviceInfo.Model,
        Brand:        req.DeviceInfo.Brand,
        OSVersion:    req.DeviceInfo.OsVersion,
        OSType:       req.DeviceInfo.OsType,
        AppVersion:   req.DeviceInfo.AppVersion,
        BuildNumber:  req.DeviceInfo.BuildNumber,
        ScreenWidth:  int(req.DeviceInfo.ScreenWidth),
        ScreenHeight: int(req.DeviceInfo.ScreenHeight),
        NetworkType:  req.DeviceInfo.NetworkType,
        Timezone:     req.DeviceInfo.Timezone,
        Locale:       req.DeviceInfo.Locale,
        Manufacturer: req.DeviceInfo.Manufacturer,
        Fingerprint:  req.DeviceInfo.Fingerprint,
        UserID:       req.UserId,
    }
    
    // 4. 存储到数据库
    err := h.deviceRepo.Upsert(ctx, device)
    if err != nil {
        return nil, fmt.Errorf("failed to save device info: %w", err)
    }
    
    // 5. 返回响应
    return &pb.ReportDeviceInfoResponse{
        Success: true,
        Message: "Device info reported successfully",
    }, nil
}
```

### 5.3 数据库设计

```sql
CREATE TABLE devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id VARCHAR(64) NOT NULL UNIQUE,
    model VARCHAR(128),
    brand VARCHAR(64),
    os_version VARCHAR(32),
    os_type VARCHAR(16),
    app_version VARCHAR(32),
    build_number VARCHAR(32),
    screen_width INT,
    screen_height INT,
    network_type VARCHAR(16),
    timezone VARCHAR(64),
    locale VARCHAR(32),
    manufacturer VARCHAR(64),
    fingerprint VARCHAR(256),
    user_id UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_devices_device_id ON devices(device_id);
CREATE INDEX idx_devices_user_id ON devices(user_id);
CREATE INDEX idx_devices_os_type ON devices(os_type);
```

---

## 六、数据安全与隐私合规

### 6.1 隐私保护措施

| 措施 | 说明 |
|-----|------|
| **用户授权** | 在首次收集前展示隐私政策，获取用户同意 |
| **数据脱敏** | 不收集敏感信息（如 IMEI、MAC 地址） |
| **加密传输** | 所有数据通过 HTTPS 传输 |
| **数据存储** | 敏感字段加密存储 |
| **数据删除** | 提供 API 允许用户删除设备信息 |
| **数据保留** | 设置数据保留期限，定期清理历史数据 |

### 6.2 合规要求

| 法规 | 要求 |
|-----|------|
| **GDPR** | 用户有权访问、更正、删除其数据 |
| **CCPA** | 告知用户收集的数据类型和用途 |
| **网络安全法** | 数据本地化存储，保护用户隐私 |

---

## 七、错误处理与重试机制

### 7.1 客户端重试策略

| 场景 | 处理方式 |
|-----|------|
| **网络错误** | 保存到本地缓存，网络恢复后重试 |
| **服务端错误** | 指数退避重试（1s → 2s → 4s → 8s） |
| **数据过期** | 超过 7 天未成功上报则放弃 |

### 7.2 服务端错误处理

| 错误类型 | 处理方式 |
|-----|------|
| **参数校验失败** | 返回 400 Bad Request |
| **认证失败** | 返回 401 Unauthorized |
| **服务器错误** | 返回 500 Internal Server Error |
| **数据库错误** | 记录日志，返回通用错误信息 |

---

## 八、监控与日志

### 8.1 客户端日志

```kotlin
// 记录设备信息收集日志
Log.d("DeviceInfo", "Collected device info: $deviceInfo")

// 记录上报结果
if (success) {
    Log.d("DeviceInfo", "Device info reported successfully")
} else {
    Log.e("DeviceInfo", "Failed to report device info: $error")
}
```

### 8.2 服务端监控指标

| 指标 | 说明 |
|-----|------|
| `device_info_reported_total` | 设备信息上报总数 |
| `device_info_report_success_total` | 成功上报数 |
| `device_info_report_failed_total` | 失败上报数 |
| `device_info_report_duration_seconds` | 上报耗时 |
| `active_devices_total` | 活跃设备总数 |
| `devices_by_os_type` | 按操作系统类型分布 |
| `devices_by_app_version` | 按应用版本分布 |

---

## 九、版本演进

| 版本 | 变更内容 |
|-----|---------|
| v1.0 | 基础设备信息收集（型号、版本、屏幕、网络） |
| v1.1 | 添加位置信息（可选）、行为埋点 |
| v1.2 | 支持数据删除、隐私设置 |
| v2.0 | 接入第三方分析平台（如 Firebase Analytics） |

---

## 附录：相关链接

- [AlfQ 安卓客户端设计文档](AlfQ-安卓客户端设计文档.md)
- [AntClaw 用户系统与鉴权](../旧文档/AntClaw-用户系统与鉴权.md)
- [隐私政策模板](https://example.com/privacy)

---

**文档结束**
