package com.antclaw.alfq.data.rpc

import com.antclaw.alfq.data.local.TokenStoreApi
import com.antclaw.alfq.data.session.SessionExpiredNotifier
import kotlinx.coroutines.runBlocking
import org.junit.Assert.*
import org.junit.Test
import java.util.concurrent.CountDownLatch
import java.util.concurrent.TimeUnit

class ConnectTransportProviderRefreshTest {

    @Test fun `init sets token provider`() = runBlocking {
        val store = FakeTokenStore(access = "tok")
        ConnectTransportProvider.init(store, SessionExpiredNotifier())
        assertEquals("tok", ConnectTransportProvider.getToken())
    }

    @Test fun `setToken and getToken roundtrip`() {
        ConnectTransportProvider.setToken("abc")
        assertEquals("abc", ConnectTransportProvider.getToken())
    }

    @Test fun `clearToken removes token`() {
        ConnectTransportProvider.setToken("abc")
        ConnectTransportProvider.clearToken()
        assertNull(ConnectTransportProvider.getToken())
    }

    @Test fun `concurrent getToken from multiple threads`() {
        val store = FakeTokenStore(access = "multi")
        ConnectTransportProvider.init(store, SessionExpiredNotifier())
        ConnectTransportProvider.setToken("multi")

        val latch = CountDownLatch(5)
        repeat(5) {
            Thread {
                assertEquals("multi", ConnectTransportProvider.getToken())
                latch.countDown()
            }.start()
        }
        assertTrue(latch.await(5, TimeUnit.SECONDS))
    }
}

class FakeTokenStore(
    private val access: String?,
) : TokenStoreApi {
    override suspend fun getAccessToken() = access
    override suspend fun getRefreshToken(): String? = null
    override suspend fun getUserId(): String? = null
    override suspend fun saveAccessToken(token: String) {}
    override suspend fun saveRefreshToken(token: String) {}
    override suspend fun saveUserId(userId: String) {}
    override suspend fun saveTokens(accessToken: String, refreshToken: String, userId: String) {}
    override suspend fun clearTokens() {}
    override suspend fun clearUserId() {}
}
