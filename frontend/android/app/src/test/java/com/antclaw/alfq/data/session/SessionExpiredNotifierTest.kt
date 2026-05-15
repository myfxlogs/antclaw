package com.antclaw.alfq.data.session

import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.launch
import kotlinx.coroutines.test.StandardTestDispatcher
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.runTest
import org.junit.Assert.*
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class SessionExpiredNotifierTest {

    @Test fun `notify emits to subscribers`() = runTest {
        val notifier = SessionExpiredNotifier()
        var received = false
        backgroundScope.launch(StandardTestDispatcher(testScheduler)) {
            notifier.events.collect { received = true }
        }
        notifier.notifySessionExpired()
        advanceUntilIdle()
        assertTrue(received)
    }

    @Test fun `no subscribers does not crash`() = runTest {
        val notifier = SessionExpiredNotifier()
        notifier.notifySessionExpired()
        // No crash = pass
    }

    @Test fun `multiple subscribers all receive`() = runTest {
        val notifier = SessionExpiredNotifier()
        var count = 0
        val td = StandardTestDispatcher(testScheduler)
        repeat(3) {
            backgroundScope.launch(td) {
                notifier.events.collect { count++ }
            }
        }
        notifier.notifySessionExpired()
        advanceUntilIdle()
        assertEquals(3, count)
    }
}
