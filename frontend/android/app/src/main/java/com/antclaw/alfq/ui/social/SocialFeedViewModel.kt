package com.antclaw.alfq.ui.social

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.antclaw.alfq.data.repository.SocialRepository
import com.antclaw.alfq.ui.feed.AsyncPhase
import com.antclaw.alfq.ui.feed.TimelineController
import com.antclaw.alfq.ui.feed.TimelineState
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.StateFlow
import javax.inject.Inject

@HiltViewModel
class SocialFeedViewModel @Inject constructor(
    repository: SocialRepository,
) : ViewModel() {

    private val controller = TimelineController(viewModelScope, repository)
    val state: StateFlow<TimelineState> = controller.state
    val uiEvent: SharedFlow<UiEvent> = controller.uiEvent

    private var _currentTab = FeedTab.FOLLOWING
    val currentTab: FeedTab get() = _currentTab

    init { loadFeed(FeedTab.FOLLOWING) }

    fun loadFeed(tab: FeedTab = _currentTab) {
        _currentTab = tab
        controller.load(tab.filter)
    }

    fun refresh() = controller.refresh(_currentTab.filter)
    fun loadMore() = controller.loadMore(_currentTab.filter)
    fun toggleLike(postId: String) = controller.toggleLike(postId)
    fun sharePost(postId: String) = controller.sharePost(postId)
}
