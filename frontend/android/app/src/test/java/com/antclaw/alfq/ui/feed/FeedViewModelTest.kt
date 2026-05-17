package com.antclaw.alfq.ui.feed

import com.antclaw.alfq.data.error.AppErrorCategory
import com.antclaw.alfq.ui.social.*
import com.antclaw.alfq.testutil.CoroutineTestBase
import com.antclaw.alfq.testutil.FakeSocialRepo
import com.antclaw.alfq.testutil.postUi
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.launch
import kotlinx.coroutines.test.*
import org.junit.*
import org.junit.Assert.*

/**
 * FeedViewModel + TimelineController 集成测试。
 * 覆盖：首屏四态、分页失败、点赞/分享失败回滚。
 */
@OptIn(ExperimentalCoroutinesApi::class)
class FeedViewModelTest : CoroutineTestBase() {

    // ══════ 首屏四态 ══════

    @Test fun `initial load success populates posts`() = runTest(scheduler) {
        val repo = FakeSocialRepo(posts = listOf(postUi("p1"), postUi("p2")))
        val vm = FeedViewModel(repo)
        advanceUntilIdle()
        val s = vm.uiState.value
        assertEquals(InitialPhase.Success, s.initialPhase)
        assertEquals(2, s.posts.size)
        assertNull(s.initialError)
    }

    @Test fun `initial load empty shows Empty phase`() = runTest(scheduler) {
        val repo = FakeSocialRepo(posts = emptyList())
        val vm = FeedViewModel(repo)
        advanceUntilIdle()
        assertEquals(InitialPhase.Empty, vm.uiState.value.initialPhase)
    }

    @Test fun `initial load failure sets Error phase with AppError`() = runTest(scheduler) {
        val repo = FakeSocialRepo(failFirstPage = true)
        val vm = FeedViewModel(repo)
        advanceUntilIdle()
        val s = vm.uiState.value
        assertEquals(InitialPhase.Error, s.initialPhase)
        assertNotNull(s.initialError)
        assertEquals(AppErrorCategory.UNKNOWN, s.initialError!!.category)
    }

    @Test fun `retry after failure re-enters Loading then succeeds`() = runTest(scheduler) {
        val repo = FakeSocialRepo(failFirstPage = true)
        val vm = FeedViewModel(repo)
        advanceUntilIdle()
        assertEquals(InitialPhase.Error, vm.uiState.value.initialPhase)

        repo.failFirstPage = false
        repo.posts = listOf(postUi("ok"))
        vm.retryLoad()
        advanceUntilIdle()
        assertEquals(InitialPhase.Success, vm.uiState.value.initialPhase)
        assertEquals(1, vm.uiState.value.posts.size)
    }

    // ══════ 分页 ══════

    @Test fun `loadMore appends posts and updates cursor`() = runTest(scheduler) {
        val repo = FakeSocialRepo(posts = listOf(postUi("p1")), cursor = "c1")
        val vm = FeedViewModel(repo)
        advanceUntilIdle()
        assertEquals(1, vm.uiState.value.posts.size)
        assertTrue(vm.uiState.value.hasMore)

        repo.posts = listOf(postUi("p2"), postUi("p3"))
        repo.cursor = "c2"
        vm.loadMore()
        advanceUntilIdle()

        val s = vm.uiState.value
        assertEquals(3, s.posts.size)
        assertEquals("c2", s.nextCursor)
        assertEquals(AppendPhase.Idle, s.appendPhase)
    }

    @Test fun `loadMore failure sets appendError without clearing posts`() = runTest(scheduler) {
        val repo = FakeSocialRepo(posts = listOf(postUi("p1")), cursor = "c1")
        val vm = FeedViewModel(repo)
        advanceUntilIdle()
        assertEquals(1, vm.uiState.value.posts.size)

        repo.failAppend = true
        vm.loadMore()
        advanceUntilIdle()

        val s = vm.uiState.value
        assertEquals(AppendPhase.Error, s.appendPhase)
        assertNotNull(s.appendError)
        assertEquals(1, s.posts.size)
    }

    @Test fun `retryLoadMore re-attempts after append failure`() = runTest(scheduler) {
        val repo = FakeSocialRepo(posts = listOf(postUi("p1")), cursor = "c1")
        val vm = FeedViewModel(repo)
        advanceUntilIdle()

        repo.failAppend = true
        vm.loadMore()
        advanceUntilIdle()
        assertEquals(AppendPhase.Error, vm.uiState.value.appendPhase)

        repo.failAppend = false
        repo.posts = listOf(postUi("p2"))
        repo.cursor = "c2"
        vm.retryLoadMore()
        advanceUntilIdle()

        val s = vm.uiState.value
        assertEquals(AppendPhase.Idle, s.appendPhase)
        assertEquals(2, s.posts.size)
    }

