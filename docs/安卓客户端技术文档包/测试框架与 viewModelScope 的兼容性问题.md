方案 A：统一使用 MainDispatcherRule
建议在 src/test/java/.../testutil/MainDispatcherRule.kt 新增一个公共测试规则。

示例结构：
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
然后每个 ViewModel 测试统一写：
@get:Rule
val mainDispatcherRule = MainDispatcherRule()
测试中使用：
@Test
fun `xxx`() = runTest {
    val vm = FeedViewModel(repo)
    advanceUntilIdle()
    assertEquals(...)
}
为什么推荐 UnconfinedTestDispatcher
对于当前项目的 ViewModel 测试，UnconfinedTestDispatcher 更适合：

viewModelScope.launch { ... } 会更积极执行。
不容易卡在“状态还没进入错误分支”的中间态。
对 ViewModel 状态机测试更简单。
当前 FeedViewModelTest 已经改成这种方式，说明方向是对的。
适用场景：

ViewModel 初始化即自动 load()。
测试关注最终状态。
需要验证错误路径、回滚路径、SharedFlow 事件。
什么时候用 StandardTestDispatcher
StandardTestDispatcher 也能用，但要更严格控制调度。

适合：

你需要精确断言中间状态，例如先断言 Loading，再推进到 Success/Error。
你需要验证并发、取消、延迟、debounce。
你希望所有协程必须通过 advanceUntilIdle() 或 runCurrent() 才执行。
但如果 ViewModel 初始化时马上 viewModelScope.launch，StandardTestDispatcher 容易出现：
val vm = FeedViewModel(repo)
assertEquals(...)
此时协程可能还没跑，需要：
runCurrent()
advanceUntilIdle()
所以建议：
普通 ViewModel 状态测试：UnconfinedTestDispatcher
精确时序/并发测试：StandardTestDispatcher
最终建议:
优先级 1：新增 MainDispatcherRule，统一替换测试中的 setMain/resetMain。
优先级 2：Feed/Post 错误路径测试使用 UnconfinedTestDispatcher + fake repository。
优先级 3：需要中间态断言的测试单独使用 StandardTestDispatcher + runCurrent/advanceUntilIdle。
优先级 4：只有当业务代码直接使用 Dispatchers.IO/Default 导致测试不可控时，再引入 AppDispatchers。

关键判断
这个问题不需要大改业务架构。
正确修复方向是测试基础设施治理：
统一 Main dispatcher rule。
避免 Robolectric 滥用。
fake repository 驱动错误路径。
SharedFlow 先 collect 再触发事件。
对中间态测试使用 StandardTestDispatcher，对最终状态测试使用 UnconfinedTestDispatcher。
这样可以解除 viewModelScope 与 JVM 测试环境的兼容性问题，同时不会破坏生产代码。