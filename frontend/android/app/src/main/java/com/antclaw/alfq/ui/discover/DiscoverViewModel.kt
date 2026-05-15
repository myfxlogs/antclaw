package com.antclaw.alfq.ui.discover

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.antclaw.alfq.data.repository.DiscoverRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

data class TraderItem(val userId: String, val displayName: String, val tier: String, val followerCount: Int)
data class DiscoverUiState(val traders: List<TraderItem> = emptyList(), val loading: Boolean = false, val error: String? = null)

@HiltViewModel
class DiscoverViewModel @Inject constructor(
    private val discoverRepo: DiscoverRepository,
) : ViewModel() {
    private val _uiState = MutableStateFlow(DiscoverUiState())
    val uiState: StateFlow<DiscoverUiState> = _uiState.asStateFlow()

    init { load() }

    fun load() {
        viewModelScope.launch {
            _uiState.value = DiscoverUiState(loading = true)
            try {
                _uiState.value = DiscoverUiState(traders = discoverRepo.listFollowing().map {
                    TraderItem(it.userId, it.displayName, it.tier, it.followerCount)
                })
            } catch (e: Exception) { _uiState.value = DiscoverUiState(error = e.message) }
        }
    }
}
