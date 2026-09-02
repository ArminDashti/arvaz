<script setup lang="ts">
defineProps<{
  page: number
  totalPages: number
  totalRows: number
  pageSize: number
}>()

const emit = defineEmits<{
  prev: []
  next: []
}>()
</script>

<template>
  <div v-if="totalRows > pageSize" class="mt-3 flex flex-wrap items-center justify-between gap-2 text-sm">
    <p class="text-muted-foreground">
      Page {{ page }} of {{ totalPages }} ({{ totalRows }} rows)
    </p>
    <div class="flex gap-2">
      <button type="button" class="field-button" :disabled="page <= 1" @click="emit('prev')">
        Prev
      </button>
      <button type="button" class="field-button" :disabled="page >= totalPages" @click="emit('next')">
        Next
      </button>
    </div>
  </div>
</template>
