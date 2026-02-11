<script lang="ts" setup>
import { computed, inject, onMounted, type Ref, ref, watch } from 'vue'
import { CoChanges } from '../../wailsjs/go/main/App'
import DateRangeSelector from '../components/DateRangeSelector.vue'
import ExcludeFilter from '../components/ExcludeFilter.vue'
import { useDateRange } from '../composables/useDateRange'
import { useExcludePatterns } from '../composables/useExcludePatterns'

type CoChangePair = {
  file_a: string
  file_b: string
  co_change_count: number
  commits_a: number
  commits_b: number
  coupling_ratio: number
}

type SortKey = 'co_change_count' | 'coupling_ratio' | 'file_a'

const repoPath = inject<Ref<string>>('repoPath', ref(''))
const { patterns, addPattern, removePattern } = useExcludePatterns(repoPath)
const { presets, activePreset, customFrom, customTo, fromStr, toStr, setPreset } = useDateRange()

const loading = ref(false)
const error = ref('')
const rawData = ref<CoChangePair[]>([])
const sortKey = ref<SortKey>('co_change_count')
const sortAsc = ref(false)

const sortedData = computed(() => {
  const data = [...rawData.value]
  const key = sortKey.value
  const dir = sortAsc.value ? 1 : -1
  data.sort((a, b) => {
    if (key === 'file_a') {
      return dir * a.file_a.localeCompare(b.file_a)
    }
    return dir * ((a[key] as number) - (b[key] as number))
  })
  return data
})

function toggleSort(key: SortKey) {
  if (sortKey.value === key) {
    sortAsc.value = !sortAsc.value
  } else {
    sortKey.value = key
    sortAsc.value = key === 'file_a'
  }
}

function sortIndicator(key: SortKey): string {
  if (sortKey.value !== key) return ''
  return sortAsc.value ? ' \u25B2' : ' \u25BC'
}

function ratioClass(ratio: number): string {
  if (ratio >= 0.8) return 'ratio-high'
  if (ratio >= 0.5) return 'ratio-medium'
  return 'ratio-low'
}

async function fetchData() {
  if (!fromStr.value || !toStr.value) return
  loading.value = true
  error.value = ''

  try {
    const data = await CoChanges(fromStr.value, toStr.value, 2, 100, patterns.value)
    rawData.value = data || []
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    loading.value = false
  }
}

onMounted(fetchData)
watch([fromStr, toStr], fetchData)
watch(patterns, fetchData)
</script>

<template>
  <div class="coupling-container">
    <div class="coupling-header">
      <h3>File Co-Change Analysis</h3>
      <div class="controls">
        <ExcludeFilter
          :patterns="patterns"
          @add="addPattern"
          @remove="removePattern"
        />
        <DateRangeSelector
          :presets="presets"
          :active-preset="activePreset"
          :custom-from="customFrom"
          :custom-to="customTo"
          @select-preset="setPreset"
          @update:custom-from="customFrom = $event"
          @update:custom-to="customTo = $event"
        />
      </div>
    </div>

    <div v-if="loading" class="coupling-status">Loading...</div>
    <div v-else-if="error" class="coupling-status coupling-error">{{ error }}</div>
    <div v-else-if="rawData.length === 0" class="coupling-status">No co-changing file pairs found in this time range.</div>
    <div v-else class="table-wrapper">
      <table class="coupling-table">
        <thead>
          <tr>
            <th class="sortable" @click="toggleSort('file_a')">
              File A{{ sortIndicator('file_a') }}
            </th>
            <th>File B</th>
            <th class="sortable num" @click="toggleSort('co_change_count')">
              Co-changes{{ sortIndicator('co_change_count') }}
            </th>
            <th class="sortable num" @click="toggleSort('coupling_ratio')">
              Ratio{{ sortIndicator('coupling_ratio') }}
            </th>
            <th class="num">Commits (A / B)</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="pair in sortedData" :key="pair.file_a + '::' + pair.file_b">
            <td class="file-path">{{ pair.file_a }}</td>
            <td class="file-path">{{ pair.file_b }}</td>
            <td class="num">{{ pair.co_change_count }}</td>
            <td class="num">
              <span :class="ratioClass(pair.coupling_ratio)">
                {{ Math.round(pair.coupling_ratio * 100) }}%
              </span>
            </td>
            <td class="num">{{ pair.commits_a }} / {{ pair.commits_b }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<style scoped>
.coupling-container {
  padding: 16px;
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
}

.coupling-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
  flex-shrink: 0;
}

.coupling-header h3 {
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

.table-wrapper {
  flex: 1;
  overflow: auto;
  min-height: 0;
}

.coupling-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.coupling-table th {
  position: sticky;
  top: 0;
  background: #161b22;
  color: #8b949e;
  font-weight: 600;
  text-align: left;
  padding: 8px 12px;
  border-bottom: 1px solid #30363d;
  white-space: nowrap;
  user-select: none;
}

.coupling-table th.sortable {
  cursor: pointer;
}

.coupling-table th.sortable:hover {
  color: #c9d1d9;
}

.coupling-table th.num,
.coupling-table td.num {
  text-align: right;
}

.coupling-table td {
  padding: 6px 12px;
  border-bottom: 1px solid #21262d;
  color: #c9d1d9;
}

.coupling-table tbody tr:hover {
  background: #161b22;
}

.file-path {
  font-family: monospace;
  font-size: 12px;
  word-break: break-all;
}

.ratio-high {
  color: #f85149;
  font-weight: 600;
}

.ratio-medium {
  color: #d29922;
  font-weight: 600;
}

.ratio-low {
  color: #8b949e;
}

.coupling-status {
  color: #8b949e;
  font-size: 14px;
  text-align: center;
  padding-top: 40px;
}

.coupling-error {
  color: #f85149;
}
</style>
