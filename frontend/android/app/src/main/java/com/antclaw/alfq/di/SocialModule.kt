package com.antclaw.alfq.di

import com.antclaw.alfq.data.repository.SocialRepository
import com.antclaw.alfq.data.local.TokenStore
import com.antclaw.alfq.data.rpc.FeedRpc
import com.antclaw.alfq.data.rpc.ProfileRpc
import com.antclaw.alfq.data.rpc.SearchRpc
import com.antclaw.alfq.data.rpc.TrendRpc
import com.antclaw.alfq.data.sse.SseClient
import com.antclaw.alfq.data.sse.SseManager
import com.connectrpc.ProtocolClientInterface
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.components.SingletonComponent
import javax.inject.Singleton

@Module
@InstallIn(SingletonComponent::class)
object SocialModule {

    @Provides
    @Singleton
    fun provideFeedRpc(client: ProtocolClientInterface): FeedRpc = FeedRpc(client)

    @Provides
    @Singleton
    fun provideProfileRpc(client: ProtocolClientInterface): ProfileRpc = ProfileRpc(client)

    @Provides
    @Singleton
    fun provideSearchRpc(client: ProtocolClientInterface): SearchRpc = SearchRpc(client)

    @Provides
    @Singleton
    fun provideTrendRpc(client: ProtocolClientInterface): TrendRpc = TrendRpc(client)

    @Provides
    @Singleton
    fun provideSseClient(sseManager: SseManager): SseClient = sseManager

    @Provides
    @Singleton
    fun provideSocialRepository(
        feedRpc: FeedRpc,
        profileRpc: ProfileRpc,
        tokenStore: TokenStore,
    ): SocialRepository = SocialRepository(feedRpc, profileRpc, tokenStore)
}
