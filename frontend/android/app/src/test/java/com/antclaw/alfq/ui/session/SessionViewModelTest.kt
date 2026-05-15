package com.antclaw.alfq.ui.session

import com.antclaw.alfq.data.local.TokenStore
import com.antclaw.alfq.data.sse.SseClient
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.test.StandardTestDispatcher
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.resetMain
import kotlinx.coroutines.test.runTest
import kotlinx.coroutines.test.setMain
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class SessionViewModelTest {

    private val testDispatcher = StandardTestDispatcher()

    @Before
    fun setup() {
        Dispatchers.setMain(testDispatcher)
    }

    @After
    fun tearDown() {
        Dispatchers.resetMain()
    }

    @Test
    fun `init with no token sets UNAUTHENTICATED`() = runTest {
        val vm = SessionViewModel(FakeTokenStore(null, null), FakeSseManager())
        advanceUntilIdle()
        assertEquals(SessionState.UNAUTHENTICATED, vm.session.first().state)
    }

    @Test
    fun `init with valid token sets AUTHENTICATED and connects SSE`() = runTest {
        val sse = FakeSseManager()
        val vm = SessionViewModel(
            FakeTokenStore("access-token", "user-1"),
            sse,
        )
        advanceUntilIdle()
        assertEquals(SessionState.AUTHENTICATED, vm.session.first().state)
        assertEquals("user-1", vm.session.first().userId)
        assertTrue(sse.connectCalled)
    }

    @Test
    fun `login success updates session and connects SSE`() = runTest {
        val sse = FakeSseManager()
        val vm = SessionViewModel(FakeTokenStore(null, null), sse)
        advanceUntilIdle()
        assertEquals(SessionState.UNAUTHENTICATED, vm.session.first().state)

        vm.onLoginSuccess("user-2", "new-token", "refresh-token")
        advanceUntilIdle()
        assertEquals(SessionState.AUTHENTICATED, vm.session.first().state)
        assertEquals("user-2", vm.session.first().userId)
        assertTrue(sse.connectCalled)
    }

    @Test
    fun `session expired clears tokens and emits RequireLogin`() = runTest {
        val tokenStore = FakeTokenStore("token", "user-1")
        val sse = FakeSseManager()
        val vm = SessionViewModel(tokenStore, sse)
        advanceUntilIdle()

        vm.onSessionExpired()
        advanceUntilIdle()
        assertEquals(SessionState.EXPIRED, vm.session.first().state)
        assertTrue(sse.disconnectCalled)

        val event = vm.events.first()
        assertEquals(SessionEvent.RequireLogin, event)
    }

    @Test
    fun `logout clears session and emits LoggedOut`() = runTest {
        val tokenStore = FakeTokenStore("token", "user-1")
        val sse = FakeSseManager()
        val vm = SessionViewModel(tokenStore, sse)
        advanceUntilIdle()
        assertEquals(SessionState.AUTHENTICATED, vm.session.first().state)

        vm.logout()
        advanceUntilIdle()
        assertEquals(SessionState.UNAUTHENTICATED, vm.session.first().state)
        assertTrue(sse.disconnectCalled)

        val event = vm.events.first()
        assertEquals(SessionEvent.LoggedOut, event)
    }

    @Test
    fun `foreground reconnects SSE when authenticated`() = runTest {
        val sse = FakeSseManager()
        val vm = SessionViewModel(FakeTokenStore("token", "user-1"), sse)
        advanceUntilIdle()

        sse.connectCalled = false
        vm.onForeground()
        advanceUntilIdle()
        assertTrue(sse.reconnectCalled)
    }

    @Test
    fun `foreground does not reconnect when unauthenticated`() = runTest {
        val sse = FakeSseManager()
        val vm = SessionViewModel(FakeTokenStore(null, null), sse)
        advanceUntilIdle()

        vm.onForeground()
        advanceUntilIdle()
        assertFalse(sse.reconnectCalled)
    }

    @Test
    fun `background disconnects SSE`() = runTest {
        val sse = FakeSseManager()
        val vm = SessionViewModel(FakeTokenStore("token", "user-1"), sse)
        advanceUntilIdle()

        vm.onBackground()
        advanceUntilIdle()
        assertTrue(sse.disconnectCalled)
    }
}

// ── Fakes ──

class FakeTokenStore(
    private var token: String?,
    private var userId: String?,
) : TokenStore(error("context not needed")) {
    private var refreshToken: String? = null
    var tokensCleared = false

    override suspend fun getAccessToken(): String? = token
    override suspend fun getRefreshToken(): String? = refreshToken
    override suspend fun getUserId(): String? = userId
    override suspend fun saveAccessToken(t: String) { token = t }
    override suspend fun saveTokens(access: String, refresh: String, uid: String) {
        token = access; refreshToken = refresh; userId = uid
    }
    override suspend fun clearTokens() { tokensCleared = true; token = null; refreshToken = null }
    fun clearUserId() { userId = null }
}

class FakeSseManager : SseClient {
    var connectCalled = false
    var disconnectCalled = false
    var reconnectCalled = false

    override fun connect() { connectCalled = true }
    override fun disconnect() { disconnectCalled = true }
    override fun reconnect() { reconnectCalled = true }
    override fun destroy() { disconnectCalled = true }
}
