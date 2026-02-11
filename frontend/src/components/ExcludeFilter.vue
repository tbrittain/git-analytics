<script lang="ts" setup>
import { ref } from 'vue'

defineProps<{
  patterns: string[]
}>()

const emit = defineEmits<{
  add: [pattern: string]
  remove: [pattern: string]
}>()

const expanded = ref(false)
const input = ref('')

function onAdd() {
  const trimmed = input.value.trim()
  if (!trimmed) return
  emit('add', trimmed)
  input.value = ''
}
</script>

<template>
  <div class="exclude-filter">
    <button
      class="filter-toggle"
      :class="{ active: patterns.length > 0 }"
      @click="expanded = !expanded"
    >
      Filter
      <span v-if="patterns.length > 0" class="badge">{{ patterns.length }}</span>
    </button>

    <div v-if="expanded" class="filter-dropdown">
      <div class="filter-input-row">
        <input
          v-model="input"
          type="text"
          placeholder="e.g. *.lock, *.pb.go"
          @keydown.enter="onAdd"
        />
        <button class="add-btn" @click="onAdd">Add</button>
      </div>
      <div v-if="patterns.length > 0" class="chips">
        <span v-for="p in patterns" :key="p" class="chip">
          {{ p }}
          <button class="chip-remove" @click="emit('remove', p)">&times;</button>
        </span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.exclude-filter {
  position: relative;
}

.filter-toggle {
  padding: 4px 12px;
  font-size: 12px;
  border: 1px solid #30363d;
  border-radius: 6px;
  background: #21262d;
  color: #c9d1d9;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
}

.filter-toggle:hover {
  background: #30363d;
}

.filter-toggle.active {
  border-color: #1f6feb;
}

.badge {
  background: #1f6feb;
  color: #fff;
  font-size: 10px;
  padding: 1px 6px;
  border-radius: 10px;
  font-weight: 600;
}

.filter-dropdown {
  position: absolute;
  top: calc(100% + 6px);
  right: 0;
  background: #161b22;
  border: 1px solid #30363d;
  border-radius: 8px;
  padding: 10px;
  min-width: 280px;
  z-index: 10;
}

.filter-input-row {
  display: flex;
  gap: 6px;
}

.filter-input-row input {
  flex: 1;
  padding: 5px 8px;
  font-size: 12px;
  border: 1px solid #30363d;
  border-radius: 6px;
  background: #0d1117;
  color: #c9d1d9;
  outline: none;
}

.filter-input-row input:focus {
  border-color: #1f6feb;
}

.add-btn {
  padding: 5px 10px;
  font-size: 12px;
  border: 1px solid #30363d;
  border-radius: 6px;
  background: #21262d;
  color: #c9d1d9;
  cursor: pointer;
}

.add-btn:hover {
  background: #30363d;
}

.chips {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 8px;
}

.chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 3px 8px;
  font-size: 11px;
  background: #21262d;
  border: 1px solid #30363d;
  border-radius: 12px;
  color: #c9d1d9;
}

.chip-remove {
  background: none;
  border: none;
  color: #8b949e;
  cursor: pointer;
  font-size: 14px;
  padding: 0 2px;
  line-height: 1;
}

.chip-remove:hover {
  color: #f85149;
}
</style>
