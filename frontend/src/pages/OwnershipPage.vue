<script lang="ts" setup>
import type { EChartsOption } from 'echarts'
import { TreemapChart } from 'echarts/charts'
import { TooltipComponent, VisualMapComponent } from 'echarts/components'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { computed, inject, onMounted, type Ref, ref, watch } from 'vue'
import VChart from 'vue-echarts'
import { FileOwnerships } from '../../wailsjs/go/main/App'
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
}

type OwnershipItem = {
  path: string
  top_author_name: string
  top_author_email: string
  top_author_pct: number
  second_author_name: string
  second_author_email: string
  second_author_pct: number
  contributor_count: number
  total_lines: number
}

type TreeNode = {
  name: string
  value?: number[]
  children?: TreeNode[]
}

type NodeStats = {
  totalLines: number
  topAuthorName: string
  topAuthorPct: number
  secondAuthorName: string
  secondAuthorPct: number
  contributorCount: number
  fileCount: number
  highRiskCount: number
  minContributors: number
}

const repoPath = inject<Ref<string>>('repoPath', ref(''))
const { patterns, addPattern, removePattern } = useExcludePatterns(repoPath)
const { presets, activePreset, customFrom, customTo, fromStr, toStr, setPreset } = useDateRange()

const loading = ref(false)
const error = ref('')
const chartOption = ref<EChartsOption | null>(null)
const rawData = ref<OwnershipItem[]>([])

const totalFiles = computed(() => rawData.value.length)
const highRiskFiles = computed(() => rawData.value.filter((f) => f.top_author_pct > 80).length)
const singleContributorFiles = computed(() => rawData.value.filter((f) => f.contributor_count === 1).length)

function buildTree(items: OwnershipItem[]): { tree: TreeNode[]; stats: Map<string, NodeStats> } {
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
        // value: [total_lines (sizing), top_author_pct (coloring)]
        child.value = [item.total_lines, item.top_author_pct]
        stats.set(item.path, {
          totalLines: item.total_lines,
          topAuthorName: item.top_author_name,
          topAuthorPct: item.top_author_pct,
          secondAuthorName: item.second_author_name,
          secondAuthorPct: item.second_author_pct,
          contributorCount: item.contributor_count,
          fileCount: 1,
          highRiskCount: item.top_author_pct > 80 ? 1 : 0,
          minContributors: item.contributor_count,
        })
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
          totalLines: 0,
          topAuthorName: '',
          topAuthorPct: 0,
          secondAuthorName: '',
          secondAuthorPct: 0,
          contributorCount: 0,
          fileCount: 0,
          highRiskCount: 0,
          minContributors: Infinity,
        }
      )
    }
    let totalLines = 0,
      fileCount = 0,
      highRiskCount = 0,
      minContrib = Infinity
    for (const child of node.children) {
      const s = aggregate(child, [...pathParts, child.name])
      totalLines += s.totalLines
      fileCount += s.fileCount
      highRiskCount += s.highRiskCount
      minContrib = Math.min(minContrib, s.minContributors)
    }
    // For directories, use average top_author_pct weighted by lines for coloring
    let weightedPct = 0
    for (const child of node.children) {
      const childPath = [...pathParts, child.name].join('/')
      const cs = stats.get(childPath)
      if (cs && cs.totalLines > 0) {
        weightedPct += cs.topAuthorPct * cs.totalLines
      }
    }
    const avgPct = totalLines > 0 ? weightedPct / totalLines : 0

    node.value = [totalLines, avgPct]
    const key = pathParts.join('/')
    const dirStats: NodeStats = {
      totalLines,
      topAuthorName: '',
      topAuthorPct: avgPct,
      secondAuthorName: '',
      secondAuthorPct: 0,
      contributorCount: 0,
      fileCount,
      highRiskCount,
      minContributors: minContrib === Infinity ? 0 : minContrib,
    }
    stats.set(key, dirStats)
    return dirStats
  }
  aggregate(root, [])

  return { tree: root.children || [], stats }
}

