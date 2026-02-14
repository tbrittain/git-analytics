<script lang="ts" setup>
import type { EChartsOption } from 'echarts'
import { BarChart } from 'echarts/charts'
import { GridComponent, TooltipComponent } from 'echarts/components'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { onMounted, ref } from 'vue'
import VChart from 'vue-echarts'
import { CommitsByHour, DashboardStats, RepoInfo } from '../../wailsjs/go/main/App'
import CommitHeatmap from '../components/CommitHeatmap.vue'
import { formatDate } from '../composables/useDateRange'

use([BarChart, GridComponent, TooltipComponent, CanvasRenderer])

const repoInfo = ref<{
  name: string
  branch: string
  head_hash: string
  last_author: string
  last_message: string
  last_commit_age: string
} | null>(null)

const stats = ref<{
  commits: number
  contributors: number
  additions: number
  deletions: number
  files_changed: number
} | null>(null)

const chartOption = ref<EChartsOption | null>(null)
const error = ref('')

function formatNumber(n: number): string {
  return n.toLocaleString()
}

onMounted(async () => {
  try {
    const to = new Date()
    to.setDate(to.getDate() + 1)
    const from = new Date()
    from.setDate(from.getDate() - 29) // 30 days including today

    const fromStr = formatDate(from)
    const toStr = formatDate(to)

    const [info, dashStats, hourData] = await Promise.all([
      RepoInfo(),
      DashboardStats(fromStr, toStr),
      CommitsByHour(fromStr, toStr),
    ])

    repoInfo.value = info
    stats.value = dashStats

    // Build full 24-hour array from sparse data
    const counts = new Array(24).fill(0)
    if (hourData) {
      for (const b of hourData) {
        counts[b.hour] = b.count
      }
    }

    chartOption.value = {
      tooltip: {
        trigger: 'axis',
        axisPointer: { type: 'shadow' },
        formatter(params: unknown) {
          const p = Array.isArray(params) ? params[0] : params
          const item = p as { dataIndex: number; value: number }
          const label = `${String(item.dataIndex).padStart(2, '0')}:00`
          return `${label}<br/>${item.value} commit${item.value === 1 ? '' : 's'}`
        },
      },
      grid: {
        left: 40,
        right: 16,
        top: 8,
        bottom: 24,
      },
      xAxis: {
        type: 'category',
        data: Array.from({ length: 24 }, (_, i) => `${String(i).padStart(2, '0')}`),
        axisLabel: {
          color: '#8b949e',
          fontSize: 11,
          interval: 2,
        },
        axisLine: { lineStyle: { color: '#30363d' } },
        axisTick: { show: false },
      },
      yAxis: {
        type: 'value',
        minInterval: 1,
        axisLabel: { color: '#8b949e', fontSize: 11 },
        splitLine: { lineStyle: { color: '#21262d' } },
      },
      series: [
        {
          type: 'bar',
          data: counts,
          itemStyle: { color: '#58a6ff', borderRadius: [2, 2, 0, 0] },
          barMaxWidth: 20,
        },
      ],
    }
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : String(e)
  }
})
</script>

<template>
  <div class="dashboard">
    <div v-if="error" class="dashboard-error">{{ error }}</div>

    <!-- Repo Header -->
    <div v-if="repoInfo" class="repo-header">
      <h1 class="repo-name">{{ repoInfo.name }}</h1>
      <div class="repo-meta">
        <span class="repo-branch">{{ repoInfo.branch }}</span>
        <span class="repo-hash">@ {{ repoInfo.head_hash }}</span>
      </div>
      <div v-if="repoInfo.last_author" class="repo-last-commit">
        {{ repoInfo.last_author }} &middot; {{ repoInfo.last_commit_age }} &middot;
        &ldquo;{{ repoInfo.last_message }}&rdquo;
      </div>
    </div>

    <!-- Stat Cards -->
    <div v-if="stats" class="stat-cards">
      <div class="stat-card">
        <div class="stat-value">{{ formatNumber(stats.commits) }}</div>
        <div class="stat-label">Commits <span class="stat-period">(30d)</span></div>
      </div>
      <div class="stat-card">
        <div class="stat-value">{{ formatNumber(stats.contributors) }}</div>
        <div class="stat-label">Contributors <span class="stat-period">(30d)</span></div>
      </div>
      <div class="stat-card">
        <div class="stat-value stat-additions">+{{ formatNumber(stats.additions) }}</div>
        <div class="stat-label">Added <span class="stat-period">(30d)</span></div>
      </div>
      <div class="stat-card">
        <div class="stat-value stat-deletions">-{{ formatNumber(stats.deletions) }}</div>
        <div class="stat-label">Deleted <span class="stat-period">(30d)</span></div>
      </div>
      <div class="stat-card">
        <div class="stat-value">{{ formatNumber(stats.files_changed) }}</div>
        <div class="stat-label">Files Changed <span class="stat-period">(30d)</span></div>
      </div>
    </div>

    <!-- Commit Time-of-Day Histogram -->
    <div v-if="chartOption" class="chart-section">
      <h3>Commit Time-of-Day (30d)</h3>
      <v-chart :option="chartOption" :autoresize="true" class="hour-chart" />
    </div>

    <!-- Existing Heatmap -->
    <CommitHeatmap />
  </div>
</template>

<style scoped>
.dashboard {
  padding: 24px;
  max-width: 960px;
  margin: 0 auto;
}

.dashboard-error {
  color: #f85149;
  font-size: 14px;
  margin-bottom: 16px;
}

/* Repo Header */
.repo-header {
  margin-bottom: 20px;
}

.repo-name {
  margin: 0;
  font-size: 24px;
  font-weight: 700;
  color: #e6edf3;
}

.repo-meta {
  margin-top: 4px;
  font-size: 14px;
  color: #8b949e;
}

.repo-branch {
  color: #58a6ff;
  font-weight: 500;
}

.repo-hash {
  margin-left: 4px;
  font-family: monospace;
}

.repo-last-commit {
  margin-top: 4px;
  font-size: 13px;
  color: #8b949e;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Stat Cards */
.stat-cards {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 12px;
  margin-bottom: 24px;
}

.stat-card {
  background: #21262d;
  border: 1px solid #30363d;
  border-radius: 8px;
  padding: 16px;
  text-align: center;
}

.stat-value {
  font-size: 22px;
  font-weight: 700;
  color: #e6edf3;
  font-variant-numeric: tabular-nums;
}

.stat-additions {
  color: #3fb950;
}

.stat-deletions {
  color: #f85149;
}

.stat-label {
  margin-top: 4px;
  font-size: 12px;
  color: #8b949e;
}

.stat-period {
  color: #6e7681;
}

/* Hour Chart */
.chart-section {
  margin-bottom: 24px;
}

.chart-section h3 {
  margin: 0 0 12px 0;
  font-size: 16px;
  font-weight: 600;
  color: #c9d1d9;
}

.hour-chart {
  height: 200px;
  width: 100%;
}
</style>
