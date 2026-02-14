<script lang="ts" setup>
import type { EChartsOption } from 'echarts'
import { TreemapChart } from 'echarts/charts'
import { TooltipComponent, VisualMapComponent } from 'echarts/components'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { computed, inject, onMounted, type Ref, ref, watch } from 'vue'
import VChart from 'vue-echarts'
import { FileHotspots, TemporalHotspots } from '../../wailsjs/go/main/App'
import DateRangeSelector from '../components/DateRangeSelector.vue'
import ExcludeFilter from '../components/ExcludeFilter.vue'
import { useDateRange } from '../composables/useDateRange'
import { useExcludePatterns } from '../composables/useExcludePatterns'

use([TreemapChart, TooltipComponent, VisualMapComponent, CanvasRenderer])

type TreePathEntry = {
  name: string
  dataIndex: number
  value: number | number[]
}

type TreemapFormatterParams = {
  name: string
  value: number | number[]
  treePathInfo?: TreePathEntry[]
  treeAncestors?: TreePathEntry[]
}

const repoPath = inject<Ref<string>>('repoPath', ref(''))
const { patterns, addPattern, removePattern } = useExcludePatterns(repoPath)
const { presets, activePreset, customFrom, customTo, fromStr, toStr, setPreset } = useDateRange()

const loading = ref(false)
const error = ref('')
const chartOption = ref<EChartsOption | null>(null)
const mode = ref<'total' | 'recency' | 'movers'>('total')
const movers = ref<{ path: string; lines_changed: number; additions: number; deletions: number; commits: number }[]>([])

type SortKey = 'lines_changed' | 'additions' | 'deletions'
const sortKey = ref<SortKey>('lines_changed')

const sortedMovers = computed(() => {
  const key = sortKey.value
  return [...movers.value].sort((a, b) => b[key] - a[key])
})

type TreeNode = {
  name: string
  value?: number | number[]
  children?: TreeNode[]
}

type NodeStats = {
  lines: number
  additions: number
  deletions: number
  commits: number
  score?: number
  lastChanged?: string
  daysSince?: number
}

// Build tree for total churn mode (value = lines_changed for sizing).
function buildTree(
  items: { path: string; lines_changed: number; additions: number; deletions: number; commits: number }[],
): { tree: TreeNode[]; stats: Map<string, NodeStats> } {
  const root: TreeNode = { name: '/', children: [] }
  const stats = new Map<string, NodeStats>()

  for (const item of items) {
    const parts = item.path.replace(/\\/g, '/').split('/')
    let current = root

    for (let i = 0; i < parts.length; i++) {
      const part = parts[i]
      if (!current.children) current.children = []

      let child = current.children.find((c) => c.name === part)
      if (!child) {
        child = { name: part, children: [] }
        current.children.push(child)
      }

      if (i === parts.length - 1) {
        child.value = item.lines_changed
        stats.set(item.path, {
          lines: item.lines_changed,
          additions: item.additions,
          deletions: item.deletions,
          commits: item.commits,
        })
        delete child.children
      }

      current = child
    }
  }

  function aggregate(node: TreeNode, pathParts: string[]): NodeStats {
    if (!node.children || node.children.length === 0) {
      const key = pathParts.join('/')
      return stats.get(key) || { lines: 0, additions: 0, deletions: 0, commits: 0 }
    }
    let lines = 0,
      adds = 0,
      dels = 0,
      commits = 0
    for (const child of node.children) {
      const s = aggregate(child, [...pathParts, child.name])
      lines += s.lines
      adds += s.additions
      dels += s.deletions
      commits += s.commits
    }
    node.value = lines
    const key = pathParts.join('/')
    const dirStats = { lines, additions: adds, deletions: dels, commits }
    stats.set(key, dirStats)
    return dirStats
  }
  aggregate(root, [])

  return { tree: root.children || [], stats }
}

