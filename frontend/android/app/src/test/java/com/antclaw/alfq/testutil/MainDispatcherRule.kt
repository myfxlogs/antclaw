package com.antclaw.alfq.testutil

import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.TestDispatcher
import kotlinx.coroutines.test.UnconfinedTestDispatcher
import kotlinx.coroutines.test.resetMain
import kotlinx.coroutines.test.setMain
import org.junit.rules.TestWatcher
import org.junit.runner.Description

/**
 * JUnit 测试规则 — 将 [Dispatchers.Main] 替换为 [TestDispatcher]。
 *
 * 用法：
 * ```
 * @get:Rule
 * val mainDispatcherRule = MainDispatcherRule()
 *
 * @Test
 * fun `my test`() = runTest {
 *     val vm = MyViewModel(repo)
 *     advanceUntilIdle()
 *     assertEquals(...)
 * }
 * ```
 *
 * 默认使用 [UnconfinedTestDispatcher]，适合 ViewModel 最终状态断言。
 * 需要精确中间态断言时，传入 [kotlinx.coroutines.test.StandardTestDispatcher]：
 * ```
 * @get:Rule
 * val mainDispatcherRule = MainDispatcherRule(StandardTestDispatcher())
 * ```
 */
@OptIn(ExperimentalCoroutinesApi::class)
class MainDispatcherRule(
    val testDispatcher: TestDispatcher = UnconfinedTestDispatcher(),
) : TestWatcher() {
    override fun starting(description: Description) {
        Dispatchers.setMain(testDispatcher)
    }

    override fun finished(description: Description) {
        Dispatchers.resetMain()
    }
}
