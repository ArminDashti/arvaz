<script setup lang="ts">
defineProps<{ size?: number | string }>()
</script>

<template>
  <svg
    xmlns="http://www.w3.org/2000/svg"
    :width="size ?? 18"
    :height="size ?? 18"
    viewBox="0 0 24 24"
    fill="#1AB394"
    aria-hidden="true"
  >
    <!-- Windscribe-style shield/W mark -->
    <path
      d="M12 2 4 5.5v5.2c0 5.1 3.4 9.8 8 11.3 4.6-1.5 8-6.2 8-11.3V5.5L12 2zm0 3.1 5.5 2.4v3.2c0 3.7-2.3 7.1-5.5 8.3-3.2-1.2-5.5-4.6-5.5-8.3V7.5L12 5.1zm-2.2 4.2 2.2 5.4 2.2-5.4h1.7L12.8 18h-1.6L7.9 9.3h1.9z"
    />
  </svg>
</template>
