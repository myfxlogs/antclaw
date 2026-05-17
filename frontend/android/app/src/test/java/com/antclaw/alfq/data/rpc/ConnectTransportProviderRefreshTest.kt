package com.antclaw.alfq.data.rpc

import com.antclaw.alfq.data.session.SessionExpiredNotifier
import com.antclaw.alfq.testutil.fakeTokenManager
import kotlinx.coroutines.runBlocking
import org.junit.*
import org.junit.Assert.*
import java.util.concurrent.CountDownLatch
import java.util.concurrent.TimeUnit

class TokenManagerTest {

    @Test fun `restore reads token from store`() = runBlocking {
        val (_, tokenManager) = fakeTokenManager(accessToken = "persisted")
        tokenManager.restore(SessionExpiredNotifier())
        assertEquals("persisted", tokenManager.getToken())
    }

    @Test fun `setToken and getToken roundtrip`() {
        val (_, tokenManager) = fakeTokenManager()
        tokenManager.setToken("abc")
        assertEquals("abc", tokenManager.getToken())
    }

    @Test fun `clearToken removes token`() {
        val (_, tokenManager) = fakeTokenManager()
        tokenManager.setToken("x")
        tokenManager.clearToken()
        assertNull(tokenManager.getToken())
    }

    @Test fun `concurrent getToken from 5 threads`() {
        val (_, tokenManager) = fakeTokenManager(accessToken = "multi")
        tokenManager.restore(SessionExpiredNotifier())
        assertEquals("multi", tokenManager.getToken())
        val latch = CountDownLatch(5)
        repeat(5) {
            Thread {
                assertEquals("multi", tokenManager.getToken())
                latch.countDown()
            }.start()
        }
        assertTrue(latch.await(5, TimeUnit.SECONDS))
    }

    @Test fun `restore with no token yields null`() {
        val (_, tokenManager) = fakeTokenManager(accessToken = null)
        tokenManager.restore(SessionExpiredNotifier())
        assertNull(tokenManager.getToken())
    }
}
