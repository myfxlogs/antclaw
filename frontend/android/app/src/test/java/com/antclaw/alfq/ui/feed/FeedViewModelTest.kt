package com.antclaw.alfq.ui.feed

import com.antclaw.alfq.data.local.TokenStoreApi
import com.antclaw.alfq.data.repository.SocialRepository
import com.antclaw.alfq.data.rpc.FeedRpc
import com.antclaw.alfq.data.rpc.ProfileRpc
import com.antclaw.alfq.ui.social.*
import com.connectrpc.ProtocolClientInterface
import io.mockk.mockk
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.*
import org.junit.*
import org.junit.Assert.*
import java.time.Instant

@OptIn(ExperimentalCoroutinesApi::class)
class FeedViewModelTest {

    private val scheduler = TestCoroutineScheduler()
    private val testDispatcher = StandardTestDispatcher(scheduler)

    @Before fun setup() { kotlinx.coroutines.Dispatchers.setMain(testDispatcher) }
    @After fun tearDown() { kotlinx.coroutines.Dispatchers.resetMain() }

    @Test fun `first load succeeds`() = runTest(scheduler) {
        val repo = FakeRepo(listOf(post("p1")))
        val vm = FeedViewModel(repo)
        advanceUntilIdle()
        assertEquals(1, vm.uiState.value.posts.size)
    }

    @Test fun `first load fails`() = runTest(scheduler) {
        val repo = FakeRepo(fail = true)
        val vm = FeedViewModel(repo)
        advanceUntilIdle()
        assertEquals("err", vm.uiState.value.error)
    }

    @Test fun `refresh keeps old data`() = runTest(scheduler) {
        val repo = FakeRepo(listOf(post("p1")))
        val vm = FeedViewModel(repo)
        advanceUntilIdle()
        repo.fail = true; vm.refresh(); advanceUntilIdle()
        assertEquals(1, vm.uiState.value.posts.size)
        assertEquals("err", vm.uiState.value.error)
    }

    @Test fun `loadMore appends`() = runTest(scheduler) {
        val repo = FakeRepo(listOf(post("p1")), "next")
        val vm = FeedViewModel(repo)
        advanceUntilIdle()
        repo.posts = listOf(post("p2")); repo.cursor = null
        vm.loadMore(); advanceUntilIdle()
        assertEquals(2, vm.uiState.value.posts.size)
    }

    @Test fun `loadMore fails`() = runTest(scheduler) {
        val repo = FakeRepo(listOf(post("p1")), "next")
        val vm = FeedViewModel(repo)
        advanceUntilIdle()
        repo.fail = true; vm.loadMore(); advanceUntilIdle()
        assertEquals(1, vm.uiState.value.posts.size)
        assertEquals("err", vm.uiState.value.appendError)
    }

    @Test fun `like optimistic then rollback`() = runTest(scheduler) {
        val repo = FakeRepo(listOf(post("p1")))
        val vm = FeedViewModel(repo)
        advanceUntilIdle()
        repo.failLike = true; vm.toggleLike("p1"); advanceUntilIdle()
        assertFalse(vm.uiState.value.posts[0].isLiked)
        assertEquals(5, vm.uiState.value.posts[0].likeCount)
    }

    @Test fun `share rollback`() = runTest(scheduler) {
        val repo = FakeRepo(listOf(post("p1")))
        val vm = FeedViewModel(repo)
        advanceUntilIdle()
        repo.failShare = true; vm.sharePost("p1"); advanceUntilIdle()
        assertEquals(0, vm.uiState.value.posts[0].shareCount)
    }
}

class FakeRepo(
    var posts: List<PostUi> = emptyList(),
    var cursor: String? = null,
    var fail: Boolean = false,
    var failLike: Boolean = false,
    var failShare: Boolean = false,
) : SocialRepository(
    FeedRpc(mockk<ProtocolClientInterface>(relaxed = true)),
    ProfileRpc(mockk<ProtocolClientInterface>(relaxed = true)),
    mockk<TokenStoreApi>(relaxed = true),
) {
    override suspend fun getFeed(c: String, ps: Int, f: String) =
        if (fail) throw RuntimeException("err") else if (c.isEmpty()) posts to cursor else posts to null
    override suspend fun getPost(id: String) = posts.first { it.postId == id }
    override suspend fun likePost(id: String) =
        if (failLike) throw RuntimeException("err") else post("x").copy(isLiked = true, likeCount = 6)
    override suspend fun unlikePost(id: String) =
        if (failLike) throw RuntimeException("err") else post("x").copy(isLiked = false, likeCount = 4)
    override suspend fun sharePost(id: String, c: String) =
        if (failShare) throw RuntimeException("err") else post("x")
    override suspend fun commentOnPost(a: String, b: String, c: String?) = CommentUi("c1", a, "u", "n", b)
    override suspend fun listComments(a: String, b: String, c: Int) = emptyList<CommentUi>() to null
    override suspend fun createPost(a: String, b: String, c: String, d: Int, e: String) = post("new", a)
    override suspend fun listUserPosts(a: String, b: String, c: Int, d: String) = posts to null
    override suspend fun getProfile(a: String) = TraderProfileUi(a, "T")
    override suspend fun follow(a: String) = 1
    override suspend fun unfollow(a: String) = 0
}

private fun post(id: String, content: String = "t") = PostUi(
    postId = id, authorId = "a", authorName = "A", content = content,
    postType = PostType.TEXT, likeCount = 5, shareCount = 0, createdAt = Instant.now(),
)