    @Test fun `loadMore does nothing when no cursor`() = runTest(scheduler) {
        val repo = FakeSocialRepo(posts = listOf(postUi("p1")), cursor = null)
        val vm = FeedViewModel(repo)
        advanceUntilIdle()
        assertEquals(InitialPhase.Success, vm.uiState.value.initialPhase)
        assertNull(vm.uiState.value.nextCursor)

        vm.loadMore()
        advanceUntilIdle()
        assertEquals(AppendPhase.Idle, vm.uiState.value.appendPhase)
    }

    // ══════ 点赞回滚 ══════

    @Test fun `toggleLike updates optimistically then confirms`() = runTest(scheduler) {
        val repo = FakeSocialRepo(posts = listOf(postUi("p1", likeCount = 5, isLiked = false)))
        val vm = FeedViewModel(repo)
        advanceUntilIdle()

        vm.toggleLike("p1")
        advanceUntilIdle()

        val p = vm.uiState.value.posts.first { it.postId == "p1" }
        assertTrue(p.isLiked)
        assertEquals(6, p.likeCount)
    }

    @Test fun `toggleLike failure rolls back isLiked and likeCount`() = runTest(scheduler) {
        val repo = FakeSocialRepo(posts = listOf(postUi("p1", likeCount = 5, isLiked = false)), failLike = true)
        val vm = FeedViewModel(repo)
        advanceUntilIdle()

        // 先启动收集再触发事件（文档要求：SharedFlow 先 collect 再触发）
        var snackbarEmitted = false
        val job = launch {
            vm.uiEvent.collect { event ->
                if (event is UiEvent.SnackbarRes) snackbarEmitted = true
            }
        }

        vm.toggleLike("p1")
        advanceUntilIdle()

        val p = vm.uiState.value.posts.first { it.postId == "p1" }
        assertFalse(p.isLiked)
        assertEquals(5, p.likeCount)
        assertTrue(snackbarEmitted)
        job.cancel()
    }

    @Test fun `toggleLike unlike rolls back on failure`() = runTest(scheduler) {
        val repo = FakeSocialRepo(posts = listOf(postUi("p1", likeCount = 5, isLiked = true)), failLike = true)
        val vm = FeedViewModel(repo)
        advanceUntilIdle()

        vm.toggleLike("p1")
        advanceUntilIdle()

        val p = vm.uiState.value.posts.first { it.postId == "p1" }
        assertTrue(p.isLiked)
        assertEquals(5, p.likeCount)
    }

    // ══════ 分享回滚 ══════

    @Test fun `sharePost increments shareCount optimistically`() = runTest(scheduler) {
        val repo = FakeSocialRepo(posts = listOf(postUi("p1", shareCount = 0)))
        val vm = FeedViewModel(repo)
        advanceUntilIdle()

        vm.sharePost("p1")
        advanceUntilIdle()

        val p = vm.uiState.value.posts.first { it.postId == "p1" }
        assertEquals(1, p.shareCount)
    }

    @Test fun `sharePost failure rolls back shareCount and emits Snackbar`() = runTest(scheduler) {
        val repo = FakeSocialRepo(posts = listOf(postUi("p1", shareCount = 0)), failShare = true)
        val vm = FeedViewModel(repo)
        advanceUntilIdle()

        // 先启动收集再触发事件
        var snackbarEmitted = false
        val job = launch {
            vm.uiEvent.collect { event ->
                if (event is UiEvent.SnackbarRes) snackbarEmitted = true
            }
        }

        vm.sharePost("p1")
        advanceUntilIdle()

        val p = vm.uiState.value.posts.first { it.postId == "p1" }
        assertEquals(0, p.shareCount)
        assertTrue(snackbarEmitted)
        job.cancel()
    }

    @Test fun `viewModel initializes with recommended tab`() = runTest(scheduler) {
        val repo = FakeSocialRepo(posts = listOf(postUi("p1")))
        val vm = FeedViewModel(repo)
        advanceUntilIdle()
        assertEquals(HomeFeedTab.RECOMMENDED, vm.currentTab)
    }

    @Test fun `selectTab changes current tab and reloads`() = runTest(scheduler) {
        val vm = FeedViewModel(FakeSocialRepo(posts = listOf(postUi("p1"))))
        advanceUntilIdle()
        vm.selectTab(HomeFeedTab.FOLLOWING)
        advanceUntilIdle()
        assertEquals(HomeFeedTab.FOLLOWING, vm.currentTab)
    }
}

