package com.antclaw.alfq.di

import com.antclaw.alfq.data.repository.*
import com.antclaw.alfq.data.local.TokenStore
import com.antclaw.alfq.data.rpc.*
import com.antclaw.alfq.data.session.SessionExpiredNotifier
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

    // RPC clients
    @Provides @Singleton fun provideFeedRpc(c: ProtocolClientInterface): FeedRpc = FeedRpc(c)
    @Provides @Singleton fun provideProfileRpc(c: ProtocolClientInterface): ProfileRpc = ProfileRpc(c)
    @Provides @Singleton fun provideSearchRpc(c: ProtocolClientInterface): SearchRpc = SearchRpc(c)
    @Provides @Singleton fun provideTrendRpc(c: ProtocolClientInterface): TrendRpc = TrendRpc(c)
    @Provides @Singleton fun provideAlertRpc(c: ProtocolClientInterface): AlertRpc = AlertRpc(c)
    @Provides @Singleton fun provideChatRpc(c: ProtocolClientInterface): ChatRpc = ChatRpc(c)
    @Provides @Singleton fun provideUserRpc(c: ProtocolClientInterface): UserRpc = UserRpc(c)
    @Provides @Singleton fun provideSignalRpc(c: ProtocolClientInterface): SignalRpc = SignalRpc(c)
    @Provides @Singleton fun providePriceRpc(c: ProtocolClientInterface): PriceRpc = PriceRpc(c)

    // SSE
    @Provides @Singleton fun provideSseClient(mgr: SseManager): SseClient = mgr

    // Repositories
    @Provides @Singleton fun provideSocialRepository(
        feed: FeedRpc, profile: ProfileRpc, token: TokenStore,
    ): SocialRepository = SocialRepository(feed, profile, token)

    @Provides @Singleton fun provideProfileRepository(
        profile: ProfileRpc,
    ): ProfileRepository = ProfileRepository(profile)

    @Provides @Singleton fun provideAlertRepository(
        rpc: AlertRpc,
    ): AlertRepository = AlertRepository(rpc)

    @Provides @Singleton fun provideChatRepository(
        rpc: ChatRpc,
    ): ChatRepository = ChatRepository(rpc)

    @Provides @Singleton fun provideUserRepository(
        rpc: UserRpc,
    ): UserRepository = UserRepository(rpc)

    @Provides @Singleton fun provideDiscoverRepository(
        profile: ProfileRpc, search: SearchRpc, trend: TrendRpc,
    ): DiscoverRepository = DiscoverRepository(profile, search, trend)

    @Provides @Singleton fun provideSignalRepository(
        signal: SignalRpc, price: PriceRpc,
    ): SignalRepository = SignalRepository(signal, price)
}
