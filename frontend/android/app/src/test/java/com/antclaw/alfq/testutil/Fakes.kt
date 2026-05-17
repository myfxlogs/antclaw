package com.antclaw.alfq.testutil

import com.antclaw.alfq.data.local.TokenStoreApi
import com.antclaw.alfq.data.repository.DeviceReportApi
import com.antclaw.alfq.data.repository.SocialRepository
import com.antclaw.alfq.data.rpc.FeedRpc
import com.antclaw.alfq.data.rpc.ProfileRpc
import com.antclaw.alfq.data.rpc.RefreshTokenCoordinator
import com.antclaw.alfq.data.rpc.TokenManager
import com.antclaw.alfq.data.rpc.TokenSnapshot
import com.antclaw.alfq.data.sse.SseClient
import com.antclaw.alfq.ui.social.*
import com.connectrpc.ProtocolClientInterface
import io.mockk.mockk
import java.time.Instant

// ══════════════════════════════════════════════════
// 共享 Test Double — 消除跨测试文件的重复 fake
// ══════════════════════════════════════════════════

/**
 * 内存 Token 存储 fake，实现 [TokenStoreApi]。
 * 替换原来的 FakeTok（SessionViewModelTest）和 SimTokenStore（ConnectTransportProviderRefreshTest）。
 */
class FakeTokenStore(
    private var accessToken: String? = null,
    private var refreshToken: String? = null,
    private var userId: String? = null,
) : TokenStoreApi {
    override suspend fun getAccessToken() = accessToken
    override suspend fun getRefreshToken() = refreshToken
    override suspend fun getUserId() = userId
    override suspend fun saveAccessToken(token: String) { accessToken = token }
    override suspend fun saveRefreshToken(token: String) { refreshToken = token }
    override suspend fun saveUserId(id: String) { userId = id }
    override suspend fun saveTokens(at: String, rt: String, uid: String) {
        accessToken = at; refreshToken = rt; userId = uid
    }
    override suspend fun clearTokens() { accessToken = null; refreshToken = null }
    override suspend fun clearUserId() { userId = null }
}

/**
 * 社交仓库 fake，扩展 [SocialRepository]。
 * 替换原来的 FeedFakeRepo（FeedViewModelTest）和 PostFakeRepo（PostViewModelTest）。
 *
 * 覆盖 Feed 首屏/分页/创建/点赞/分享的失败注入和计数追踪。
 */
class FakeSocialRepo(
    var posts: List<PostUi> = emptyList(),
    var cursor: String? = null,
    var failFirstPage: Boolean = false,
    var failAppend: Boolean = false,
    var failCreate: Boolean = false,
    var failLike: Boolean = false,
    var failShare: Boolean = false,
) : SocialRepository(
    FeedRpc(mockk<ProtocolClientInterface>(relaxed = true)),
    ProfileRpc(mockk<ProtocolClientInterface>(relaxed = true)),
    mockk<TokenStoreApi>(relaxed = true),
) {
    var submitCount = 0
    var lastContent: String = ""
    var lastSignalPair: String = ""
    var lastVisibility: String = ""

    override suspend fun getFeed(c: String, ps: Int, f: String): Pair<List<PostUi>, String?> {
        if (c.isEmpty()) {
            if (failFirstPage) throw RuntimeException("first page error")
            return posts to cursor
        }
        if (failAppend) throw RuntimeException("append error")
        return posts to cursor
    }

    override suspend fun getPost(id: String) =
        posts.firstOrNull { it.postId == id }
            ?: PostUi(postId = id, authorId = "a", authorName = "A", content = "x", postType = PostType.TEXT, createdAt = Instant.now())

    override suspend fun likePost(id: String): PostUi {
        if (failLike) throw RuntimeException("like error")
        return postUi("x", likeCount = 6, isLiked = true)
    }

    override suspend fun unlikePost(id: String): PostUi {
        if (failLike) throw RuntimeException("unlike error")
        return postUi("x", likeCount = 4, isLiked = false)
    }

    override suspend fun sharePost(id: String, c: String): PostUi {
        if (failShare) throw RuntimeException("share error")
        return postUi("x")
    }

    override suspend fun createPost(a: String, b: String, c: String, d: Int, e: String): PostUi {
        submitCount++
        lastContent = a; lastSignalPair = b; lastVisibility = e
        if (failCreate) throw RuntimeException("create post error")
        return PostUi(
            postId = "new_${submitCount}", authorId = "u", authorName = "U",
            content = a, postType = PostType.TEXT, createdAt = Instant.now(),
        )
    }

    override suspend fun commentOnPost(a: String, b: String, c: String?) =
        CommentUi("c1", a, "u", "n", b)

    override suspend fun listComments(a: String, b: String, c: Int) =
        emptyList<CommentUi>() to null

    override suspend fun listUserPosts(a: String, b: String, c: Int, d: String) =
        posts to null

    override suspend fun getProfile(a: String) = TraderProfileUi(a, "T")
    override suspend fun follow(a: String) = 1
    override suspend fun unfollow(a: String) = 0
}

/**
 * SSE 客户端 fake。
 * 替换原来内联在 SessionViewModelTest 中的 FakeSse。
 */
class FakeSse : SseClient {
    var connected = false; private set
    override fun connect() { connected = true }
    override fun disconnect() { connected = false }
    override fun reconnect() { disconnect(); connect() }
    override fun destroy() { connected = false }
}

/**
 * 设备上报 fake。
 * 替换原来内联在 SessionViewModelTest 中的 FakeDeviceReportApi。
 */
class FakeDeviceReportApi(val fail: Boolean = false) : DeviceReportApi {
    var reported = false; private set
    override suspend fun reportDeviceInfo() {
        reported = true
        if (fail) throw RuntimeException("simulated device report failure")
    }
}

/**
 * 创建测试用 [TokenManager]，绑定 [FakeTokenStore]。
 * 返回 Pair(FakeTokenStore, TokenManager)，调用方可直接操作 tokenStore 和 tokenManager。
 */
fun fakeTokenManager(
    accessToken: String? = null,
    refreshToken: String? = null,
    userId: String? = null,
): Pair<FakeTokenStore, TokenManager> {
    val store = FakeTokenStore(accessToken, refreshToken, userId)
    val snapshot = TokenSnapshot()
    val coordinator = RefreshTokenCoordinator(store, snapshot)
    val tm = TokenManager(snapshot, coordinator)
    return store to tm
}

/** 创建用于测试的 [PostUi]。 */
fun postUi(
    id: String,
    content: String = "t",
    likeCount: Int = 5,
    shareCount: Int = 0,
    isLiked: Boolean = false,
) = PostUi(
    postId = id, authorId = "a", authorName = "A", content = content,
    postType = PostType.TEXT, likeCount = likeCount, shareCount = shareCount,
    isLiked = isLiked, createdAt = Instant.now(),
)
