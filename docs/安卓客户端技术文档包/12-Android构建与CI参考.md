# Android 客户端构建与 CI 命令

> 目标读者：AI Agent、CI Runner  
> 工作空间：`frontend/android/`

## 本地构建命令

```bash
cd frontend/android

# Debug 构建
./gradlew :app:assembleDebug

# Release 构建（需签名配置）
./gradlew :app:assembleRelease

# 单元测试
./gradlew :app:testDebugUnitTest

# Lint 检查
./gradlew :app:lintDebug

# 全部检查（构建 + 测试 + Lint）
./gradlew :app:assembleDebug :app:testDebugUnitTest :app:lintDebug
```

## 签名要求

- Release 构建依赖 `app/antclaw-release.jks`
- 签名信息从 `local.properties` 读取：`RELEASE_STORE_FILE`、`RELEASE_STORE_PASSWORD`、`RELEASE_KEY_ALIAS`、`RELEASE_KEY_PASSWORD`
- 缺少签名配置时 Release 构建会 fail-fast，不会产出不可发布包

## 依赖版本管理

- 依赖版本集中在 `gradle/libs.versions.toml`（如存在），或 `app/build.gradle.kts`
- 新增依赖必须声明明确版本号

## 资源压缩

- Release 构建启用 `isMinifyEnabled = true`（ProGuard）
- `proguard-rules.pro` 位于 `app/` 目录

## CI 快速参考

| 步骤 | 命令 |
|------|------|
| 编译检查 | `./gradlew :app:assembleDebug` |
| 单元测试 | `./gradlew :app:testDebugUnitTest` |
| Lint | `./gradlew :app:lintDebug` |
| Release | `./gradlew :app:assembleRelease` |
