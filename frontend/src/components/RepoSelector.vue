<script lang="ts" setup>
defineProps<{
  repoPath: string
  loading: boolean
}>()

defineEmits<{
  select: []
}>()
</script>

<template>
  <div class="repo-selector">
    <button class="btn" :disabled="loading" @click="$emit('select')">
      Open Repository
    </button>
    <span v-if="repoPath" class="repo-path">{{ repoPath }}</span>
  </div>
</template>

<style scoped>
.repo-selector {
  display: flex;
  align-items: center;
  gap: 12px;
}

.btn {
  padding: 6px 16px;
  border: 1px solid #444;
  border-radius: 6px;
  background: #21262d;
  color: #c9d1d9;
  cursor: pointer;
  font-size: 14px;
  white-space: nowrap;
}

.btn:hover:not(:disabled) {
  background: #30363d;
  border-color: #6e7681;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.repo-path {
  color: #8b949e;
  font-size: 13px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
