<script lang="ts" setup>
import { onMounted, provide, ref } from 'vue'
import { CheckForUpdate, OpenRepository, SelectDirectory, Version } from '../wailsjs/go/main/App'
import RecentReposList from './components/RecentReposList.vue'
import RepoSelector from './components/RepoSelector.vue'

const repoPath = ref('')
provide('repoPath', repoPath)
const loading = ref(false)
const error = ref('')
const repoReady = ref(false)
const appVersion = ref('')
const updateURL = ref('')
const updateTag = ref('')

onMounted(async () => {
  appVersion.value = await Version()
  try {
    const info = await CheckForUpdate()
    if (info.available) {
      updateTag.value = info.tag
      updateURL.value = info.url
    }
  } catch {
    // Silently ignore update check failures
  }
})

async function onSelectRepo() {
  const path = await SelectDirectory()
  if (!path) return

  repoPath.value = path
  loading.value = true
  error.value = ''
  repoReady.value = false

  try {
    await OpenRepository(path)
    repoReady.value = true
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    loading.value = false
  }
}

async function onOpenRecent(path: string) {
  repoPath.value = path
  loading.value = true
  error.value = ''
  repoReady.value = false

  try {
    await OpenRepository(path)
    repoReady.value = true
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div id="layout">
    <header>
      <h1>Git Analytics</h1>
      <RepoSelector
        :repo-path="repoPath"
        :loading="loading"
        @select="onSelectRepo"
      />
      <nav v-if="repoReady" class="nav-tabs">
        <router-link to="/" exact-active-class="active">Activity</router-link>
        <router-link to="/hotspots" active-class="active">Hotspots</router-link>
        <router-link to="/contributors" active-class="active">Contributors</router-link>
        <router-link to="/ownership" active-class="active">Ownership</router-link>
        <router-link to="/coupling" active-class="active">Coupling</router-link>
      </nav>
    </header>
    <main>
      <div v-if="loading" class="status">
        <div class="spinner"></div>
        <span>Indexing repository...</span>
      </div>
      <div v-else-if="error" class="status error-message">
        {{ error }}
      </div>
      <div v-else-if="!repoReady" class="status welcome">
        <RecentReposList @select="onOpenRecent" />
      </div>
      <router-view v-else :key="repoPath" />
    </main>
    <footer v-if="appVersion">
      <a href="https://github.com/tbrittain/git-analytics" target="_blank" rel="noopener">Git Analytics</a>
      <span class="separator">·</span>
      <span>{{ appVersion }}</span>
      <a v-if="updateURL" :href="updateURL" target="_blank" rel="noopener" class="update-link">
        Update available: {{ updateTag }}
      </a>
    </footer>
  </div>
</template>

<style scoped>
#layout {
  display: flex;
  flex-direction: column;
  height: 100vh;
}

header {
  display: flex;
  align-items: center;
  gap: 24px;
  padding: 12px 20px;
  border-bottom: 1px solid #30363d;
  flex-shrink: 0;
}

header h1 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #f0f6fc;
  white-space: nowrap;
}

.nav-tabs {
  display: flex;
  gap: 4px;
  margin-left: auto;
}

.nav-tabs a {
  padding: 6px 14px;
  font-size: 13px;
  font-weight: 500;
  color: #c9d1d9;
  text-decoration: none;
  border: 1px solid #30363d;
  border-radius: 6px;
  background: #21262d;
  transition: background 0.15s, border-color 0.15s;
}

.nav-tabs a:hover {
  background: #30363d;
}

.nav-tabs a.active {
  background: #1f6feb;
  border-color: #1f6feb;
  color: #ffffff;
}

main {
  flex: 1;
  overflow: auto;
  padding: 20px;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

footer {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 20px;
  border-top: 1px solid #30363d;
  color: #8b949e;
  font-size: 12px;
  flex-shrink: 0;
}

footer a {
  color: #8b949e;
  text-decoration: none;
}

footer a:hover {
  color: #c9d1d9;
  text-decoration: underline;
}

.separator {
  color: #30363d;
}

.update-link {
  margin-left: auto;
  color: #58a6ff;
}

.update-link:hover {
  color: #79c0ff;
}

.status {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  height: 100%;
  color: #8b949e;
  font-size: 15px;
}

.error-message {
  color: #f85149;
}

.welcome {
  color: #8b949e;
}

.spinner {
  width: 20px;
  height: 20px;
  border: 2px solid #30363d;
  border-top-color: #58a6ff;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>
