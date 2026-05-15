package com.antclaw.alfq.ui.session

import com.antclaw.alfq.data.local.TokenStoreApi
import com.antclaw.alfq.data.rpc.ConnectTransportProvider
import com.antclaw.alfq.data.session.SessionExpiredNotifier
import com.antclaw.alfq.data.sse.SseClient
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.test.*
import kotlin.coroutines.cancellation.CancellationException
import kotlinx.coroutines.Job
import org.junit.*
import org.junit.Assert.*
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner

@OptIn(ExperimentalCoroutinesApi::class)
@RunWith(RobolectricTestRunner::class)
class SessionViewModelTest {

    private val scheduler = TestCoroutineScheduler()
    private val testDispatcher = StandardTestDispatcher(scheduler)

    @Before fun setup() {
        kotlinx.coroutines.Dispatchers.setMain(testDispatcher)
        ConnectTransportProvider.clearToken()
    }

    @After fun tearDown() {
        kotlinx.coroutines.Dispatchers.resetMain()
    }

    private fun SessionViewModel.clearScope() {
        viewModelScope.coroutineContext[Job]?.cancelChildren(CancellationException("test done"))
    }

    // ═══════════════════════════════════════════
    // 初始化 - 无 Token
    // ═══════════════════════════════════════════

    @Test fun `init without token`() = runTest(testDispatcher.scheduler) {
        val tokenStore = FakeTokenStore(accessToken = null)
        val sseClient = FakeSseClient()
        val notifier = SessionExpiredNotifier()

        val vm = SessionViewModel(tokenStore, sseClient, notifier)
        runCurrent()
        advanceUntilIdle()

        assertEquals(SessionState.UNAUTHENTICATED, vm.session.first().state)
        assertEquals("", vm.session.first().userId)
        assertFalse(sseClient.connected)
                vm.clearScope()
    }

    // ═══════════════════════════════════════════
    // 初始化 - 有 Token
    // ═══════════════════════════════════════════

    @Test fun `init with valid token`() = runTest(testDispatcher.scheduler) {
        val tokenStore = FakeTokenStore(accessToken = "tok", userId = "u1")
        val sseClient = FakeSseClient()
        val notifier = SessionExpiredNotifier()

        val vm = SessionViewModel(tokenStore, sseClient, notifier)
        runCurrent()
        advanceUntilIdle()

        assertEquals(SessionState.AUTHENTICATED, vm.session.first().state)
        assertEquals("u1", vm.session.first().userId)
        assertTrue(sseClient.connected)
    }

    // ═══════════════════════════════════════════
    // 登录成功
    // ═══════════════════════════════════════════

    @Test fun `login success`() = runTest(testDispatcher.scheduler) {
        val tokenStore = FakeTokenStore()
        val sseClient = FakeSseClient()
        val notifier = SessionExpiredNotifier()

        val vm = SessionViewModel(tokenStore, sseClient, notifier)
        advanceUntilIdle()

        vm.onLoginSuccess("u1", "acc", "ref", "Alice")
        advanceUntilIdle()

        assertEquals(SessionState.AUTHENTICATED, vm.session.value.state)
        assertEquals("u1", vm.session.value.userId)
        assertEquals("Alice", vm.session.value.displayName)
        assertEquals("acc", ConnectTransportProvider.getToken())
        assertTrue(sseClient.connected)
    }

    // ═══════════════════════════════════════════
    // Session 过期
    // ═══════════════════════════════════════════

    @Test fun `session expired emits RequireLogin`() = runTest(testDispatcher.scheduler) {
        val tokenStore = FakeTokenStore(accessToken = "tok", userId = "u1")
        val sseClient = FakeSseClient()
        val notifier = SessionExpiredNotifier()

        val vm = SessionViewModel(tokenStore, sseClient, notifier)
        runCurrent(); advanceUntilIdle()
        assertEquals(SessionState.AUTHENTICATED, vm.session.value.state)

        vm.onSessionExpired()
        advanceUntilIdle()

        assertEquals(SessionState.EXPIRED, vm.session.value.state)
        assertEquals(SessionEvent.RequireLogin, vm.events.first())
        assertFalse(sseClient.connected)
    }

    // ═══════════════════════════════════════════
    // 登出
    // ═══════════════════════════════════════════

    @Test fun `logout clears session`() = runTest(testDispatcher.scheduler) {
        val tokenStore = FakeTokenStore(accessToken = "tok", userId = "u1")
        val sseClient = FakeSseClient()
        val notifier = SessionExpiredNotifier()

        val vm = SessionViewModel(tokenStore, sseClient, notifier)
        runCurrent(); advanceUntilIdle()

        vm.logout()
        advanceUntilIdle()

        assertEquals(SessionState.UNAUTHENTICATED, vm.session.value.state)
        assertEquals(SessionEvent.LoggedOut, vm.events.first())
        assertFalse(sseClient.connected)
    }

    // ═══════════════════════════════════════════
    // 前后台
    // ═══════════════════════════════════════════

    @Test fun `foreground reconnects if authenticated`() = runTest(testDispatcher.scheduler) {
        val tokenStore = FakeTokenStore(accessToken = "tok", userId = "u1")
        val sseClient = FakeSseClient()
        val notifier = SessionExpiredNotifier()

        val vm = SessionViewModel(tokenStore, sseClient, notifier)
        runCurrent(); advanceUntilIdle()
        sseClient.disconnect()
        assertFalse(sseClient.connected)

        vm.onForeground()
        advanceUntilIdle()
        assertTrue(sseClient.connected)
    }

    @Test fun `background disconnects SSE`() = runTest(testDispatcher.scheduler) {
        val tokenStore = FakeTokenStore(accessToken = "tok", userId = "u1")
        val sseClient = FakeSseClient()
        val notifier = SessionExpiredNotifier()

        val vm = SessionViewModel(tokenStore, sseClient, notifier)
        advanceUntilIdle()
        assertTrue(sseClient.connected)

        vm.onBackground()
        assertFalse(sseClient.connected)
    }
}

// ── Fakes ──

class FakeTokenStore(
    private var accessToken: String? = null,
    private var refreshToken: String? = null,
    private var userId: String? = null,
) : TokenStoreApi {
    override suspend fun getAccessToken() = accessToken
    override suspend fun getRefreshToken() = refreshToken
    override suspend fun getUserId() = userId
    override suspend fun saveAccessToken(token: String) { accessToken = token }
    override suspend fun saveRefreshToken(token: String) { refreshToken = token }
    override suspend fun saveUserId(id: String) { userId = id }
    override suspend fun saveTokens(at: String, rt: String, uid: String) {
        accessToken = at; refreshToken = rt; userId = uid
    }
    override suspend fun clearTokens() { accessToken = null; refreshToken = null }
    override suspend fun clearUserId() { userId = null }
}

class FakeSseClient : SseClient {
    var connected = false
        private set
    var connectCount = 0
        private set
    var disconnectCount = 0
        private set

    override fun connect() { connected = true; connectCount++ }
    override fun disconnect() { connected = false; disconnectCount++ }
    override fun reconnect() { disconnect(); connect() }
    override fun destroy() { connected = false }
}
