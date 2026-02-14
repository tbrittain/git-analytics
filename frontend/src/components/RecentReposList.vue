<script lang="ts" setup>
import { onMounted, ref } from 'vue'
import { RecentRepos, RemoveRecentRepo } from '../../wailsjs/go/main/App'
import type { config } from '../../wailsjs/go/models'

const emit = defineEmits<{
  select: [path: string]
}>()

const repos = ref<config.RecentRepo[]>([])

onMounted(async () => {
  try {
    repos.value = (await RecentRepos()) ?? []
  } catch {
    repos.value = []
  }
})

async function remove(path: string) {
  try {
    await RemoveRecentRepo(path)
    repos.value = repos.value.filter((r) => r.path !== path)
  } catch {
    // ignore
  }
}
</script>

<template>
  <div class="recent-repos">
    <template v-if="repos.length">
      <p class="prompt">Open a Git repository to get started, or pick a recent one:</p>
      <ul class="repo-list">
        <li v-for="repo in repos" :key="repo.path" class="repo-item" @click="emit('select', repo.path)">
          <div class="repo-info">
            <span class="repo-name">{{ repo.name }}</span>
            <span class="repo-path">{{ repo.path }}</span>
          </div>
          <button class="remove-btn" title="Remove from recents" @click.stop="remove(repo.path)">
            &times;
          </button>
        </li>
      </ul>
    </template>
    <p v-else class="prompt">Open a Git repository to get started.</p>
  </div>
</template>

<style scoped>
.recent-repos {
  text-align: center;
}

.prompt {
  color: #8b949e;
  font-size: 15px;
  margin-bottom: 16px;
}

.repo-list {
  list-style: none;
  padding: 0;
  margin: 0 auto;
  max-width: 600px;
}

.repo-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 10px 14px;
  border: 1px solid #30363d;
  border-radius: 6px;
  margin-bottom: 6px;
  cursor: pointer;
  transition: background 0.15s, border-color 0.15s;
}

.repo-item:hover {
  background: #21262d;
  border-color: #6e7681;
}

.repo-info {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  overflow: hidden;
  min-width: 0;
}

.repo-name {
  font-size: 14px;
  font-weight: 600;
  color: #c9d1d9;
}

.repo-path {
  font-size: 12px;
  color: #8b949e;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 100%;
}

.remove-btn {
  flex-shrink: 0;
  background: none;
  border: none;
  color: #8b949e;
  font-size: 18px;
  cursor: pointer;
  padding: 2px 6px;
  border-radius: 4px;
  line-height: 1;
}

.remove-btn:hover {
  color: #f85149;
  background: rgba(248, 81, 73, 0.1);
}
</style>
