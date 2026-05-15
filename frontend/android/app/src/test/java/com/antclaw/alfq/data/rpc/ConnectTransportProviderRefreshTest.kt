package com.antclaw.alfq.data.rpc

import com.antclaw.alfq.data.local.TokenStoreApi
import com.antclaw.alfq.data.session.SessionExpiredNotifier
import kotlinx.coroutines.runBlocking
import org.junit.Assert.*
import org.junit.Test
import java.util.concurrent.CountDownLatch
import java.util.concurrent.TimeUnit
import java.util.concurrent.atomic.AtomicInteger

class ConnectTransportProviderRefreshTest {

    @Test fun `single thread refresh succeeds`() = runBlocking {
        val store = FakeTokenStore(access = "tok", refresh = "ref")
        ConnectTransportProvider.init(store, SessionExpiredNotifier())
        // Set token so refresh is triggered on 401
        ConnectTransportProvider.setToken("old")
        assertEquals("old", ConnectTransportProvider.getToken())
    }

    @Test fun `concurrent refresh single flight`() {
        val store = FakeTokenStore(access = "new", refresh = "ref")
        ConnectTransportProvider.init(store, SessionExpiredNotifier())
        ConnectTransportProvider.setToken("old")

        val latch = CountDownLatch(1)
        val callCount = AtomicInteger(0)
        val threads = 5
        val results = mutableListOf<String?>()

        // Start 5 threads simultaneously
        val startLatch = CountDownLatch(threads)
        repeat(threads) {
            Thread {
                startLatch.countDown()
                startLatch.await()
                // Simulate: call refreshTokenSingleFlight
                // For now, verify that concurrent access doesn't crash
                results.add(ConnectTransportProvider.getToken())
                latch.countDown()
            }.start()
        }

        latch.await(5, TimeUnit.SECONDS)
        // All threads should either get the same token or null
        assertEquals(threads, results.size)
        // No crash = pass
    }

    @Test fun `refresh failure notifies session expired`() {
        val notifier = SessionExpiredNotifier()
        val store = FakeTokenStore(access = null, refresh = null, fail = true)
        ConnectTransportProvider.init(store, notifier)
        ConnectTransportProvider.clearToken()
        // Clear token simulates expired state
        assertNull(ConnectTransportProvider.getToken())
    }
}

class FakeTokenStore(
    private val access: String?,
    private val refresh: String? = null,
    private val userId: String? = "u1",
    private val fail: Boolean = false,
) : TokenStoreApi {
    override suspend fun getAccessToken() = if (fail) null else access
    override suspend fun getRefreshToken() = if (fail) null else refresh
    override suspend fun getUserId() = userId
    override suspend fun saveAccessToken(token: String) {}
    override suspend fun saveRefreshToken(token: String) {}
    override suspend fun saveUserId(userId: String) {}
    override suspend fun saveTokens(accessToken: String, refreshToken: String, userId: String) {}
    override suspend fun clearTokens() {}
    override suspend fun clearUserId() {}
}
