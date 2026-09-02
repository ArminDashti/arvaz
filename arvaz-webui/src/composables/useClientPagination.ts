import { computed, ref, type Ref } from 'vue'

export const PAGE_SIZE = 50

export function useClientPagination<T>(rows: Ref<T[]>, pageSize = PAGE_SIZE) {
  const page = ref(1)

  const totalPages = computed(() => {
    const n = rows.value.length
    if (n <= 0) return 1
    return Math.max(1, Math.ceil(n / pageSize))
  })

  const pageRows = computed(() => {
    const p = Math.min(page.value, totalPages.value)
    const start = (p - 1) * pageSize
    return rows.value.slice(start, start + pageSize)
  })

  function resetPage() {
    page.value = 1
  }

  function prevPage() {
    if (page.value > 1) page.value -= 1
  }

  function nextPage() {
    if (page.value < totalPages.value) page.value += 1
  }

  return { page, pageRows, pageSize, prevPage, nextPage, resetPage, totalPages }
}
