package com.antclaw.alfq.data.local

import android.content.Context
import android.content.SharedPreferences
import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import androidx.security.crypto.EncryptedSharedPreferences
import androidx.security.crypto.MasterKey
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.map
import javax.inject.Inject
import javax.inject.Singleton

private val Context.dataStore: DataStore<Preferences> by preferencesDataStore(name = "alfq_tokens")

@Singleton
open class TokenStore @Inject constructor(
    @ApplicationContext private val context: Context
) {
    companion object {
        private val KEY_ACCESS_TOKEN = stringPreferencesKey("access_token")
        private val KEY_USER_ID = stringPreferencesKey("user_id")
        private const val ENCRYPTED_PREFS_FILE_NAME = "alfq_encrypted_tokens"
        private const val KEY_REFRESH_TOKEN = "refresh_token"
    }

    private val encryptedPrefs: SharedPreferences by lazy {
        val masterKey = MasterKey.Builder(context)
            .setKeyScheme(MasterKey.KeyScheme.AES256_GCM)
            .build()

        EncryptedSharedPreferences.create(
            context,
            ENCRYPTED_PREFS_FILE_NAME,
            masterKey,
            EncryptedSharedPreferences.PrefKeyEncryptionScheme.AES256_SIV,
            EncryptedSharedPreferences.PrefValueEncryptionScheme.AES256_GCM
        )
    }

    open suspend fun saveTokens(accessToken: String, refreshToken: String, userId: String = "") {
        context.dataStore.edit { prefs ->
            prefs[KEY_ACCESS_TOKEN] = accessToken
            if (userId.isNotBlank()) prefs[KEY_USER_ID] = userId
        }
        encryptedPrefs.edit().putString(KEY_REFRESH_TOKEN, refreshToken).apply()
    }

    open suspend fun saveAccessToken(token: String) {
        context.dataStore.edit { prefs ->
            prefs[KEY_ACCESS_TOKEN] = token
        }
    }

    open suspend fun getAccessToken(): String? {
        return context.dataStore.data.map { it[KEY_ACCESS_TOKEN] }.first()
    }

    open suspend fun getRefreshToken(): String? {
        return encryptedPrefs.getString(KEY_REFRESH_TOKEN, null)
    }

    open suspend fun getUserId(): String? {
        return context.dataStore.data.map { it[KEY_USER_ID] }.first()
    }

    open suspend fun clearTokens() {
        context.dataStore.edit { it.clear() }
        encryptedPrefs.edit().clear().apply()
    }

    open suspend fun saveRefreshToken(token: String) {
        encryptedPrefs.edit().putString(KEY_REFRESH_TOKEN, token).apply()
    }

    open suspend fun saveUserId(userId: String) {
        context.dataStore.edit { prefs ->
            prefs[KEY_USER_ID] = userId
        }
    }

    open suspend fun clearUserId() {
        context.dataStore.edit { prefs ->
            prefs.remove(KEY_USER_ID)
        }
    }
}
