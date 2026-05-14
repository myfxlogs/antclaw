plugins {
    alias(libs.plugins.kotlin.android)
    alias(libs.plugins.android.app)
    alias(libs.plugins.hilt.plugin)
    alias(libs.plugins.ksp)
    alias(libs.plugins.compose.compiler)
    id("com.google.protobuf")
}

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

    buildTypes {
        release {
            isMinifyEnabled = true
            proguardFiles(getDefaultProguardFile("proguard-android-optimize.txt"), "proguard-rules.pro")
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
    implementation("androidx.compose.material:material-icons-extended")

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

    // Image
    implementation(libs.coil)

    // Coroutines
    implementation(libs.coroutines.core)
    implementation(libs.coroutines.android)

    // Core
    implementation(libs.core.ktx)
    implementation(libs.activity.compose)
}
