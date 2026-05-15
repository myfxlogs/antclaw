package com.antclaw.alfq.ui.session

import com.antclaw.alfq.data.local.TokenStoreApi
import com.antclaw.alfq.data.rpc.ConnectTransportProvider
import com.antclaw.alfq.data.session.SessionExpiredNotifier
import com.antclaw.alfq.data.sse.SseClient
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.test.*
import org.junit.*
import org.junit.Assert.*
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner

@OptIn(ExperimentalCoroutinesApi::class)
@RunWith(RobolectricTestRunner::class)
class SessionViewModelTest {

    private val scheduler = TestCoroutineScheduler()
    private val dispatcher = StandardTestDispatcher(scheduler)

    @Before fun setup() {
        kotlinx.coroutines.Dispatchers.setMain(dispatcher)
        ConnectTransportProvider.clearToken()
    }
    @After fun tearDown() {
        kotlinx.coroutines.Dispatchers.resetMain()
    }

    @Test fun `login success`() = runTest(scheduler) {
        val vm = SessionViewModel(FakeTok(), FakeSse(), SessionExpiredNotifier())
        advanceUntilIdle()
        vm.onLoginSuccess("u1", "acc", "ref", "Alice")
        advanceUntilIdle()
        assertEquals(SessionState.AUTHENTICATED, vm.session.value.state)
        assertEquals("acc", ConnectTransportProvider.getToken())
    }

    @Test fun `session expired`() = runTest(scheduler) {
        val vm = SessionViewModel(FakeTok("tok", userId = "u1"), FakeSse(), SessionExpiredNotifier())
        advanceUntilIdle()
        vm.onSessionExpired(); advanceUntilIdle()
        assertEquals(SessionState.EXPIRED, vm.session.value.state)
    }

    @Test fun `logout`() = runTest(scheduler) {
        val vm = SessionViewModel(FakeTok("tok", userId = "u1"), FakeSse(), SessionExpiredNotifier())
        advanceUntilIdle()
        vm.logout(); advanceUntilIdle()
        assertEquals(SessionState.UNAUTHENTICATED, vm.session.value.state)
    }

    @Test fun `foreground reconnect`() = runTest(scheduler) {
        val sse = FakeSse()
        val vm = SessionViewModel(FakeTok("tok", userId = "u1"), sse, SessionExpiredNotifier())
        advanceUntilIdle()
        sse.disconnect(); assertFalse(sse.connected)
        vm.onForeground(); advanceUntilIdle()
        assertTrue(sse.connected)
    }

    @Test fun `background disconnect`() = runTest(scheduler) {
        val sse = FakeSse()
        val vm = SessionViewModel(FakeTok("tok", userId = "u1"), sse, SessionExpiredNotifier())
        advanceUntilIdle(); assertTrue(sse.connected)
        vm.onBackground()
        assertFalse(sse.connected)
    }
}

class FakeTok(access: String? = null, userId: String? = null) : TokenStoreApi {
    private var a = access; private var u = userId
    override suspend fun getAccessToken() = a
    override suspend fun getRefreshToken(): String? = null
    override suspend fun getUserId() = u
    override suspend fun saveAccessToken(t: String) { a = t }
    override suspend fun saveRefreshToken(t: String) {}
    override suspend fun saveUserId(id: String) { u = id }
    override suspend fun saveTokens(at: String, rt: String, uid: String) { a = at; u = uid }
    override suspend fun clearTokens() { a = null }
    override suspend fun clearUserId() { u = null }
}

class FakeSse : SseClient {
    var connected = false; private set
    override fun connect() { connected = true }
    override fun disconnect() { connected = false }
    override fun reconnect() { disconnect(); connect() }
    override fun destroy() { connected = false }
}
