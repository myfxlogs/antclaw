package com.antclaw.alfq.data.session

import com.antclaw.alfq.testutil.CoroutineTestBase
import kotlinx.coroutines.*
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.test.*
import org.junit.*
import org.junit.Assert.*

@OptIn(ExperimentalCoroutinesApi::class)
class SessionExpiredNotifierTest : CoroutineTestBase() {

    @Test fun `notify emits to subscriber`() = runTest(scheduler) {
        val notifier = SessionExpiredNotifier()
        var received = false
        val job = launch(kotlinx.coroutines.Dispatchers.Main) {
            notifier.events.collect { received = true }
        }
        runCurrent()
        notifier.notifySessionExpired()
        advanceUntilIdle()
        assertTrue(received)
        job.cancel()
    }

    @Test fun `no subscribers does not crash`() = runTest(scheduler) {
        val notifier = SessionExpiredNotifier()
        notifier.notifySessionExpired()
    }

    @Test fun `multiple subscribers all receive`() = runTest(scheduler) {
        val notifier = SessionExpiredNotifier()
        var count = 0
        val jobs = (1..3).map {
            launch(kotlinx.coroutines.Dispatchers.Main) {
                notifier.events.collect { count++ }
            }
        }
        runCurrent()
        notifier.notifySessionExpired()
        advanceUntilIdle()
        assertEquals(3, count)
        jobs.forEach { it.cancel() }
    }
}
