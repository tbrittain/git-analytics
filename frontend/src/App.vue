<script lang="ts" setup>
import { ref } from 'vue'
import { SelectDirectory, OpenRepository } from '../wailsjs/go/main/App'
import RepoSelector from './components/RepoSelector.vue'
import CommitHeatmap from './components/CommitHeatmap.vue'

const repoPath = ref('')
const loading = ref(false)
const error = ref('')
const repoReady = ref(false)

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
  } catch (e: any) {
    error.value = e?.message || String(e)
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
        Open a Git repository to get started.
      </div>
      <CommitHeatmap v-else :key="repoPath" />
    </main>
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

main {
  flex: 1;
  overflow: auto;
  padding: 20px;
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
