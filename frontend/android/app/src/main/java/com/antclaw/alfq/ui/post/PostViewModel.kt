package com.antclaw.alfq.ui.post

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import antclaw.v1.AlfqFeed
import com.antclaw.alfq.data.rpc.ConnectTransportProvider
import com.connectrpc.MethodSpec
import com.connectrpc.StreamType
import com.connectrpc.getOrThrow
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class PostViewModel @Inject constructor() : ViewModel() {
    private val _posted = MutableStateFlow(false)
    val posted: StateFlow<Boolean> = _posted.asStateFlow()

    fun post(content: String, signalPair: String, signalDirection: String, signalConfidence: Int, visibility: String) {
        viewModelScope.launch {
            try {
                val req = AlfqFeed.CreatePostRequest.newBuilder()
                    .setContent(content)
                    .setPostType(if (signalPair.isBlank()) "post" else "signal")
                    .setSignalPair(if (signalPair.isBlank()) "" else signalPair)
                    .setSignalDirection(signalDirection)
                    .setSignalConfidence(signalConfidence)
                    .setVisibility(visibility).build()
                val spec = MethodSpec("antclaw.v1.FeedService/CreatePost",
                    AlfqFeed.CreatePostRequest::class, AlfqFeed.Post::class, StreamType.UNARY)
                ConnectTransportProvider.createProtocolClient().unary(req, emptyMap(), spec).getOrThrow()
                _posted.value = true
            } catch (_: Exception) { _posted.value = false }
        }
    }
}