// Build tree for recency mode. value = [score, daysSince] for sizing + coloring.
function buildRecencyTree(
  items: {
    path: string
    lines_changed: number
    additions: number
    deletions: number
    commits: number
    last_changed: string
    days_since: number
    score: number
  }[],
): { tree: TreeNode[]; stats: Map<string, NodeStats>; maxDays: number } {
  const root: TreeNode = { name: '/', children: [] }
  const stats = new Map<string, NodeStats>()
  let maxDays = 0

  for (const item of items) {
    const parts = item.path.replace(/\\/g, '/').split('/')
    let current = root

    for (let i = 0; i < parts.length; i++) {
      const part = parts[i]
      if (!current.children) current.children = []

      let child = current.children.find((c) => c.name === part)
      if (!child) {
        child = { name: part, children: [] }
        current.children.push(child)
      }

      if (i === parts.length - 1) {
        child.value = [item.score, item.days_since]
        stats.set(item.path, {
          lines: item.lines_changed,
          additions: item.additions,
          deletions: item.deletions,
          commits: item.commits,
          score: item.score,
          lastChanged: item.last_changed,
          daysSince: item.days_since,
        })
        if (item.days_since > maxDays) maxDays = item.days_since
        delete child.children
      }

      current = child
    }
  }

  function aggregate(node: TreeNode, pathParts: string[]): NodeStats {
    if (!node.children || node.children.length === 0) {
      const key = pathParts.join('/')
      return (
        stats.get(key) || {
          lines: 0,
          additions: 0,
          deletions: 0,
          commits: 0,
          score: 0,
          daysSince: 0,
        }
      )
    }
    let lines = 0,
      adds = 0,
      dels = 0,
      commits = 0,
      score = 0,
      minDays = Infinity
    for (const child of node.children) {
      const s = aggregate(child, [...pathParts, child.name])
      lines += s.lines
      adds += s.additions
      dels += s.deletions
      commits += s.commits
      score += s.score ?? 0
      if ((s.daysSince ?? Infinity) < minDays) minDays = s.daysSince ?? Infinity
    }
    node.value = [score, minDays === Infinity ? 0 : minDays]
    const key = pathParts.join('/')
    const dirStats: NodeStats = {
      lines,
      additions: adds,
      deletions: dels,
      commits,
      score,
      daysSince: minDays === Infinity ? 0 : minDays,
    }
    stats.set(key, dirStats)
    return dirStats
  }
  aggregate(root, [])

  return { tree: root.children || [], stats, maxDays }
}

function buildFullPath(info: TreemapFormatterParams): string {
  const treePath = info.treePathInfo ?? []
  return treePath
    .slice(1)
    .map((n: TreePathEntry) => n.name)
    .join('/')
}

