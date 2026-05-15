package com.antclaw.alfq.di

import com.antclaw.alfq.data.repository.SocialRepository
import com.antclaw.alfq.data.local.TokenStore
import com.antclaw.alfq.data.rpc.SocialRpc
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
    fun provideSocialRpc(client: ProtocolClientInterface): SocialRpc = SocialRpc(client)

    @Provides
    @Singleton
    fun provideSocialRepository(rpc: SocialRpc, tokenStore: TokenStore): SocialRepository = SocialRepository(rpc, tokenStore)
}
