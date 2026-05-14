plugins {
    alias(libs.plugins.kotlin.android) apply false
    alias(libs.plugins.android.app) apply false
    alias(libs.plugins.hilt.plugin) apply false
    alias(libs.plugins.ksp) apply false
    alias(libs.plugins.compose.compiler) apply false
    id("com.google.protobuf") version "0.9.5" apply false
}
