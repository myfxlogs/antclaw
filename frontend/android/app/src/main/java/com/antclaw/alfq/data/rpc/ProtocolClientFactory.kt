package com.antclaw.alfq.data.rpc

import antclaw.v1.System
import com.antclaw.alfq.BuildConfig
import com.connectrpc.MethodSpec
import com.connectrpc.ProtocolClientConfig
import com.connectrpc.ProtocolClientInterface
import com.connectrpc.ResponseMessage
import com.connectrpc.StreamType
import com.connectrpc.extensions.GoogleJavaProtobufStrategy
import com.connectrpc.getOrThrow
import com.connectrpc.impl.ProtocolClient
import com.connectrpc.okhttp.ConnectOkHttpClient
import kotlinx.coroutines.runBlocking
import okhttp3.OkHttpClient
import javax.inject.Inject
import javax.inject.Singleton

/**
 * ProtocolClient 工厂。
 *
 * 职责：
 * 1. 创建带认证的 auth ProtocolClient（装配 [AuthInterceptor]）
 * 2. 创建无认证的 noAuth ProtocolClient（用于 healthz / token refresh）
 * 3. 健康检查
 */
@Singleton
class ProtocolClientFactory @Inject constructor(
    authInterceptor: AuthInterceptor,
) {
    val baseUrl: String = BuildConfig.BASE_URL

    private val noAuthOkHttpClient = OkHttpClient()

    private val authOkHttpClient =
        OkHttpClient.Builder().addInterceptor(authInterceptor).build()

    private val authProtocolClient: ProtocolClientInterface by lazy {
        ProtocolClient(
            httpClient = ConnectOkHttpClient(unaryClient = authOkHttpClient),
            config = ProtocolClientConfig(
                host = baseUrl,
                serializationStrategy = GoogleJavaProtobufStrategy(),
            ),
        )
    }

    private val noAuthProtocolClient: ProtocolClientInterface by lazy {
        ProtocolClient(
            httpClient = ConnectOkHttpClient(unaryClient = noAuthOkHttpClient),
            config = ProtocolClientConfig(
                host = baseUrl,
                serializationStrategy = GoogleJavaProtobufStrategy(),
            ),
        )
    }

    fun create(): ProtocolClientInterface = authProtocolClient

    fun healthz(): Result<String> {
        return runBlocking {
            try {
                val spec = MethodSpec(
                    path = "antclaw.v1.SystemService/Healthz",
                    requestClass = System.HealthzRequest::class,
                    responseClass = System.HealthzResponse::class,
                    streamType = StreamType.UNARY,
                )
                val resp: ResponseMessage<System.HealthzResponse> = noAuthProtocolClient.unary(
                    System.HealthzRequest.getDefaultInstance(), emptyMap(), spec
                )
                Result.success(resp.getOrThrow().status)
            } catch (e: Exception) { Result.failure(e) }
        }
    }
}
