package com.antclaw.alfq.data.sse

import org.junit.Assert.*
import org.junit.Test

class SseManagerBackoffTest {

    @Test fun `retry 1 → 3s`() = assertEquals(3000L, SseManager.backoffDelay(1))
    @Test fun `retry 2 → 6s`() = assertEquals(6000L, SseManager.backoffDelay(2))
    @Test fun `retry 3 → 12s`() = assertEquals(12000L, SseManager.backoffDelay(3))
    @Test fun `retry 4 → 24s`() = assertEquals(24000L, SseManager.backoffDelay(4))
    @Test fun `retry 5 → 30s (capped)`() = assertEquals(30000L, SseManager.backoffDelay(5))
    @Test fun `retry 10 → still 30s`() = assertEquals(30000L, SseManager.backoffDelay(10))
    @Test fun `custom base 1000 retry 3 → 4s`() = assertEquals(4000L, SseManager.backoffDelay(3, baseMs = 1000))
}
