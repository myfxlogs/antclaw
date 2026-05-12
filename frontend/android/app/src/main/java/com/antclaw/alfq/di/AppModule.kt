package com.antclaw.alfq.di

import android.content.Context
import androidx.room.Room
import com.antclaw.alfq.data.local.AppDatabase
import com.antclaw.alfq.data.rpc.ConnectTransportProvider
import com.connectrpc.okhttp.ConnectOkHttpClient
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.android.qualifiers.ApplicationContext
import dagger.hilt.components.SingletonComponent
import javax.inject.Singleton

@Module
@InstallIn(SingletonComponent::class)
object AppModule {

    @Provides
    @Singleton
    fun provideConnectClient(): ConnectOkHttpClient {
        return ConnectTransportProvider.create()
    }

    @Provides
    @Singleton
    fun provideAppDatabase(@ApplicationContext context: Context): AppDatabase {
        return Room.databaseBuilder(context, AppDatabase::class.java, "alfq.db")
            .fallbackToDestructiveMigration()
            .build()
    }
}
