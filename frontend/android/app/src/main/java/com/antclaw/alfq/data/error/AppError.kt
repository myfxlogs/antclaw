package com.antclaw.alfq.data.error

import android.util.Log
import com.antclaw.alfq.R
import com.connectrpc.Code
import com.connectrpc.ConnectException
import java.io.IOException
import java.net.SocketTimeoutException
import java.net.UnknownHostException

/**
 * 错误领域分类 — 在 RPC/Repository 边界将 Throwable 映射为用户可读错误。
 * ViewModel 只暴露 [AppError.userMessageRes] 对应的 string resource id，
 * 技术详情仅写入日志，不进入 UI。
 */
enum class AppErrorCategory {
    /** 网络层错误：无网、DNS、连接超时等 */
    NETWORK,
    /** 认证/授权错误：token 过期、权限不足 */
    AUTH,
    /** 输入校验失败 */
    VALIDATION,
    /** 服务端错误：5xx / 内部异常 */
    SERVER,
    /** 无法分类的未知错误 */
    UNKNOWN,
}

data class AppError(
    val category: AppErrorCategory,
    /** 对应 strings.xml 中的 resource id，供 UI 直接使用 */
    val userMessageRes: Int,
    /** 技术日志（仅写 Log，不进入 UI） */
    val technicalDetail: String,
)

/**
 * 在 RPC/Repository 边界调用：将 [Throwable] 映射为 [AppError]。
 * 优先匹配 Connect-RPC [ConnectException] 的 [Code] 枚举，回退到网络异常类型。
 */
fun Throwable.toAppError(): AppError {
    val detail = this.message ?: this::class.java.simpleName
    try { Log.w("AppError", "Mapping error: $detail", this) } catch (_: RuntimeException) { /* stub in unit test */ }

    return when {
        // ── Connect-RPC 协议异常（优先） ──
        this is ConnectException -> mapConnectCode(this.code, detail)

        // ── 网络层 ──
        this is UnknownHostException ||
        this is java.net.ConnectException ||  // java.net.ConnectException（连接拒绝）
        this is SocketTimeoutException ||
        this is IOException -> AppError(
            category = AppErrorCategory.NETWORK,
            userMessageRes = R.string.error_network,
            technicalDetail = detail,
        )

        // ── 兜底 ──
        else -> AppError(
            category = AppErrorCategory.UNKNOWN,
            userMessageRes = R.string.error_unknown,
            technicalDetail = detail,
        )
    }
}

/**
 * 将 Connect-RPC [Code] 映射为 [AppError]。
 * 遵循 gRPC 状态码语义：
 *   - 16 (UNAUTHENTICATED) / 7 (PERMISSION_DENIED) → AUTH
 *   - 3 (INVALID_ARGUMENT) / 9 (FAILED_PRECONDITION) / 11 (OUT_OF_RANGE) → VALIDATION
 *   - 12 (UNIMPLEMENTED) / 13 (INTERNAL) / 14 (UNAVAILABLE) / 15 (DATA_LOSS) → SERVER
 *   - 其余 → UNKNOWN
 */
private fun mapConnectCode(code: Code, detail: String): AppError {
    return when (code) {
        Code.UNAUTHENTICATED,
        Code.PERMISSION_DENIED -> AppError(
            category = AppErrorCategory.AUTH,
            userMessageRes = R.string.error_auth_expired,
            technicalDetail = detail,
        )
        Code.INVALID_ARGUMENT,
        Code.FAILED_PRECONDITION,
        Code.OUT_OF_RANGE -> AppError(
            category = AppErrorCategory.VALIDATION,
            userMessageRes = R.string.error_validation,
            technicalDetail = detail,
        )
        Code.UNIMPLEMENTED,
        Code.INTERNAL_ERROR,
        Code.UNAVAILABLE,
        Code.DATA_LOSS -> AppError(
            category = AppErrorCategory.SERVER,
            userMessageRes = R.string.error_server,
            technicalDetail = detail,
        )
        // NOT_FOUND, ALREADY_EXISTS, ABORTED, CANCELLED, DEADLINE_EXCEEDED,
        // RESOURCE_EXHAUSTED, UNKNOWN, OK — 归类为未知
        else -> AppError(
            category = AppErrorCategory.UNKNOWN,
            userMessageRes = R.string.error_unknown,
            technicalDetail = detail,
        )
    }
}

/**
 * 便捷方法：将异常映射为 AppError 并返回用户可读消息的 string resource id。
 */
fun Throwable.toUserErrorRes(): Int = this.toAppError().userMessageRes
