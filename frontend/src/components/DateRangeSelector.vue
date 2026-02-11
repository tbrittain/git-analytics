<script lang="ts" setup>
import { computed, ref } from 'vue'
import { formatDate, type Preset } from '../composables/useDateRange'

defineProps<{
  presets: Preset[]
  activePreset: number | null
  customFrom: string
  customTo: string
}>()

const emit = defineEmits<{
  'select-preset': [index: number]
  'update:customFrom': [value: string]
  'update:customTo': [value: string]
}>()

const today = computed(() => formatDate(new Date()))
const showCustom = ref(false)

function onSelectPreset(index: number) {
  showCustom.value = false
  emit('select-preset', index)
}

function onToggleCustom() {
  showCustom.value = !showCustom.value
}
</script>

<template>
  <div class="date-range-selector">
    <div class="presets">
      <button
        v-for="(preset, i) in presets"
        :key="preset.label"
        :class="['preset-btn', { active: activePreset === i }]"
        @click="onSelectPreset(i)"
      >
        {{ preset.label }}
      </button>
      <button
        :class="['preset-btn', { active: activePreset === null }]"
        @click="onToggleCustom"
      >
        Custom
      </button>
    </div>
    <div v-if="showCustom || activePreset === null" class="custom-range">
      <input
        type="date"
        class="date-input"
        :value="customFrom"
        :max="customTo || today"
        @input="emit('update:customFrom', ($event.target as HTMLInputElement).value)"
      />
      <span class="date-separator">to</span>
      <input
        type="date"
        class="date-input"
        :value="customTo"
        :min="customFrom"
        :max="today"
        @input="emit('update:customTo', ($event.target as HTMLInputElement).value)"
      />
    </div>
  </div>
</template>

<style scoped>
.date-range-selector {
  display: flex;
  align-items: center;
  gap: 8px;
}

.presets {
  display: flex;
  gap: 4px;
}

.preset-btn {
  padding: 4px 12px;
  font-size: 12px;
  border: 1px solid #30363d;
  border-radius: 6px;
  background: #21262d;
  color: #c9d1d9;
  cursor: pointer;
}

.preset-btn:hover {
  background: #30363d;
}

.preset-btn.active {
  background: #1f6feb;
  border-color: #1f6feb;
  color: #ffffff;
}

.custom-range {
  display: flex;
  align-items: center;
  gap: 6px;
}

.date-input {
  padding: 3px 8px;
  font-size: 12px;
  border: 1px solid #30363d;
  border-radius: 6px;
  background: #21262d;
  color: #c9d1d9;
  color-scheme: dark;
}

.date-separator {
  font-size: 12px;
  color: #8b949e;
}
</style>
