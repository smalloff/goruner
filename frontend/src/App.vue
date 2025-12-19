<template>
  <el-container class="app-container">
    <el-header height="40px" class="header-tabs">
      <el-tabs v-model="activeTab">
        <el-tab-pane :label="t('console')" name="console" />
        <el-tab-pane :label="t('settings')" name="settings" />
      </el-tabs>
    </el-header>

    <el-main v-if="activeTab === 'console'" class="console-view">
      <div class="console-wrapper">
        <div class="console-output" ref="consoleRef">
          <div v-for="(line, index) in formattedOutput" :key="index" :class="line.type" class="console-line">
            <span>{{ line.text }}</span>
            <el-button 
              v-if="line.text.includes(t('errorsLabel'))" 
              class="inline-copy-btn"
              type="danger" 
              size="small" 
              circle
              @click.stop="copyErrors"
              :title="t('copyTip')"
            >
              <template #icon>
                <svg viewBox="0 0 1024 1024" width="15" height="15">
                  <path fill="currentColor" d="M768 832a128 128 0 0 1-128 128H192A128 128 0 0 1 64 832V384a128 128 0 0 1 128-128h64v-64a128 128 0 0 1 128-128h448a128 128 0 0 1 128 128v448a128 128 0 0 1-128 128h-64v64zM384 128a64 64 0 0 0-64 64v448a64 64 0 0 0 64 64h448a64 64 0 0 0 64-64V192a64 64 0 0 0-64-64H384zm-192 192a64 64 0 0 0-64 64v448a64 64 0 0 0 64 64h448a64 64 0 0 0 64-64v-64H384a128 128 0 0 1-128-128V320h-64z"></path>
                </svg>
              </template>
            </el-button>
          </div>
        </div>
      </div>
      <div class="controls-footer">
        <div class="controls">
          <el-button class="icon-btn restart" @click="runTests" :loading="loading" :title="t('restartTip')">
            <template #icon>
              <svg viewBox="0 0 1024 1024" width="22" height="22">
                <path fill="currentColor" d="M224 416h192a32 32 0 0 0 0-64H306.432A352.064 352.064 0 0 1 896 512c0 194.4-157.6 352-352 352a351.488 351.488 0 0 1-295.552-160.896 32 32 0 0 0-54.656 33.152A415.488 415.488 0 0 0 544 928c229.76 0 416-186.24 416-416S773.76 96 544 96c-184.256 0-341.312 119.744-397.632 286.336V224a32 32 0 0 0-64 0v192a32 32 0 0 0 32 32h109.632z" style="stroke-width: 40px; stroke: currentColor;"></path>
              </svg>
            </template>
          </el-button>
          
          <el-button class="icon-btn toggle" :class="{ 'is-paused': isPaused }" @click="togglePause" :title="isPaused ? t('resumeTip') : t('pauseTip')">
            <template #icon>
              <svg v-if="!isPaused" viewBox="0 0 1024 1024" width="22" height="22">
                <rect x="256" y="160" width="160" height="704" rx="40" fill="currentColor" />
                <rect x="608" y="160" width="160" height="704" rx="40" fill="currentColor" />
              </svg>
              <svg v-else viewBox="0 0 1024 1024" width="22" height="22">
                <path fill="currentColor" d="M256 160l576 352-576 352z" style="stroke-width: 40px; stroke: currentColor; stroke-linejoin: round;"></path>
              </svg>
            </template>
          </el-button>

          <el-button class="icon-btn lang-btn" @click="toggleLang" :title="t('langTip')">
            <span class="lang-label">{{ (config.lang || 'ru').toUpperCase() }}</span>
          </el-button>
        </div>
      </div>
    </el-main>

    <el-main v-if="activeTab === 'settings'" class="settings-view">
      <el-form :model="config" label-position="top" class="settings-form">
        <el-form-item :label="t('exclusionsLabel')">
          <el-input type="textarea" v-model="exclusionText" rows="4" />
        </el-form-item>
        <el-form-item>
          <el-checkbox v-model="config.show_passed">{{ t('showPassed') }}</el-checkbox>
        </el-form-item>
        <el-form-item>
          <el-checkbox v-model="config.auto_watch">{{ t('autoWatch') }}</el-checkbox>
        </el-form-item>
        <el-form-item>
          <el-checkbox v-model="config.show_notifications">{{ t('showNotifications') }}</el-checkbox>
        </el-form-item>
        <el-form-item v-if="config.show_notifications" style="margin-left: 20px;">
          <el-checkbox v-model="config.notify_only_on_failure">{{ t('notifyOnlyOnFailure') }}</el-checkbox>
        </el-form-item>
        <el-form-item>
          <el-checkbox v-model="config.always_on_top">{{ t('alwaysOnTop') }}</el-checkbox>
        </el-form-item>
        <el-form-item>
          <el-checkbox v-model="config.auto_copy_errors">{{ t('autoCopyErrors') }}</el-checkbox>
        </el-form-item>
        <el-button type="success" @click="saveSettings">{{ t('saveBtn') }}</el-button>
      </el-form>
    </el-main>
  </el-container>
