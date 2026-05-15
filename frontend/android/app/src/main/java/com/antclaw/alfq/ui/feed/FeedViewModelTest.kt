package com.antclaw.alfq.ui.feed

import com.antclaw.alfq.data.repository.SocialRepository
import com.antclaw.alfq.ui.social.CommentUi
import com.antclaw.alfq.ui.social.PostType
import com.antclaw.alfq.ui.social.PostUi
import com.antclaw.alfq.ui.social.PostVisibility
import com.antclaw.alfq.ui.social.TraderProfileUi
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.test.StandardTestDispatcher
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.resetMain
import kotlinx.coroutines.test.runTest
import kotlinx.coroutines.test.setMain
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Test
import java.time.Instant

@OptIn(ExperimentalCoroutinesApi::class)
class FeedViewModelTest {

    private val testDispatcher = StandardTestDispatcher()

    @Before
    fun setup() {
        Dispatchers.setMain(testDispatcher)
    }

    @After
    fun tearDown() {
        Dispatchers.resetMain()
    }

    @Test
    fun `initial load succeeds with posts`() = runTest {
        val repo = FakeSocialRepository(posts = listOf(samplePost("p1", "hello")))
        val vm = FeedViewModel(repo)
        advanceUntilIdle()
        val state = vm.uiState.first()
        assertEquals(1, state.posts.size)
        assertEquals("p1", state.posts[0].postId)
        assertFalse(state.isLoading)
        assertNull(state.error)
    }

    @Test
    fun `initial load fails with error`() = runTest {
        val repo = FakeSocialRepository(failGetFeed = true)
        val vm = FeedViewModel(repo)
        advanceUntilIdle()
        val state = vm.uiState.first()
        assertTrue(state.posts.isEmpty())
        assertFalse(state.isLoading)
        assertNotNull(state.error)
    }

    @Test
    fun `refresh does not clear old data on failure`() = runTest {
        val repo = FakeSocialRepository(posts = listOf(samplePost("p1", "initial")))
        val vm = FeedViewModel(repo)
        advanceUntilIdle()
        assertEquals(1, vm.uiState.value.posts.size)

        repo.failGetFeed = true
        vm.refresh()
        advanceUntilIdle()
        // Old data should still be present (refresh failed but didn't clear)
        assertTrue(vm.uiState.value.posts.isNotEmpty())
        assertNotNull(vm.uiState.value.error)
    }

    @Test
    fun `like toggle optimistic then rollback on failure`() = runTest {
        val repo = FakeSocialRepository(posts = listOf(samplePost("p1", "hello")))
        val vm = FeedViewModel(repo)
        advanceUntilIdle()

        repo.failLike = true
        vm.toggleLike("p1")
        advanceUntilIdle()
        // After rollback, isLiked should be false again
        val post = vm.uiState.value.posts.find { it.postId == "p1" }
        assertEquals(false, post?.isLiked)
        assertEquals(5, post?.likeCount) // original count
    }

    @Test
    fun `append page fails only affects append state`() = runTest {
        val repo = FakeSocialRepository(
            posts = listOf(samplePost("p1")),
            nextCursor = "cursor2",
        )
        val vm = FeedViewModel(repo)
        advanceUntilIdle()
        assertEquals(1, vm.uiState.value.posts.size)

        repo.failGetFeed = true
        vm.loadMore()
        advanceUntilIdle()
        // Main list should still have 1 post
        assertEquals(1, vm.uiState.value.posts.size)
        assertNotNull(vm.uiState.value.appendError)
    }
}

// ── Fake Repository ──

class FakeSocialRepository(
    private var posts: List<PostUi> = emptyList(),
    private var nextCursor: String? = null,
    var failGetFeed: Boolean = false,
    var failLike: Boolean = false,
) : SocialRepository(
    feedRpc = error("not needed"),
    profileRpc = error("not needed"),
    tokenStore = error("not needed"),
) {
    override suspend fun getFeed(cursor: String, pageSize: Int, filter: String): Pair<List<PostUi>, String?> {
        if (failGetFeed) throw RuntimeException("simulated error")
        return if (cursor.isEmpty()) posts to nextCursor
        else posts to null
    }

    override suspend fun getPost(postId: String): PostUi =
        posts.first { it.postId == postId }

    override suspend fun likePost(postId: String): PostUi {
        if (failLike) throw RuntimeException("simulated error")
        return posts.first { it.postId == postId }.copy(isLiked = true, likeCount = 6)
    }

    override suspend fun unlikePost(postId: String): PostUi {
        if (failLike) throw RuntimeException("simulated error")
        return posts.first { it.postId == postId }.copy(isLiked = false, likeCount = 4)
    }

    override suspend fun commentOnPost(postId: String, content: String, parentCommentId: String?): CommentUi =
        CommentUi(commentId = "c1", postId = postId, authorId = "u", authorName = "U", content = content)

    override suspend fun listComments(postId: String, cursor: String, pageSize: Int): Pair<List<CommentUi>, String?> =
        emptyList<CommentUi>() to null

    override suspend fun sharePost(postId: String, comment: String): PostUi =
        posts.first { it.postId == postId }

    override suspend fun createPost(content: String, signalPair: String, signalDirection: String,
                                    signalConfidence: Int, visibility: String): PostUi =
        samplePost("new", content)

    override suspend fun listUserPosts(userId: String, cursor: String, pageSize: Int, filter: String): Pair<List<PostUi>, String?> =
        posts to null

    override suspend fun getProfile(userId: String): TraderProfileUi =
        TraderProfileUi(userId = userId, displayName = "Test")

    override suspend fun follow(userId: String): Int = 1

    override suspend fun unfollow(userId: String): Int = 0
}

private fun samplePost(id: String, content: String = "test") = PostUi(
    postId = id,
    authorId = "a1",
    authorName = "Author",
    content = content,
    postType = PostType.TEXT,
    likeCount = 5,
    commentCount = 0,
    shareCount = 0,
    isLiked = false,
    createdAt = Instant.now(),
)
