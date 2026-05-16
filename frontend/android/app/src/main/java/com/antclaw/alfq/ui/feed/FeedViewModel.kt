package com.antclaw.alfq.ui.feed

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.antclaw.alfq.data.repository.SocialRepository
import com.antclaw.alfq.ui.social.UiEvent
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.StateFlow
import javax.inject.Inject

enum class HomeFeedTab(val filter: String) {
    FOLLOWING("following"),
    RECOMMENDED("all"),
    SIGNALS("signals_only"),
}

@HiltViewModel
class FeedViewModel @Inject constructor(
    repository: SocialRepository,
) : ViewModel() {

    private val controller = TimelineController(viewModelScope, repository)
    val uiState: StateFlow<TimelineState> = controller.state
    val uiEvent: SharedFlow<UiEvent> = controller.uiEvent

    private var _currentTab = HomeFeedTab.RECOMMENDED
    val currentTab: HomeFeedTab get() = _currentTab

    init { load(HomeFeedTab.RECOMMENDED) }

    fun selectTab(tab: HomeFeedTab) {
        load(tab)
    }

    fun load(tab: HomeFeedTab = _currentTab) {
        _currentTab = tab
        controller.load(tab.filter)
    }

    fun refresh() = controller.refresh(_currentTab.filter)
    fun loadMore() = controller.loadMore(_currentTab.filter)
    fun toggleLike(postId: String) = controller.toggleLike(postId)
    fun sharePost(postId: String) = controller.sharePost(postId)
}
