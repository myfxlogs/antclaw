package com.antclaw.alfq.data.session

import kotlinx.coroutines.*
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.test.*
import org.junit.*
import org.junit.Assert.*

@OptIn(ExperimentalCoroutinesApi::class)
class SessionExpiredNotifierTest {

    private val testDispatcher = StandardTestDispatcher()

    @Before fun setup() { Dispatchers.setMain(testDispatcher) }
    @After fun tearDown() { Dispatchers.resetMain() }

    @Test fun `notify emits to subscriber`() = runTest {
        val notifier = SessionExpiredNotifier()
        var received = false
        val job = launch(Dispatchers.Main) {
            notifier.events.collect { received = true }
        }
        runCurrent() // let subscriber start
        notifier.notifySessionExpired()
        advanceUntilIdle()
        assertTrue(received)
        job.cancel()
    }

    @Test fun `no subscribers does not crash`() = runTest {
        val notifier = SessionExpiredNotifier()
        notifier.notifySessionExpired()
        // No crash = pass
    }

    @Test fun `multiple subscribers all receive`() = runTest {
        val notifier = SessionExpiredNotifier()
        var count = 0
        val jobs = (1..3).map {
            launch(Dispatchers.Main) {
                notifier.events.collect { count++ }
            }
        }
        runCurrent() // let subscribers start
        notifier.notifySessionExpired()
        advanceUntilIdle()
        assertEquals(3, count)
        jobs.forEach { it.cancel() }
    }
}
