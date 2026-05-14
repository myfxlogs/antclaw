package com.antclaw.alfq

import android.app.Application
import com.antclaw.alfq.data.local.LocaleManager
import com.antclaw.alfq.data.local.TokenStore
import com.antclaw.alfq.data.rpc.ConnectTransportProvider
import dagger.hilt.android.HiltAndroidApp
import javax.inject.Inject

@HiltAndroidApp
class AlfQApplication : Application() {

    @Inject lateinit var tokenStore: TokenStore

    override fun onCreate() {
        super.onCreate()
        LocaleManager.init(this)
        // Initialize ConnectTransportProvider with TokenStore (was a side-effect in AppModule)
        ConnectTransportProvider.init(tokenStore)
    }
}
