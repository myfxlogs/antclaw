package com.antclaw.alfq.testutil

import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.StandardTestDispatcher
import kotlinx.coroutines.test.TestCoroutineScheduler
import org.junit.Rule

/**
 * 协程测试基类 — 提供共享的 [TestCoroutineScheduler] 和 [MainDispatcherRule]。
 *
 * 用法：
 * ```
 * class MyTest : CoroutineTestBase() {
 *     @Test fun `case`() = runTest(scheduler) {
 *         ...
 *         advanceUntilIdle()
 *     }
 * }
 * ```
 *
 * 省去子类中重复的：
 * ```
 * private val scheduler = TestCoroutineScheduler()
 * @get:Rule val mainDispatcherRule = MainDispatcherRule(StandardTestDispatcher(scheduler))
 * ```
 */
@OptIn(ExperimentalCoroutinesApi::class)
abstract class CoroutineTestBase {
    protected val scheduler = TestCoroutineScheduler()

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule(StandardTestDispatcher(scheduler))
}
