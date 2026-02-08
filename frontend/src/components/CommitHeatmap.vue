<script lang="ts" setup>
import { ref, onMounted } from 'vue'
import { CommitHeatmap } from '../../wailsjs/go/main/App'

const cells = ref<{ date: string; count: number; dayOfWeek: number }[]>([])
const months = ref<{ label: string; col: number }[]>([])
const error = ref('')

const MONTH_NAMES = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
const DAY_LABELS = ['', 'Mon', '', 'Wed', '', 'Fri', '']

function formatDate(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

function colorForCount(count: number): string {
  if (count === 0) return '#161b22'
  if (count <= 2) return '#0e4429'
  if (count <= 5) return '#006d32'
  if (count <= 9) return '#26a641'
  return '#39d353'
}

onMounted(async () => {
  try {
    const to = new Date()
    const from = new Date()
    from.setDate(from.getDate() - 364)

    // Align start to Sunday
    const startDow = from.getDay()
    if (startDow !== 0) {
      from.setDate(from.getDate() - startDow)
    }

    const toStr = formatDate(to)
    // Add one day to `to` since backend uses exclusive end
    const toExclusive = new Date(to)
    toExclusive.setDate(toExclusive.getDate() + 1)
    const toExclusiveStr = formatDate(toExclusive)
    const fromStr = formatDate(from)

    const data = await CommitHeatmap(fromStr, toExclusiveStr, '')

    // Build sparse lookup
    const countMap = new Map<string, number>()
    if (data) {
      for (const d of data) {
        countMap.set(d.date, d.count)
      }
    }

    // Generate cells from `from` to `to`
    const allCells: typeof cells.value = []
    const monthMarkers: typeof months.value = []
    let lastMonth = -1
    let weekCol = 0

    const cursor = new Date(from)
    while (cursor <= to) {
      const dateStr = formatDate(cursor)
      const dow = cursor.getDay()

      if (dow === 0 && allCells.length > 0) {
        weekCol++
      }

      const currentMonth = cursor.getMonth()
      if (currentMonth !== lastMonth) {
        monthMarkers.push({ label: MONTH_NAMES[currentMonth], col: weekCol })
        lastMonth = currentMonth
      }

      allCells.push({
        date: dateStr,
        count: countMap.get(dateStr) || 0,
        dayOfWeek: dow,
      })

      cursor.setDate(cursor.getDate() + 1)
    }

    cells.value = allCells
    months.value = monthMarkers
  } catch (e: any) {
    error.value = e?.message || String(e)
  }
})
</script>

<template>
  <div class="heatmap-container">
    <h3>Commit Activity</h3>
    <div v-if="error" class="heatmap-error">{{ error }}</div>
    <div v-else-if="cells.length === 0" class="heatmap-loading">Loading...</div>
    <template v-else>
      <div class="heatmap-scroll">
        <div class="heatmap-wrapper">
          <!-- Day-of-week labels -->
          <div class="day-labels">
            <div v-for="(label, i) in DAY_LABELS" :key="i" class="day-label">{{ label }}</div>
          </div>
          <div class="grid-area">
            <!-- Month labels -->
            <div class="month-labels">
              <span
                v-for="(m, i) in months"
                :key="i"
                class="month-label"
                :style="{ gridColumn: m.col + 1 }"
              >{{ m.label }}</span>
            </div>
            <!-- Heatmap grid -->
            <div class="heatmap-grid">
              <div
                v-for="(cell, i) in cells"
                :key="i"
                class="heatmap-cell"
                :style="{
                  backgroundColor: colorForCount(cell.count),
                  gridRow: cell.dayOfWeek + 1,
                }"
                :title="`${cell.count} commit${cell.count === 1 ? '' : 's'} on ${cell.date}`"
              ></div>
            </div>
          </div>
        </div>
      </div>
      <!-- Legend -->
      <div class="legend">
        <span class="legend-label">Less</span>
        <div class="legend-cell" :style="{ backgroundColor: '#161b22' }"></div>
        <div class="legend-cell" :style="{ backgroundColor: '#0e4429' }"></div>
        <div class="legend-cell" :style="{ backgroundColor: '#006d32' }"></div>
        <div class="legend-cell" :style="{ backgroundColor: '#26a641' }"></div>
        <div class="legend-cell" :style="{ backgroundColor: '#39d353' }"></div>
        <span class="legend-label">More</span>
      </div>
    </template>
  </div>
</template>

<style scoped>
.heatmap-container {
  padding: 16px;
}

.heatmap-container h3 {
  margin: 0 0 12px 0;
  font-size: 16px;
  font-weight: 600;
  color: #c9d1d9;
}

.heatmap-error {
  color: #f85149;
  font-size: 14px;
}

.heatmap-loading {
  color: #8b949e;
  font-size: 14px;
}

.heatmap-scroll {
  overflow-x: auto;
}

.heatmap-wrapper {
  display: flex;
  gap: 4px;
}

.day-labels {
  display: grid;
  grid-template-rows: repeat(7, 13px);
  gap: 3px;
  padding-top: 20px; /* align with grid, below month labels */
}

.day-label {
  font-size: 10px;
  color: #8b949e;
  line-height: 13px;
  text-align: right;
  padding-right: 4px;
  width: 28px;
}

.grid-area {
  display: flex;
  flex-direction: column;
}

.month-labels {
  display: grid;
  grid-auto-columns: 13px;
  gap: 3px;
  height: 17px;
  margin-bottom: 3px;
}

.month-label {
  font-size: 10px;
  color: #8b949e;
  line-height: 17px;
}

.heatmap-grid {
  display: grid;
  grid-template-rows: repeat(7, 13px);
  grid-auto-flow: column;
  grid-auto-columns: 13px;
  gap: 3px;
}

.heatmap-cell {
  width: 13px;
  height: 13px;
  border-radius: 2px;
  outline: 1px solid rgba(27, 31, 35, 0.06);
}

.legend {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-top: 8px;
  justify-content: flex-end;
}

.legend-label {
  font-size: 11px;
  color: #8b949e;
}

.legend-cell {
  width: 13px;
  height: 13px;
  border-radius: 2px;
}
</style>
