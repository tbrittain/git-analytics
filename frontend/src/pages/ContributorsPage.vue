<script lang="ts" setup>
import { inject, onMounted, type Ref, ref, watch } from 'vue'
import { Contributors } from '../../wailsjs/go/main/App'
import type { query } from '../../wailsjs/go/models'
import ExcludeFilter from '../components/ExcludeFilter.vue'
import { useExcludePatterns } from '../composables/useExcludePatterns'

type Preset = {
  label: string
  days: number | null
}

const presets: Preset[] = [
  { label: '30d', days: 30 },
  { label: '90d', days: 90 },
  { label: '6mo', days: 182 },
  { label: '1yr', days: 365 },
  { label: 'All', days: null },
]

const repoPath = inject<Ref<string>>('repoPath', ref(''))
const { patterns, addPattern, removePattern } = useExcludePatterns(repoPath)

const activePreset = ref(2) // default 6mo
const loading = ref(false)
const error = ref('')
const contributors = ref<query.Contributor[]>([])

function formatDate(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

function formatNumber(n: number): string {
  return n.toLocaleString()
}

async function fetchData() {
  loading.value = true
  error.value = ''

  try {
    const to = new Date()
    to.setDate(to.getDate() + 1)
    const toStr = formatDate(to)

    let fromStr: string
    const preset = presets[activePreset.value]
    if (preset.days !== null) {
      const from = new Date()
      from.setDate(from.getDate() - preset.days)
      fromStr = formatDate(from)
    } else {
      fromStr = '1970-01-01'
    }

    const data = await Contributors(fromStr, toStr, patterns.value)
    contributors.value = data || []
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    loading.value = false
  }
}

onMounted(fetchData)
watch(activePreset, fetchData)
watch(patterns, fetchData)
</script>

<template>
  <div class="contributors-container">
    <div class="contributors-header">
      <h3>Contributors</h3>
      <div class="controls">
        <ExcludeFilter
          :patterns="patterns"
          @add="addPattern"
          @remove="removePattern"
        />
        <div class="presets">
          <button
            v-for="(preset, i) in presets"
            :key="preset.label"
            :class="['preset-btn', { active: activePreset === i }]"
            @click="activePreset = i"
          >
            {{ preset.label }}
          </button>
        </div>
      </div>
    </div>

    <div v-if="loading" class="contributors-status">Loading...</div>
    <div v-else-if="error" class="contributors-status contributors-error">{{ error }}</div>
    <div v-else-if="contributors.length === 0" class="contributors-status">No contributors found in this time range.</div>
    <div v-else class="table-wrapper">
      <table>
        <thead>
          <tr>
            <th class="col-rank">#</th>
            <th class="col-contributor">Contributor</th>
            <th class="col-num">Commits</th>
            <th class="col-num">Additions</th>
            <th class="col-num">Deletions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(c, i) in contributors" :key="c.author_email">
            <td class="col-rank">{{ i + 1 }}</td>
            <td class="col-contributor">
              <span class="author-name">{{ c.author_name }}</span>
              <span class="author-email">{{ c.author_email }}</span>
            </td>
            <td class="col-num">{{ formatNumber(c.commits) }}</td>
            <td class="col-num additions">+{{ formatNumber(c.additions) }}</td>
            <td class="col-num deletions">-{{ formatNumber(c.deletions) }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<style scoped>
.contributors-container {
  padding: 16px;
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
}

.contributors-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
  flex-shrink: 0;
}

.contributors-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #c9d1d9;
}

.controls {
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

.table-wrapper {
  flex: 1;
  overflow: auto;
  min-height: 0;
}

table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

thead {
  position: sticky;
  top: 0;
  z-index: 1;
}

th {
  background: #161b22;
  color: #8b949e;
  font-weight: 600;
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  padding: 8px 12px;
  text-align: left;
  border-bottom: 1px solid #30363d;
}

td {
  padding: 8px 12px;
  color: #c9d1d9;
  border-bottom: 1px solid #21262d;
}

tr:hover td {
  background: #161b22;
}

.col-rank {
  width: 48px;
  text-align: center;
  color: #8b949e;
}

th.col-rank {
  text-align: center;
}

.col-contributor {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.author-name {
  font-weight: 500;
}

.author-email {
  font-size: 12px;
  color: #8b949e;
}

.col-num {
  text-align: right;
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}

th.col-num {
  text-align: right;
}

.additions {
  color: #3fb950;
}

.deletions {
  color: #f85149;
}

.contributors-status {
  color: #8b949e;
  font-size: 14px;
  text-align: center;
  padding-top: 40px;
}

.contributors-error {
  color: #f85149;
}
</style>