</template>

<script setup>
import { ref, onMounted, computed, watch, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import { EventsOn, ClipboardSetText } from '../wailsjs/runtime/runtime'
import { RunTests, TogglePause, GetConfig, SaveConfig } from '../wailsjs/go/main/App'

const activeTab = ref('console')
const output = ref('')
const loading = ref(false)
const isPaused = ref(false)
const config = ref({
  exclusions: [],
  show_passed: true,
  auto_watch: true,
  show_notifications: true,
  notify_only_on_failure: false,
  always_on_top: false,
  minimize_to_tray: false,
  auto_copy_errors: false,
  lang: 'ru'
})
const exclusionText = ref('')
const consoleRef = ref(null)

const i18n = {
  ru: {
    console: 'Консоль', settings: 'Настройки', errorsLabel: 'Ошибки в тестах:', successLabel: 'Успешные тесты:',
    copyTip: 'Скопировать ошибки', restartTip: 'Перезапустить тесты', resumeTip: 'Возобновить', pauseTip: 'Пауза',
    langTip: 'Сменить язык', exclusionsLabel: 'Исключения (по одному в строке)', showPassed: 'Показывать успешные тесты',
    autoWatch: 'Автоматически перезапускать тесты', showNotifications: 'Показывать уведомления', notifyOnlyOnFailure: 'Только при ошибках', 
    alwaysOnTop: 'Поверх всех окон', autoCopyErrors: 'Автокопирование ошибок', saveBtn: 'Сохранить настройки', copied: 'Ошибки скопированы в буфер',
      running: 'Запуск тестов...', waiting: 'Ожидание изменений...'
    },
    en: {
      console: 'Console', settings: 'Settings', errorsLabel: 'Errors in tests:', successLabel: 'Successful tests:',
      copyTip: 'Copy errors', restartTip: 'Restart tests', resumeTip: 'Resume', pauseTip: 'Pause',
      langTip: 'Change language', exclusionsLabel: 'Exclusions (one per line)', showPassed: 'Show successful tests',
      autoWatch: 'Automatically restart tests', showNotifications: 'Show notifications', notifyOnlyOnFailure: 'Only on failure', 
      alwaysOnTop: 'Always on Top', autoCopyErrors: 'Auto-copy Errors', saveBtn: 'Save Settings', copied: 'Errors copied to clipboard',
      running: 'Running tests...', waiting: 'Waiting for changes...'
    }
};

const t = (key) => i18n[config.value.lang || 'en'][key];

const toggleLang = async () => {
  const isWaiting = output.value === t('waiting');
  config.value.lang = config.value.lang === 'ru' ? 'en' : 'ru';
  await SaveConfig(config.value);
  if (isWaiting) {
    output.value = t('waiting');
  } else if (output.value && !output.value.includes('Waiting') && !output.value.includes('Ожидание')) {
    runTests();
  }
};

const formattedOutput = computed(() => {
  let inErrorSection = false;
  const errMarker = t('errorsLabel');
  const successMarker = t('successLabel');

  return output.value.split('\n').map(line => {
    if (line.includes(errMarker)) inErrorSection = true;
    if (line.includes(successMarker)) inErrorSection = false;
  
    const trimmed = line.trim();
    let type = 'default';
    
    const isError = trimmed.startsWith('#') || 
                    line.includes('FAIL') || 
                    line.includes('.go:') || 
                    line.toLowerCase().includes('error:') || 
                    line.includes('panic:') ||
                    (inErrorSection && trimmed && !line.includes(errMarker));
  
    if (isError) {
      type = 'error-text';
    } else if (line.includes('PASS') || line.includes('✅') || trimmed.startsWith('ok')) {
      type = 'success-text';
    }
    return { text: line, type };
  });
});

const copyErrors = async () => {
  const errorLines = formattedOutput.value
    .filter(l => l.type === 'error-text' && !l.text.includes(t('errorsLabel')))
    .map(l => l.text);
  
  if (errorLines.length > 0) {
    const success = await ClipboardSetText(errorLines.join('\n'));
    if (success) {
      ElMessage({ message: t('copied'), type: 'success', offset: 50 });
    }
  }
};

const scrollToBottom = async () => {
  await nextTick();
  if (consoleRef.value) {
    consoleRef.value.scrollTop = consoleRef.value.scrollHeight;
  }
};

watch(output, () => scrollToBottom());

const runTests = async () => {
  loading.value = true
  output.value = t('running') + "\n"
  output.value = await RunTests()
  loading.value = false
}

const togglePause = async () => {
  isPaused.value = await TogglePause()
}

const saveSettings = async () => {
  config.value.exclusions = exclusionText.value.split('\n').filter(s => s.trim())
  await SaveConfig(config.value)
}

onMounted(async () => {
  try {
    config.value = await GetConfig()
    output.value = t('waiting')
    exclusionText.value = (config.value.exclusions || []).join('\n')
    
    EventsOn('trigger_test', (msg) => {
      if (!loading.value) {
        output.value = `[${new Date().toLocaleTimeString()}] ${msg}...\n` + output.value
        runTests()
      }
    })
  } catch (err) {
    output.value = "Error: " + err
  }
})
</script>

<style>
body {
  margin: 0;
  background: #1e1e1e;
  color: #d4d4d4;
  font-family: 'Consolas', monospace;
  overflow: hidden;
}
.app-container {
  height: 100vh;
  display: flex;
  flex-direction: column;
}
.console-view {
  display: flex;
  flex-direction: column;
  padding: 0 !important;
  flex: 1;
  overflow: hidden;
}
.console-wrapper {
  position: relative;
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  padding: 10px 10px 0 10px;
}
.console-output {
  flex: 1;
  background: #000;
  color: #d4d4d4;
  padding: 15px;
  overflow-y: auto;
  border: 1px solid #333;
  white-space: pre-wrap;
  font-size: 13px;
  line-height: 1.4;
  scroll-behavior: smooth;
}
.console-line {
  display: flex;
  align-items: center;
  gap: 10px;
  min-height: 20px;
}
.inline-copy-btn {
  width: 28px !important;
  height: 28px !important;
  min-height: 28px !important;
  padding: 0 !important;
  margin-left: 6px;
  background: rgba(255, 85, 85, 0.2) !important;
  border: 1px solid rgba(255, 85, 85, 0.6) !important;
  color: #ff5555 !important;
  display: inline-flex !important;
  align-items: center;
  justify-content: center;
  transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1) !important;
  cursor: pointer;
  box-shadow: 0 0 8px rgba(255, 85, 85, 0.4) !important;
}
.inline-copy-btn:hover {
  background: #ff5555 !important;
  border-color: #ff5555 !important;
  transform: scale(1.15);
  color: #000 !important;
  box-shadow: 0 0 18px rgba(255, 85, 85, 1) !important;
}
.error-text {
  color: #ff5555 !important;
  font-weight: bold;
}
.success-text {
  color: #50fa7b !important;
}
.default {
  color: #888;
}
.controls-footer {
  padding: 8px 12px;
  background: #1e1e1e;
  border-top: 1px solid #333;
  display: flex;
  align-items: center;
}
.controls {
  display: flex;
  gap: 2px;
}
.icon-btn {
  width: 36px !important;
  height: 36px !important;
  padding: 0 !important;
  border-radius: 6px !important;
  background: #2d2d2d !important;
  border: 1px solid #444 !important;
  color: #aaa !important;
  transition: all 0.2s ease;
}
.icon-btn:hover {
  color: #fff !important;
  border-color: #666 !important;
  background: #3d3d3d !important;
}
.icon-btn.restart:hover {
  color: #409eff !important;
  box-shadow: 0 0 10px rgba(64, 158, 255, 0.3);
}
.icon-btn.toggle {
  color: #e6a23c !important;
}
.icon-btn.toggle.is-paused {
  color: #67c23a !important;
}
.icon-btn.toggle:hover {
  box-shadow: 0 0 10px currentColor;
}
.lang-btn {
  border-color: #555 !important;
}
.lang-label {
  font-size: 11px;
  font-weight: bold;
}
.header-tabs {
  background: #252526;
  border-bottom: 1px solid #333;
}
.el-tabs__item {
  color: #888 !important;
}
.el-tabs__item.is-active {
  color: #fff !important;
}
.settings-view {
  padding: 20px !important;
  background: #1e1e1e;
}
.settings-form {
  max-width: 600px;
}
.settings-form .el-form-item {
  margin-bottom: 2px;
}
.settings-form .el-form-item:first-child {
  margin-bottom: 22px;
}
.settings-form .el-button {
  margin-top: 12px;
}
</style>