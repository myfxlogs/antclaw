package com.antclaw.alfq.ui.post

import com.antclaw.alfq.data.local.TokenStoreApi
import com.antclaw.alfq.data.rpc.FeedRpc
import com.antclaw.alfq.data.rpc.ProfileRpc
import com.antclaw.alfq.data.repository.SocialRepository
import com.antclaw.alfq.ui.social.*
import com.connectrpc.ProtocolClientInterface
import io.mockk.mockk
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.*
import org.junit.*
import org.junit.Assert.*
import java.time.Instant

@OptIn(ExperimentalCoroutinesApi::class)
class PostViewModelTest {

    private val scheduler = TestCoroutineScheduler()
    private val testDispatcher = StandardTestDispatcher(scheduler)

    @Before fun setup() { kotlinx.coroutines.Dispatchers.setMain(testDispatcher) }
    @After fun tearDown() { kotlinx.coroutines.Dispatchers.resetMain() }

    @Test fun `submit succeeds`() = runTest(scheduler) {
        val vm = PostViewModel(FakePostRepo())
        vm.post(PostDraft(content = "hello"))
        advanceUntilIdle()
        assertTrue(vm.postState.value is PostState.Success)
    }

    @Test fun `submit fails preserves draft for retry`() = runTest(scheduler) {
        val vm = PostViewModel(FakePostRepo(fail = true))
        vm.post(PostDraft(content = "hello", visibility = "followers"))
        advanceUntilIdle()
        assertTrue(vm.postState.value is PostState.Error)
        // 重试应复用上次 payload
        val vm2 = PostViewModel(FakePostRepo())
        vm2.post(PostDraft(content = "hello"))
        advanceUntilIdle()
        assertTrue(vm2.postState.value is PostState.Success)
    }

    @Test fun `retry submits last failed draft`() = runTest(scheduler) {
        val repo = FakePostRepo(fail = true)
        val vm = PostViewModel(repo)
        vm.post(PostDraft(content = "keep me"))
        advanceUntilIdle()
        assertTrue(vm.postState.value is PostState.Error)

        // 切换为成功模式，验证 retry 复用 lastFailedDraft
        repo.fail = false
        vm.retry()
        advanceUntilIdle()
        assertTrue(vm.postState.value is PostState.Success)
    }

    @Test fun `guard prevents submit while loading`() = runTest(scheduler) {
        val repo = FakePostRepo()
        val vm = PostViewModel(repo)
        // 第一次提交触发 Loading
        vm.post(PostDraft(content = "first"))
        advanceUntilIdle()
        assertTrue(vm.postState.value is PostState.Success)
        assertEquals(1, repo.submitCount)
        // 重置后再次提交
        vm.reset()
        vm.post(PostDraft(content = "second"))
        advanceUntilIdle()
        assertEquals(2, repo.submitCount)
    }

    @Test fun `reset clears state`() = runTest(scheduler) {
        val vm = PostViewModel(FakePostRepo())
        vm.post(PostDraft(content = "test"))
        advanceUntilIdle()
        assertTrue(vm.postState.value is PostState.Success)
        vm.reset()
        assertTrue(vm.postState.value is PostState.Idle)
    }
}

class FakePostRepo(
    var fail: Boolean = false,
) : SocialRepository(
    FeedRpc(mockk<ProtocolClientInterface>(relaxed = true)),
    ProfileRpc(mockk<ProtocolClientInterface>(relaxed = true)),
    mockk<TokenStoreApi>(relaxed = true),
) {
    var submitCount = 0

    override suspend fun createPost(a: String, b: String, c: String, d: Int, e: String): PostUi {
        submitCount++
        if (fail) throw RuntimeException("err")
        return PostUi(
            postId = "new", authorId = "u", authorName = "U", content = a,
            postType = PostType.TEXT, createdAt = Instant.now(),
        )
    }

    override suspend fun getFeed(c: String, ps: Int, f: String) = emptyList<PostUi>() to null
    override suspend fun getPost(id: String) = PostUi(id, "a", "A", content = "x", postType = PostType.TEXT, createdAt = Instant.now())
    override suspend fun likePost(id: String) = post("x")
    override suspend fun unlikePost(id: String) = post("x")
    override suspend fun sharePost(id: String, c: String) = post("x")
    override suspend fun commentOnPost(a: String, b: String, c: String?) = CommentUi("c1", a, "u", "n", b)
    override suspend fun listComments(a: String, b: String, c: Int) = emptyList<CommentUi>() to null
    override suspend fun listUserPosts(a: String, b: String, c: Int, d: String) = emptyList<PostUi>() to null
    override suspend fun getProfile(a: String) = TraderProfileUi(a, "T")
    override suspend fun follow(a: String) = 1
    override suspend fun unfollow(a: String) = 0
}

private fun post(id: String, content: String = "t") = PostUi(
    postId = id, authorId = "a", authorName = "A", content = content,
    postType = PostType.TEXT, createdAt = Instant.now(),
)
