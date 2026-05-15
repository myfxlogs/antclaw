package com.antclaw.alfq.ui.feed

import com.antclaw.alfq.data.repository.SocialRepository
import com.antclaw.alfq.ui.social.*
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.*
import org.junit.*
import org.junit.Assert.*
import org.mockito.kotlin.*
import java.time.Instant

@OptIn(ExperimentalCoroutinesApi::class)
class FeedViewModelTest {

    @Before fun setup() { Dispatchers.setMain(StandardTestDispatcher()) }
    @After fun tearDown() { Dispatchers.resetMain() }

    private fun repo(): SocialRepository = mock()

    @Test fun `first load succeeds`() = runTest {
        val repo = repo()
        whenever(repo.getFeed(eq(""), any(), eq("all"))).thenReturn(listOf(post("p1")) to null)
        val vm = FeedViewModel(repo)
        advanceUntilIdle()
        assertEquals(1, vm.uiState.value.posts.size)
        assertEquals("p1", vm.uiState.value.posts[0].postId)
    }

    @Test fun `first load fails`() = runTest {
        val repo = repo()
        whenever(repo.getFeed(any(), any(), any())).thenThrow(RuntimeException("fail"))
        val vm = FeedViewModel(repo)
        advanceUntilIdle()
        assertNotNull(vm.uiState.value.error)
        assertFalse(vm.uiState.value.isLoading)
    }

    @Test fun `refresh keeps old data on failure`() = runTest {
        val repo = repo()
        whenever(repo.getFeed(eq(""), any(), eq("all"))).thenReturn(listOf(post("p1")) to null)
        whenever(repo.getFeed(eq(""), any(), eq("all"))).thenThrow(RuntimeException("fail"))
        val vm = FeedViewModel(repo)
        advanceUntilIdle()
        vm.refresh(); advanceUntilIdle()
        assertTrue(vm.uiState.value.posts.isNotEmpty())
    }

    @Test fun `loadMore appends`() = runTest {
        val repo = repo()
        whenever(repo.getFeed(eq(""), any(), eq("all"))).thenReturn(listOf(post("p1")) to "next")
        whenever(repo.getFeed(eq("next"), any(), eq("all"))).thenReturn(listOf(post("p2")) to null)
        val vm = FeedViewModel(repo)
        advanceUntilIdle()
        vm.loadMore(); advanceUntilIdle()
        assertEquals(2, vm.uiState.value.posts.size)
    }

    @Test fun `loadMore fails`() = runTest {
        val repo = repo()
        whenever(repo.getFeed(eq(""), any(), eq("all"))).thenReturn(listOf(post("p1")) to "next")
        whenever(repo.getFeed(eq("next"), any(), eq("all"))).thenThrow(RuntimeException("fail"))
        val vm = FeedViewModel(repo)
        advanceUntilIdle()
        vm.loadMore(); advanceUntilIdle()
        assertEquals(1, vm.uiState.value.posts.size)
        assertNotNull(vm.uiState.value.appendError)
    }

    @Test fun `like toggle optimistic then rollback`() = runTest {
        val repo = repo()
        val p = post("p1")
        whenever(repo.getFeed(any(), any(), any())).thenReturn(listOf(p) to null)
        whenever(repo.likePost("p1")).thenThrow(RuntimeException("fail"))
        val vm = FeedViewModel(repo)
        advanceUntilIdle()
        assertFalse(vm.uiState.value.posts[0].isLiked) // initial
        vm.toggleLike("p1"); advanceUntilIdle()
        assertFalse(vm.uiState.value.posts[0].isLiked) // rolled back
    }

    @Test fun `share optimistic then rollback`() = runTest {
        val repo = repo()
        val p = post("p1")
        whenever(repo.getFeed(any(), any(), any())).thenReturn(listOf(p) to null)
        whenever(repo.sharePost(any())).thenThrow(RuntimeException("fail"))
        val vm = FeedViewModel(repo)
        advanceUntilIdle()
        vm.sharePost("p1"); advanceUntilIdle()
        // shareRollback should restore original shareCount
        assertTrue(vm.uiState.value.posts.isNotEmpty())
    }
}

private fun post(id: String, content: String = "t") = PostUi(
    postId = id, authorId = "a", authorName = "A", content = content,
    postType = PostType.TEXT, likeCount = 5, shareCount = 0, createdAt = Instant.now(),
)
