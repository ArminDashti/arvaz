<script setup lang="ts">
defineProps<{ size?: number | string }>()
</script>

<template>
  <svg
    xmlns="http://www.w3.org/2000/svg"
    :width="size ?? 18"
    :height="size ?? 18"
    viewBox="0 0 24 24"
    fill="none"
    aria-hidden="true"
  >
    <!-- Speed gauge / network meter for Iperf -->
    <circle cx="12" cy="14" r="8" stroke="#F0A500" stroke-width="2" fill="none" />
    <path d="M12 14 L17 9" stroke="#F0A500" stroke-width="2" stroke-linecap="round" />
    <circle cx="12" cy="14" r="1.5" fill="#F0A500" />
    <path d="M6.5 18.5 A8 8 0 0 1 17.5 18.5" stroke="#F0A500" stroke-width="2" fill="none" opacity="0.35" />
  </svg>
</template>
