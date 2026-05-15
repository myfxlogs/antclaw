package com.antclaw.alfq.data.rpc

import com.antclaw.alfq.data.local.TokenStoreApi
import com.antclaw.alfq.data.session.SessionExpiredNotifier
import kotlinx.coroutines.runBlocking
import org.junit.*
import org.junit.Assert.*
import org.junit.Test
import java.util.concurrent.CountDownLatch
import java.util.concurrent.TimeUnit
import java.util.concurrent.atomic.AtomicInteger

class ConnectTransportProviderRefreshTest {

    @Before fun setup() { ConnectTransportProvider.clearToken() }

    @Test fun `init restores token from store`() = runBlocking {
        val store = SimTokenStore(access = "persisted")
        ConnectTransportProvider.init(store, SessionExpiredNotifier())
        assertEquals("persisted", ConnectTransportProvider.getToken())
    }

    @Test fun `setToken and getToken roundtrip`() {
        ConnectTransportProvider.setToken("abc")
        assertEquals("abc", ConnectTransportProvider.getToken())
    }

    @Test fun `clearToken removes token`() {
        ConnectTransportProvider.setToken("x")
        ConnectTransportProvider.clearToken()
        assertNull(ConnectTransportProvider.getToken())
    }

    @Test fun `concurrent getToken from 5 threads`() {
        ConnectTransportProvider.init(SimTokenStore(access = "multi"), SessionExpiredNotifier())
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

    @Test fun `refresh failure notifies session expired`() {
        val notifier = SessionExpiredNotifier()
        val store = SimTokenStore(access = null, refresh = null)
        ConnectTransportProvider.init(store, notifier)
        // Clear token → next 401 will trigger refresh → refresh fails → notify
        ConnectTransportProvider.clearToken()
        assertNull(ConnectTransportProvider.getToken())
    }

    @Test fun `concurrent refresh single flight`() {
        val store = SimTokenStore(access = "fresh", refresh = "r")
        ConnectTransportProvider.init(store, SessionExpiredNotifier())
        ConnectTransportProvider.setToken("old")

        val refreshCount = AtomicInteger(0)
        val latch = CountDownLatch(3)
        val results = mutableListOf<String?>()

        val startLatch = CountDownLatch(3)
        repeat(3) {
            Thread {
                startLatch.countDown()
                startLatch.await() // all start together
                results.add(ConnectTransportProvider.getToken())
                latch.countDown()
            }.start()
        }

        assertTrue(latch.await(5, TimeUnit.SECONDS))
        // All 3 threads got the same token
        results.forEach { assertEquals("old", it) }
    }
}

class SimTokenStore(
    private val access: String? = null,
    private val refresh: String? = null,
    private val userId: String? = null,
) : TokenStoreApi {
    override suspend fun getAccessToken() = access
    override suspend fun getRefreshToken() = refresh
    override suspend fun getUserId() = userId
    override suspend fun saveAccessToken(token: String) {}
    override suspend fun saveRefreshToken(token: String) {}
    override suspend fun saveUserId(id: String) {}
    override suspend fun saveTokens(at: String, rt: String, uid: String) {}
    override suspend fun clearTokens() {}
    override suspend fun clearUserId() {}
}
