plugins {
    alias(libs.plugins.kotlin.android)
    alias(libs.plugins.android.app)
    alias(libs.plugins.hilt.plugin)
    alias(libs.plugins.ksp)
    alias(libs.plugins.compose.compiler)
    id("com.google.protobuf")
}

import java.util.Properties
import com.google.protobuf.gradle.*

android {
    namespace = "com.antclaw.alfq"
    compileSdk = 35

    defaultConfig {
        applicationId = "com.antclaw.alfq"
        minSdk = 26
        targetSdk = 35
        versionCode = 1
        versionName = "0.1.0"
        testInstrumentationRunner = "androidx.test.runner.AndroidJUnitRunner"

        buildConfigField("String", "BASE_URL", "\"https://api.alfq.org/\"")
        // 本地开发覆盖（取消注释以连接本地 antclaw-api:8080）
        // buildConfigField("String", "BASE_URL", "\"http://10.0.2.2:8080/\"")
    }

    signingConfigs {
        create("release") {
            // 签名信息从环境变量读取，不在仓库中硬编码
            // 示例：在 local.properties 中配置：
            // RELEASE_STORE_FILE=/path/to/release.jks
            // RELEASE_STORE_PASSWORD=your_password
            // RELEASE_KEY_ALIAS=your_alias
            // RELEASE_KEY_PASSWORD=your_password
            val propsFile = file("../local.properties")
            if (propsFile.exists()) {
                val props = Properties()
                props.load(propsFile.inputStream())
                val storeFilePath = props.getProperty("RELEASE_STORE_FILE")
                val storePass = props.getProperty("RELEASE_STORE_PASSWORD")
                val keyAlias = props.getProperty("RELEASE_KEY_ALIAS")
                val keyPass = props.getProperty("RELEASE_KEY_PASSWORD")
                
                if (storeFilePath != null && storePass != null && keyAlias != null && keyPass != null) {
                    storeFile = file(storeFilePath)
                    storePassword = storePass
                    this.keyAlias = keyAlias
                    this.keyPassword = keyPass
                }
            }
        }
    }

    buildTypes {
        release {
            isMinifyEnabled = true
            proguardFiles(getDefaultProguardFile("proguard-android-optimize.txt"), "proguard-rules.pro")
            // 仅在配置了签名信息时才签名
            signingConfig = if (signingConfigs["release"].storeFile?.exists() == true) {
                signingConfigs["release"]
            } else {
                null
            }
        }
    }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }

    kotlinOptions {
        jvmTarget = "17"
    }

    buildFeatures {
        compose = true
        buildConfig = true
    }

    sourceSets {
        getByName("main") {
            proto {
                srcDir("../../../proto")
            }
        }
    }

    testOptions {
        unitTests.isIncludeAndroidResources = true
    }

    lint {
        disable += "MissingTranslation"
        disable += "UnusedResources"
        disable += "IconLauncherShape"
        disable += "IconLocation"
    }
}

// ── Protobuf code generation ──
protobuf {
    protoc {
        artifact = "com.google.protobuf:protoc:4.29.3"
    }
    generateProtoTasks {
        all().forEach { task ->
            task.builtins {
                create("java") {
                    option("lite")
                }
            }
        }
    }
}

dependencies {
    // Compose
    implementation(platform(libs.compose.bom))
    implementation(libs.compose.ui)
    implementation(libs.compose.ui.tooling)
    implementation(libs.compose.material3)
    implementation(libs.compose.navigation)
    // material-icons-extended 移除（20MB+ 未用）→ 各模块按需引用具体 icon 模块

    // Lifecycle
    implementation(libs.lifecycle.runtime)
    implementation(libs.lifecycle.viewmodel)

    // Hilt
    implementation(libs.hilt.android)
    ksp(libs.hilt.compiler)
    implementation(libs.hilt.navigation)

    // Connect-RPC
    implementation(libs.connect.protobuf)
    implementation(libs.connect.okhttp)
    implementation(libs.connect.google.java.ext)
    implementation(libs.protobuf.kotlin)

    // Room
    implementation(libs.room.runtime)
    ksp(libs.room.compiler)
    implementation(libs.room.ktx)

    // Networking
    implementation(libs.okhttp.core)
    implementation(libs.okhttp.sse)

    // DataStore
    implementation(libs.datastore.preferences)

    // Jetpack Security (for token encryption)
    implementation(libs.security.crypto)

    // Coroutines
    implementation(libs.coroutines.core)
    implementation(libs.coroutines.android)

    // Core
    implementation(libs.core.ktx)
    implementation(libs.activity.compose)

    // Test
    testImplementation(libs.junit)
    testImplementation(libs.kotlinx.coroutines.test)
    testImplementation(libs.core.testing)
    testImplementation(libs.mockk)
    testImplementation(libs.robolectric)

    // Android Test (Compose UI)
    androidTestImplementation(platform(libs.compose.bom))
    androidTestImplementation(libs.test.ext.junit)
    androidTestImplementation(libs.compose.ui.test.junit4)
    androidTestImplementation(libs.compose.ui.test.manifest)
    debugImplementation(platform(libs.compose.bom))
    debugImplementation(libs.compose.ui.tooling)
}
