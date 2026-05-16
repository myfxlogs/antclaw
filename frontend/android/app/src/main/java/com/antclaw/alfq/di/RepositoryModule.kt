package com.antclaw.alfq.di

import com.antclaw.alfq.data.repository.DeviceReportApi
import com.antclaw.alfq.data.repository.DeviceRepository
import dagger.Binds
import dagger.Module
import dagger.hilt.InstallIn
import dagger.hilt.components.SingletonComponent

@Module
@InstallIn(SingletonComponent::class)
abstract class RepositoryModule {

    @Binds
    abstract fun bindDeviceReportApi(impl: DeviceRepository): DeviceReportApi
}
