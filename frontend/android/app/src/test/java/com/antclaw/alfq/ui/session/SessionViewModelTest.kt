package com.antclaw.alfq.ui.session

import com.antclaw.alfq.data.local.TokenStoreApi
import com.antclaw.alfq.data.repository.DeviceReportApi
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

    @Test fun `login success saves token sets session connects sse and reports device`() = runTest(scheduler) {
        val sse = FakeSse()
        val deviceApi = FakeDeviceReportApi()
        val vm = SessionViewModel(FakeTok(), sse, SessionExpiredNotifier(), deviceApi)
        advanceUntilIdle()
        vm.onLoginSuccess("u1", "acc", "ref", "Alice")
        advanceUntilIdle()
        assertEquals(SessionState.AUTHENTICATED, vm.session.value.state)
        assertEquals("acc", ConnectTransportProvider.getToken())
        assertEquals("Alice", vm.session.value.displayName)
        assertTrue(sse.connected)
        assertTrue(deviceApi.reported)
    }

    @Test fun `device report failure does not block login`() = runTest(scheduler) {
        val sse = FakeSse()
        val deviceApi = FakeDeviceReportApi(fail = true)
        val vm = SessionViewModel(FakeTok(), sse, SessionExpiredNotifier(), deviceApi)
        advanceUntilIdle()
        vm.onLoginSuccess("u1", "acc", "ref", "Alice")
        advanceUntilIdle()
        assertEquals(SessionState.AUTHENTICATED, vm.session.value.state)
        assertTrue(sse.connected)
    }

    @Test fun `session expired`() = runTest(scheduler) {
        val vm = SessionViewModel(FakeTok("tok", userId = "u1"), FakeSse(), SessionExpiredNotifier(), FakeDeviceReportApi())
        advanceUntilIdle()
        vm.onSessionExpired(); advanceUntilIdle()
        assertEquals(SessionState.EXPIRED, vm.session.value.state)
    }

    @Test fun `logout`() = runTest(scheduler) {
        val vm = SessionViewModel(FakeTok("tok", userId = "u1"), FakeSse(), SessionExpiredNotifier(), FakeDeviceReportApi())
        advanceUntilIdle()
        vm.logout(); advanceUntilIdle()
        assertEquals(SessionState.UNAUTHENTICATED, vm.session.value.state)
    }

    @Test fun `foreground reconnect`() = runTest(scheduler) {
        val sse = FakeSse()
        val vm = SessionViewModel(FakeTok("tok", userId = "u1"), sse, SessionExpiredNotifier(), FakeDeviceReportApi())
        advanceUntilIdle()
        sse.disconnect(); assertFalse(sse.connected)
        vm.onForeground(); advanceUntilIdle()
        assertTrue(sse.connected)
    }

    @Test fun `background disconnect`() = runTest(scheduler) {
        val sse = FakeSse()
        val vm = SessionViewModel(FakeTok("tok", userId = "u1"), sse, SessionExpiredNotifier(), FakeDeviceReportApi())
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

class FakeDeviceReportApi(val fail: Boolean = false) : DeviceReportApi {
    var reported = false; private set
    override suspend fun reportDeviceInfo() {
        reported = true
        if (fail) throw RuntimeException("simulated device report failure")
    }
}
