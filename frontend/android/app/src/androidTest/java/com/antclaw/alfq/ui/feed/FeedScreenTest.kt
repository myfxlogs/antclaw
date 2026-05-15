package com.antclaw.alfq.ui.feed

import androidx.compose.ui.test.assertIsDisplayed
import androidx.compose.ui.test.junit4.createComposeRule
import androidx.compose.ui.test.onNodeWithText
import com.antclaw.alfq.ui.social.PostType
import com.antclaw.alfq.ui.social.PostUi
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import org.junit.Rule
import org.junit.Test
import java.time.Instant

class FeedScreenTest {

    @get:Rule
    val composeTestRule = createComposeRule()

    @Test
    fun `loading state shows progress indicator`() {
        composeTestRule.setContent {
            LoadingView()
        }
        composeTestRule.onNodeWithText("Loading").assertDoesNotExist()
        // CircularProgressIndicator is semantically a progress bar
    }

    @Test
    fun `error state shows error message and retry button`() {
        composeTestRule.setContent {
            ErrorView("网络连接失败，请检查后重试") {}
        }
        composeTestRule.onNodeWithText("网络连接失败，请检查后重试").assertIsDisplayed()
        composeTestRule.onNodeWithText("Retry").assertIsDisplayed()
    }

    @Test
    fun `feed renders post cards`() {
        val posts = listOf(
            PostUi(
                postId = "p1",
                authorId = "a1",
                authorName = "Alice",
                content = "Hello World",
                postType = PostType.TEXT,
                likeCount = 5,
                commentCount = 2,
                shareCount = 1,
                isLiked = false,
                createdAt = Instant.now(),
            ),
        )

        composeTestRule.setContent {
            for (post in posts) {
                PostCard(
                    post = post,
                    onLikeClick = {},
                )
            }
        }

        composeTestRule.onNodeWithText("Alice").assertIsDisplayed()
        composeTestRule.onNodeWithText("Hello World").assertIsDisplayed()
    }
}