async function fetchData() {
  if (!fromStr.value || !toStr.value) return
  loading.value = true
  error.value = ''

  try {
    if (mode.value === 'movers') {
      const data = await FileHotspots(fromStr.value, toStr.value, patterns.value)
      movers.value = data || []
      chartOption.value = null
      return
    }

    if (mode.value === 'recency') {
      const data = await TemporalHotspots(fromStr.value, toStr.value, 90, patterns.value)
      if (!data || data.length === 0) {
        chartOption.value = null
        return
      }

      const { tree, stats, maxDays } = buildRecencyTree(data)

      chartOption.value = {
        tooltip: {
          formatter(params) {
            const info = params as TreemapFormatterParams
            const fullPath = buildFullPath(info)
            const s = stats.get(fullPath)
            if (!s) return `<b>${fullPath || info.name}</b>`
            return (
              `<b>${fullPath || info.name}</b><br/>` +
              `Score: ${s.score?.toFixed(1)}<br/>` +
              `Lines changed: ${s.lines.toLocaleString()}<br/>` +
              `<span style="color:#3fb950">+${s.additions.toLocaleString()}</span>` +
              ` / <span style="color:#f85149">-${s.deletions.toLocaleString()}</span><br/>` +
              `Last changed: ${s.lastChanged ?? '—'} (${s.daysSince ?? 0}d ago)<br/>` +
              `Commits: ${s.commits.toLocaleString()}`
            )
          },
        },
        visualMap: {
          type: 'continuous',
          dimension: 1,
          min: 0,
          max: Math.max(maxDays, 1),
          inRange: {
            color: ['#f85149', '#d29922', '#58a6ff'],
          },
          textStyle: { color: '#8b949e' },
          text: ['Old', 'Recent'],
          right: 10,
          bottom: 50,
          calculable: false,
        },
        series: [
          {
            type: 'treemap',
            data: tree,
            leafDepth: 2,
            width: '100%',
            height: '90%',
            roam: false,
            nodeClick: 'zoomToNode',
            breadcrumb: {
              show: true,
              bottom: 10,
              itemStyle: {
                color: '#21262d',
                borderColor: '#30363d',
                textStyle: { color: '#c9d1d9' },
              },
              emphasis: {
                itemStyle: { color: '#30363d' },
              },
            },
            upperLabel: {
              show: true,
              height: 20,
              color: '#c9d1d9',
              fontSize: 12,
              backgroundColor: 'transparent',
            },
            itemStyle: {
              borderColor: '#0d1117',
              borderWidth: 2,
              gapWidth: 2,
            },
            levels: [
              {
                itemStyle: {
                  borderColor: '#30363d',
                  borderWidth: 3,
                  gapWidth: 3,
                },
                upperLabel: {
                  show: true,
                  color: '#c9d1d9',
                  fontSize: 13,
                  fontWeight: 600,
                },
              },
              {
                itemStyle: {
                  gapWidth: 1,
                },
              },
            ],
            visualMin: 0,
            label: {
              show: true,
              formatter: '{b}',
              color: '#c9d1d9',
              fontSize: 11,
            },
          },
        ],
      }
    } else {
      const data = await FileHotspots(fromStr.value, toStr.value, patterns.value)
      if (!data || data.length === 0) {
        chartOption.value = null
        return
      }

      const { tree, stats } = buildTree(data)

      chartOption.value = {
        tooltip: {
          formatter(params) {
            const info = params as TreemapFormatterParams
            const fullPath = buildFullPath(info)
            const s = stats.get(fullPath)
            const lines = s ? s.lines.toLocaleString() : '—'
            const adds = s ? s.additions.toLocaleString() : '—'
            const dels = s ? s.deletions.toLocaleString() : '—'
            const commits = s ? s.commits.toLocaleString() : '—'
            return (
              `<b>${fullPath || info.name}</b><br/>` +
              `Lines changed: ${lines}<br/>` +
              `<span style="color:#3fb950">+${adds}</span>` +
              ` / <span style="color:#f85149">-${dels}</span><br/>` +
              `Commits: ${commits}`
            )
          },
        },
        series: [
          {
            type: 'treemap',
            data: tree,
            leafDepth: 2,
            width: '100%',
            height: '100%',
            roam: false,
            nodeClick: 'zoomToNode',
            breadcrumb: {
              show: true,
              bottom: 10,
              itemStyle: {
                color: '#21262d',
                borderColor: '#30363d',
                textStyle: { color: '#c9d1d9' },
              },
              emphasis: {
                itemStyle: { color: '#30363d' },
              },
            },
            upperLabel: {
              show: true,
              height: 20,
              color: '#c9d1d9',
              fontSize: 12,
              backgroundColor: 'transparent',
            },
            itemStyle: {
              borderColor: '#0d1117',
              borderWidth: 2,
              gapWidth: 2,
            },
            levels: [
              {
                itemStyle: {
                  borderColor: '#30363d',
                  borderWidth: 3,
                  gapWidth: 3,
                },
                upperLabel: {
                  show: true,
                  color: '#c9d1d9',
                  fontSize: 13,
                  fontWeight: 600,
                },
              },
              {
                colorSaturation: [0.3, 0.7],
                itemStyle: {
                  borderColorSaturation: 0.6,
                  gapWidth: 1,
                },
              },
            ],
            visualMin: 0,
            color: ['#0e4429', '#006d32', '#26a641', '#39d353', '#58a6ff'],
            label: {
              show: true,
              formatter: '{b}',
              color: '#c9d1d9',
              fontSize: 11,
            },
          },
        ],
      }
    }
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    loading.value = false
  }
}

