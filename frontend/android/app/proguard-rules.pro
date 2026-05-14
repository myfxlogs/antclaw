# ProGuard / R8 rules for AlfQ Android app
# Minimal rules — main config comes from proguard-android-optimize.txt

# Connect-RPC / Protobuf
# Proto-generated Java classes are in antclaw.v1 package (protobuf-lite, not gen/kotlin)
-keep class antclaw.v1.** { *; }
-keep class com.connectrpc.** { *; }
-keep class com.google.protobuf.** { *; }
# Prevent R8 from stripping proto builder/setter/getter methods
-keepclassmembers class * extends com.google.protobuf.GeneratedMessageLite {
  <fields>;
  <methods>;
}

# Hilt
-keep class dagger.hilt.** { *; }
-keep class javax.inject.** { *; }

# OkHttp SSE
-dontwarn okhttp3.internal.platform.**
-keep class okhttp3.** { *; }

# Room
-keep class * extends androidx.room.RoomDatabase
-keep @androidx.room.Entity class *
-dontwarn androidx.room.paging.**

# DataStore
-keep class androidx.datastore.** { *; }

# Coroutines
-keepnames class kotlinx.coroutines.internal.MainDispatcherFactory {}
-keepnames class kotlinx.coroutines.CoroutineExceptionHandler {}
