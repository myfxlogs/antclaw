package com.antclaw.alfq.ui.session

import com.antclaw.alfq.data.session.SessionExpiredNotifier
import com.antclaw.alfq.testutil.CoroutineTestBase
import com.antclaw.alfq.testutil.FakeDeviceReportApi
import com.antclaw.alfq.testutil.FakeSse
import com.antclaw.alfq.testutil.fakeTokenManager
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.*
import org.junit.*
import org.junit.Assert.*
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner

@OptIn(ExperimentalCoroutinesApi::class)
@RunWith(RobolectricTestRunner::class)
class SessionViewModelTest : CoroutineTestBase() {

    @Test fun `init unknown to authenticated when token exists`() = runTest(scheduler) {
        val (tokenStore, tokenManager) = fakeTokenManager(accessToken = "existing-token", userId = "u123")
        val vm = SessionViewModel(tokenStore, tokenManager, FakeSse(), SessionExpiredNotifier(), FakeDeviceReportApi())
        advanceUntilIdle()
        assertEquals(SessionState.AUTHENTICATED, vm.session.value.state)
        assertEquals("u123", vm.session.value.userId)
    }

    @Test fun `init unknown to unauthenticated when no token`() = runTest(scheduler) {
        val (tokenStore, tokenManager) = fakeTokenManager()
        val vm = SessionViewModel(tokenStore, tokenManager, FakeSse(), SessionExpiredNotifier(), FakeDeviceReportApi())
        advanceUntilIdle()
        assertEquals(SessionState.UNAUTHENTICATED, vm.session.value.state)
    }

    @Test fun `login success saves token sets session connects sse and reports device`() = runTest(scheduler) {
        val (tokenStore, tokenManager) = fakeTokenManager()
        val sse = FakeSse()
        val deviceApi = FakeDeviceReportApi()
        val vm = SessionViewModel(tokenStore, tokenManager, sse, SessionExpiredNotifier(), deviceApi)
        advanceUntilIdle()
        vm.onLoginSuccess("u1", "acc", "ref", "Alice")
        advanceUntilIdle()
        assertEquals(SessionState.AUTHENTICATED, vm.session.value.state)
        assertEquals("acc", tokenManager.getToken())
        assertEquals("Alice", vm.session.value.displayName)
        assertTrue(sse.connected)
        assertTrue(deviceApi.reported)
    }

    @Test fun `device report failure does not block login`() = runTest(scheduler) {
        val (tokenStore, tokenManager) = fakeTokenManager()
        val sse = FakeSse()
        val deviceApi = FakeDeviceReportApi(fail = true)
        val vm = SessionViewModel(tokenStore, tokenManager, sse, SessionExpiredNotifier(), deviceApi)
        advanceUntilIdle()
        vm.onLoginSuccess("u1", "acc", "ref", "Alice")
        advanceUntilIdle()
        assertEquals(SessionState.AUTHENTICATED, vm.session.value.state)
        assertTrue(sse.connected)
    }

    @Test fun `session expired`() = runTest(scheduler) {
        val (tokenStore, tokenManager) = fakeTokenManager(accessToken = "tok", userId = "u1")
        val vm = SessionViewModel(tokenStore, tokenManager, FakeSse(), SessionExpiredNotifier(), FakeDeviceReportApi())
        advanceUntilIdle()
        vm.onSessionExpired(); advanceUntilIdle()
        assertEquals(SessionState.EXPIRED, vm.session.value.state)
    }

    @Test fun `logout`() = runTest(scheduler) {
        val (tokenStore, tokenManager) = fakeTokenManager(accessToken = "tok", userId = "u1")
        val vm = SessionViewModel(tokenStore, tokenManager, FakeSse(), SessionExpiredNotifier(), FakeDeviceReportApi())
        advanceUntilIdle()
        vm.logout(); advanceUntilIdle()
        assertEquals(SessionState.UNAUTHENTICATED, vm.session.value.state)
    }

    @Test fun `foreground reconnect`() = runTest(scheduler) {
        val (tokenStore, tokenManager) = fakeTokenManager(accessToken = "tok", userId = "u1")
        val sse = FakeSse()
        val vm = SessionViewModel(tokenStore, tokenManager, sse, SessionExpiredNotifier(), FakeDeviceReportApi())
        advanceUntilIdle()
        sse.disconnect(); assertFalse(sse.connected)
        vm.onForeground(); advanceUntilIdle()
        assertTrue(sse.connected)
    }

    @Test fun `background disconnect`() = runTest(scheduler) {
        val (tokenStore, tokenManager) = fakeTokenManager(accessToken = "tok", userId = "u1")
        val sse = FakeSse()
        val vm = SessionViewModel(tokenStore, tokenManager, sse, SessionExpiredNotifier(), FakeDeviceReportApi())
        advanceUntilIdle(); assertTrue(sse.connected)
        vm.onBackground()
        assertFalse(sse.connected)
    }
}
