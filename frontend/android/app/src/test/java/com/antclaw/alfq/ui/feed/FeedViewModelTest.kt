package com.antclaw.alfq.ui.feed

import com.antclaw.alfq.data.local.TokenStore
import com.antclaw.alfq.data.repository.SocialRepository
import com.antclaw.alfq.data.rpc.FeedRpc
import com.antclaw.alfq.data.rpc.ProfileRpc
import com.connectrpc.ProtocolClientInterface
import io.mockk.mockk
import com.antclaw.alfq.ui.social.*
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.*
import org.junit.*
import org.junit.Assert.*
import java.time.Instant

@OptIn(ExperimentalCoroutinesApi::class)
class FeedViewModelTest {

    @Before fun setup() { Dispatchers.setMain(StandardTestDispatcher()) }
    @After fun tearDown() { Dispatchers.resetMain() }

    @Test fun `first load succeeds`() = runTest {
        val repo = stubRepo(listOf(post("p1")))
        val vm = FeedViewModel(repo)
        advanceUntilIdle()
        assertEquals(1, vm.uiState.value.posts.size)
        assertEquals("p1", vm.uiState.value.posts[0].postId)
    }

    @Test fun `first load fails`() = runTest {
        val repo = stubRepo(fail = true)
        val vm = FeedViewModel(repo)
        advanceUntilIdle()
        assertNotNull(vm.uiState.value.error)
    }

    @Test fun `refresh keeps old data on failure`() = runTest {
        val repo = stubRepo(listOf(post("p1")))
        val vm = FeedViewModel(repo)
        advanceUntilIdle()
        repo.fail = true; vm.refresh(); advanceUntilIdle()
        assertTrue(vm.uiState.value.posts.isNotEmpty())
        assertNotNull(vm.uiState.value.error)
    }

    @Test fun `loadMore appends`() = runTest {
        val repo = stubRepo(listOf(post("p1")), "next")
        val vm = FeedViewModel(repo)
        advanceUntilIdle()
        repo.posts = listOf(post("p2")); repo.cursor = null
        vm.loadMore(); advanceUntilIdle()
        assertEquals(2, vm.uiState.value.posts.size)
    }

    @Test fun `loadMore fails`() = runTest {
        val repo = stubRepo(listOf(post("p1")), "next")
        val vm = FeedViewModel(repo)
        advanceUntilIdle()
        repo.fail = true; vm.loadMore(); advanceUntilIdle()
        assertEquals(1, vm.uiState.value.posts.size)
        assertNotNull(vm.uiState.value.appendError)
    }

    @Test fun `like optimistic then rollback`() = runTest {
        val repo = stubRepo(listOf(post("p1")))
        val vm = FeedViewModel(repo)
        advanceUntilIdle()
        repo.failLike = true; vm.toggleLike("p1"); advanceUntilIdle()
        assertFalse(vm.uiState.value.posts[0].isLiked)
        assertEquals(5, vm.uiState.value.posts[0].likeCount)
    }

    @Test fun `share rollback on failure`() = runTest {
        val repo = stubRepo(listOf(post("p1")))
        val vm = FeedViewModel(repo)
        advanceUntilIdle()
        repo.failShare = true; vm.sharePost("p1"); advanceUntilIdle()
        assertEquals(0, vm.uiState.value.posts[0].shareCount)
    }
}

// ── Stub: real subclass, no mocking ──

class StubSocialRepo(
    var posts: List<PostUi> = emptyList(),
    var cursor: String? = null,
    var fail: Boolean = false,
    var failLike: Boolean = false,
    var failShare: Boolean = false,
) : SocialRepository(
    FeedRpc(mockk<ProtocolClientInterface>(relaxed = true)),
    ProfileRpc(mockk<ProtocolClientInterface>(relaxed = true)),
    null as TokenStore,
) {
    override suspend fun getFeed(c: String, ps: Int, f: String) = if (fail) throw RuntimeException("err") else if (c.isEmpty()) posts to cursor else posts to null
    override suspend fun getPost(id: String) = posts.first { it.postId == id }
    override suspend fun likePost(id: String) = if (failLike) throw RuntimeException("like fail") else post("x")
    override suspend fun unlikePost(id: String) = if (failLike) throw RuntimeException("unlike fail") else post("x")
    override suspend fun sharePost(id: String, c: String) = if (failShare) throw RuntimeException("share fail") else post("x")
    override suspend fun commentOnPost(a: String, b: String, c: String?) = CommentUi("c1", a, "u", "n", b)
    override suspend fun listComments(a: String, b: String, c: Int) = emptyList<CommentUi>() to null
    override suspend fun createPost(a: String, b: String, c: String, d: Int, e: String) = post("new", a)
    override suspend fun listUserPosts(a: String, b: String, c: Int, d: String) = posts to null
    override suspend fun getProfile(a: String) = TraderProfileUi(a, "T")
    override suspend fun follow(a: String) = 1
    override suspend fun unfollow(a: String) = 0
}

private fun stubRepo(posts: List<PostUi> = emptyList(), cursor: String? = null, fail: Boolean = false) =
    StubSocialRepo(posts, cursor, fail)

private fun post(id: String, content: String = "t") = PostUi(
    postId = id, authorId = "a", authorName = "A", content = content,
    postType = PostType.TEXT, likeCount = 5, shareCount = 0, createdAt = Instant.now(),
)
