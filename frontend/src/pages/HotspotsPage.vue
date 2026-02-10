<script lang="ts" setup>
import type { EChartsOption } from 'echarts'
import { TreemapChart } from 'echarts/charts'
import { TooltipComponent } from 'echarts/components'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { onMounted, ref, watch } from 'vue'
import VChart from 'vue-echarts'
import { FileHotspots } from '../../wailsjs/go/main/App'

use([TreemapChart, TooltipComponent, CanvasRenderer])

interface TreePathEntry {
  name: string
  dataIndex: number
  value: number | number[]
}

interface TreemapFormatterParams {
  name: string
  value: number | number[]
  treePathInfo?: TreePathEntry[]
  treeAncestors?: TreePathEntry[]
}

interface Preset {
  label: string
  days: number | null // null = all time
}

const presets: Preset[] = [
  { label: '30d', days: 30 },
  { label: '90d', days: 90 },
  { label: '6mo', days: 182 },
  { label: '1yr', days: 365 },
  { label: 'All', days: null },
]

const activePreset = ref(2) // default 6mo
const loading = ref(false)
const error = ref('')
const chartOption = ref<EChartsOption | null>(null)

function formatDate(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

interface TreeNode {
  name: string
  value?: number
  children?: TreeNode[]
}

interface NodeStats {
  lines: number
  additions: number
  deletions: number
  commits: number
}

// Build tree for ECharts (value = lines_changed for sizing) and a side-channel
// stats map keyed by full path for the tooltip.
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

  // Roll up stats for directory nodes and set value for sizing.
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

async function fetchData() {
  loading.value = true
  error.value = ''

  try {
    const to = new Date()
    to.setDate(to.getDate() + 1) // exclusive end
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

    const data = await FileHotspots(fromStr, toStr)
    if (!data || data.length === 0) {
      chartOption.value = null
      return
    }

    const { tree, stats } = buildTree(data)

    chartOption.value = {
      tooltip: {
        formatter(params) {
          const info = params as TreemapFormatterParams
          const treePath = info.treePathInfo ?? []
          const fullPath = treePath
            .slice(1)
            .map((n: TreePathEntry) => n.name)
            .join('/')
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
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    loading.value = false
  }
}

onMounted(fetchData)
watch(activePreset, fetchData)
</script>

<template>
  <div class="hotspots-container">
    <div class="hotspots-header">
      <h3>Code Hotspots</h3>
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

    <div v-if="loading" class="hotspots-status">Loading...</div>
    <div v-else-if="error" class="hotspots-status hotspots-error">{{ error }}</div>
    <div v-else-if="!chartOption" class="hotspots-status">No file changes found in this time range.</div>
    <v-chart
      v-else
      class="treemap-chart"
      :option="chartOption"
      autoresize
    />
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
</style>
