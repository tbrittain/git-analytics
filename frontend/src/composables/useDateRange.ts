import { computed, ref, watch } from 'vue'

export type Preset = {
  label: string
  days: number | null // null = all time
}

export const defaultPresets: Preset[] = [
  { label: '30d', days: 30 },
  { label: '90d', days: 90 },
  { label: '6mo', days: 182 },
  { label: '1yr', days: 365 },
  { label: 'All', days: null },
]

export function formatDate(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

export function useDateRange(presets: Preset[] = defaultPresets, defaultIndex = 2) {
  const activePreset = ref<number | null>(defaultIndex)
  const customFrom = ref('')
  const customTo = ref('')

  const fromStr = computed(() => {
    if (activePreset.value !== null) {
      const preset = presets[activePreset.value]
      if (preset.days !== null) {
        const from = new Date()
        from.setDate(from.getDate() - preset.days)
        return formatDate(from)
      }
      return '1970-01-01'
    }
    return customFrom.value
  })

  const toStr = computed(() => {
    if (activePreset.value !== null) {
      const to = new Date()
      to.setDate(to.getDate() + 1) // exclusive end
      return formatDate(to)
    }
    return customTo.value
  })

  function setPreset(index: number) {
    activePreset.value = index
    customFrom.value = ''
    customTo.value = ''
  }

  // When custom dates change, deactivate preset
  watch([customFrom, customTo], ([from, to]) => {
    if (from && to) {
      activePreset.value = null
    }
  })

  return {
    presets,
    activePreset,
    customFrom,
    customTo,
    fromStr,
    toStr,
    setPreset,
  }
}