onMounted(fetchData)
watch([fromStr, toStr, mode], fetchData)
watch(patterns, fetchData)
</script>

<template>
  <div class="hotspots-container">
    <div class="hotspots-header">
      <h3>Code Hotspots</h3>
      <div class="controls">
        <div class="mode-toggle">
          <button
            :class="['mode-btn', { active: mode === 'total' }]"
            @click="mode = 'total'"
          >
            Total Churn
          </button>
          <button
            :class="['mode-btn', { active: mode === 'recency' }]"
            @click="mode = 'recency'"
          >
            Churn x Recency
          </button>
          <button
            :class="['mode-btn', { active: mode === 'movers' }]"
            @click="mode = 'movers'"
          >
            Top Movers
          </button>
        </div>
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

    <div v-if="loading" class="hotspots-status">Loading...</div>
    <div v-else-if="error" class="hotspots-status hotspots-error">{{ error }}</div>
    <template v-else-if="mode === 'movers'">
      <div v-if="movers.length === 0" class="hotspots-status">No file changes found in this time range.</div>
      <div v-else class="table-wrapper">
        <table>
          <thead>
            <tr>
              <th class="col-rank">#</th>
              <th class="col-file">File</th>
              <th class="col-num sortable" @click="sortKey = 'lines_changed'">
                Lines Changed <span v-if="sortKey === 'lines_changed'" class="sort-indicator">&#x25BC;</span>
              </th>
              <th class="col-num sortable" @click="sortKey = 'additions'">
                Additions <span v-if="sortKey === 'additions'" class="sort-indicator">&#x25BC;</span>
              </th>
              <th class="col-num sortable" @click="sortKey = 'deletions'">
                Deletions <span v-if="sortKey === 'deletions'" class="sort-indicator">&#x25BC;</span>
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(f, i) in sortedMovers" :key="f.path">
              <td class="col-rank">{{ i + 1 }}</td>
              <td class="col-file">{{ f.path }}</td>
              <td class="col-num">{{ f.lines_changed.toLocaleString() }}</td>
              <td class="col-num additions">+{{ f.additions.toLocaleString() }}</td>
              <td class="col-num deletions">-{{ f.deletions.toLocaleString() }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </template>
    <template v-else>
      <div v-if="!chartOption" class="hotspots-status">No file changes found in this time range.</div>
      <v-chart
        v-else
        class="treemap-chart"
        :option="chartOption"
        autoresize
      />
    </template>
  </div>
</template>

<style scoped>
.hotspots-container {
  padding: 16px;
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
}

.hotspots-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
  flex-shrink: 0;
}

.hotspots-header h3 {
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

.mode-toggle {
  display: flex;
  border: 1px solid #30363d;
  border-radius: 6px;
  overflow: hidden;
}

.mode-btn {
  padding: 4px 12px;
  font-size: 12px;
  background: #21262d;
  color: #8b949e;
  border: none;
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
}

.mode-btn + .mode-btn {
  border-left: 1px solid #30363d;
}

.mode-btn.active {
  background: #1f6feb;
  color: #ffffff;
}

.mode-btn:hover:not(.active) {
  background: #30363d;
  color: #c9d1d9;
}

.treemap-chart {
  flex: 1;
  min-height: 0;
}

.hotspots-status {
  color: #8b949e;
  font-size: 14px;
  text-align: center;
  padding-top: 40px;
}

.hotspots-error {
  color: #f85149;
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

.col-file {
  font-family: monospace;
  font-size: 12px;
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

.sortable {
  cursor: pointer;
  user-select: none;
}

.sortable:hover {
  color: #c9d1d9;
}

.sort-indicator {
  font-size: 10px;
  margin-left: 2px;
}
</style>
