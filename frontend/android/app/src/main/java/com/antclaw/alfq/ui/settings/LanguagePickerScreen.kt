package com.antclaw.alfq.ui.settings

import android.app.Activity
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Check
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import com.antclaw.alfq.R
import com.antclaw.alfq.data.local.LocaleManager

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun LanguagePickerScreen(onBack: () -> Unit) {
    val context = LocalContext.current
    val currentLang = remember { LocaleManager.getSelectedLanguage(context) }
    var selected by remember { mutableStateOf(currentLang) }
    val languages = remember { LocaleManager.getAvailableLanguages() }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(stringResource(R.string.settings_language_title), color = MaterialTheme.colorScheme.onSurface) },
                navigationIcon = { TextButton(onClick = onBack) { Text(stringResource(R.string.common_back)) } },
                colors = TopAppBarDefaults.topAppBarColors(containerColor = MaterialTheme.colorScheme.background),
            )
        },
        containerColor = MaterialTheme.colorScheme.background,
    ) { padding ->
        LazyColumn(modifier = Modifier.fillMaxSize().padding(padding).padding(horizontal = 16.dp),
            verticalArrangement = Arrangement.spacedBy(4.dp)) {
            items(languages) { (code, name) ->
                Surface(
                    modifier = Modifier.fillMaxWidth().clickable {
                        selected = code
                        LocaleManager.setSelectedLanguage(context, code)
                        LocaleManager.applyLocale(context, code)
                        (context as? Activity)?.recreate()
                    },
                    color = if (selected == code) MaterialTheme.colorScheme.primaryContainer
                           else MaterialTheme.colorScheme.surface,
                    shape = MaterialTheme.shapes.small,
                ) {
                    Row(
                        modifier = Modifier.fillMaxWidth().padding(horizontal = 16.dp, vertical = 14.dp),
                        horizontalArrangement = Arrangement.SpaceBetween,
                        verticalAlignment = Alignment.CenterVertically,
                    ) {
                        Text(name, style = MaterialTheme.typography.bodyLarge,
                            fontWeight = if (selected == code) FontWeight.Bold else FontWeight.Normal)
                        if (selected == code) {
                            Icon(Icons.Default.Check, contentDescription = stringResource(R.string.common_ok),
                                tint = MaterialTheme.colorScheme.primary, modifier = Modifier.size(20.dp))
                        }
                    }
                }
            }
        }
    }
}
