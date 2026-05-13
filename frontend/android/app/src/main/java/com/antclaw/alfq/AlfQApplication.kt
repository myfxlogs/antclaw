package com.antclaw.alfq

import android.app.Application
import com.antclaw.alfq.data.local.LocaleManager
import dagger.hilt.android.HiltAndroidApp

@HiltAndroidApp
class AlfQApplication : Application() {
    override fun onCreate() {
        super.onCreate()
        LocaleManager.init(this)
    }
}
