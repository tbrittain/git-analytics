import { type Ref, ref, watch } from 'vue'

const STORAGE_PREFIX = 'exclude-patterns:'

function load(repoPath: string): string[] {
  if (!repoPath) return []
  try {
    const raw = localStorage.getItem(STORAGE_PREFIX + repoPath)
    return raw ? JSON.parse(raw) : []
  } catch {
    return []
  }
}

function save(repoPath: string, patterns: string[]) {
  if (!repoPath) return
  localStorage.setItem(STORAGE_PREFIX + repoPath, JSON.stringify(patterns))
}

export function useExcludePatterns(repoPath: Ref<string>) {
  const patterns = ref<string[]>(load(repoPath.value))

  watch(repoPath, (path) => {
    patterns.value = load(path)
  })

  function addPattern(p: string) {
    const trimmed = p.trim()
    if (!trimmed || patterns.value.includes(trimmed)) return
    patterns.value = [...patterns.value, trimmed]
    save(repoPath.value, patterns.value)
  }

  function removePattern(p: string) {
    patterns.value = patterns.value.filter((x) => x !== p)
    save(repoPath.value, patterns.value)
  }

  return { patterns, addPattern, removePattern }
}