async function fetchData() {
  if (!fromStr.value || !toStr.value) return
  loading.value = true
  error.value = ''

  try {
    const data = await FileOwnerships(fromStr.value, toStr.value, patterns.value)
    rawData.value = data || []
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
            .map((n) => n.name)
            .join('/')
          const s = stats.get(fullPath)
          if (!s) return `<b>${info.name}</b>`

          if (s.fileCount > 1) {
            // Directory tooltip
            return (
              `<b>${fullPath || info.name}</b><br/>` +
              `Files: ${s.fileCount}<br/>` +
              `High-risk files (&gt;80%): ${s.highRiskCount}<br/>` +
              `Min contributors: ${s.minContributors}<br/>` +
              `Total lines changed: ${s.totalLines.toLocaleString()}`
            )
          }

          // File tooltip
          let html =
            `<b>${fullPath || info.name}</b><br/>` + `Top: ${s.topAuthorName} (${s.topAuthorPct.toFixed(1)}%)<br/>`
          if (s.secondAuthorName) {
            html += `2nd: ${s.secondAuthorName} (${s.secondAuthorPct.toFixed(1)}%)<br/>`
          }
          html += `Contributors: ${s.contributorCount}<br/>` + `Lines changed: ${s.totalLines.toLocaleString()}`
          return html
        },
      },
      visualMap: {
        type: 'continuous',
        dimension: 1,
        min: 0,
        max: 100,
        text: ['High risk', 'Shared'],
        inRange: {
          color: ['#3fb950', '#d29922', '#f85149'],
        },
        textStyle: { color: '#8b949e' },
        orient: 'horizontal',
        left: 'center',
        bottom: 0,
        itemWidth: 14,
        itemHeight: 140,
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
            bottom: 30,
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
watch([fromStr, toStr], fetchData)
watch(patterns, fetchData)
</script>

<template>
  <div class="ownership-container">
    <div class="ownership-header">
      <h3>Code Ownership</h3>
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

    <div v-if="rawData.length > 0" class="summary-bar">
      <div class="summary-card">
        <span class="summary-value">{{ totalFiles }}</span>
        <span class="summary-label">Total files</span>
      </div>
      <div class="summary-card risk">
        <span class="summary-value">{{ highRiskFiles }}</span>
        <span class="summary-label">&gt;80% one owner</span>
      </div>
      <div class="summary-card warn">
        <span class="summary-value">{{ singleContributorFiles }}</span>
        <span class="summary-label">Single contributor</span>
      </div>
    </div>

    <div v-if="loading" class="ownership-status">Loading...</div>
    <div v-else-if="error" class="ownership-status ownership-error">{{ error }}</div>
    <div v-else-if="!chartOption" class="ownership-status">No file changes found in this time range.</div>
    <v-chart
      v-else
      class="treemap-chart"
      :option="chartOption"
      autoresize
    />
  </div>
</template>

<style scoped>
.ownership-container {
  padding: 16px;
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
}

.ownership-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
  flex-shrink: 0;
}

.ownership-header h3 {
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

.summary-bar {
  display: flex;
  gap: 12px;
  margin-bottom: 12px;
  flex-shrink: 0;
}

.summary-card {
  background: #161b22;
  border: 1px solid #30363d;
  border-radius: 6px;
  padding: 8px 16px;
  display: flex;
  flex-direction: column;
  align-items: center;
  min-width: 100px;
}

.summary-value {
  font-size: 20px;
  font-weight: 600;
  color: #c9d1d9;
}

.summary-label {
  font-size: 11px;
  color: #8b949e;
  margin-top: 2px;
}

.summary-card.risk .summary-value {
  color: #f85149;
}

.summary-card.warn .summary-value {
  color: #d29922;
}

.treemap-chart {
  flex: 1;
  min-height: 0;
}

.ownership-status {
  color: #8b949e;
  font-size: 14px;
  text-align: center;
  padding-top: 40px;
}

.ownership-error {
  color: #f85149;
}
</style>
