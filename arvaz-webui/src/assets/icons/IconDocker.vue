<script setup lang="ts">
defineProps<{ size?: number | string }>()
</script>

<template>
  <svg
    xmlns="http://www.w3.org/2000/svg"
    :width="size ?? 18"
    :height="size ?? 18"
    viewBox="0 0 24 24"
    fill="#2496ED"
    aria-hidden="true"
  >
    <!-- Docker whale mark (simplified) -->
    <path
      d="M4.2 11.1h1.5v1.5H4.2zm1.9 0h1.5v1.5H6.1zm1.9 0h1.5v1.5H8zm1.9 0h1.5v1.5h-1.5zm-3.8-1.9h1.5v1.5H8zm1.9 0h1.5v1.5H9.9zm1.9 0h1.5v1.5h-1.5zm1.9 0h1.5v1.5H13.7zm-5.7-1.9h1.5v1.5H9.9zm1.9 0h1.5v1.5h-1.5zM21.4 12c-.3-.2-1.9-.3-2.9-.2-.1-.7-.5-1.3-1.1-1.8l-.2-.2-.2.2c-.4.5-.6 1.2-.5 1.9-.5.1-1 .4-1.3.8H2.8c-.2 1.3.1 2.6.9 3.7 1 1.3 2.5 2 4.2 2.1 2 .1 3.9-.7 5.3-2.1.9.4 1.9.5 2.9.3 1.2-.2 2.2-1 2.7-2.1.5 0 1.5 0 2-.7.1-.1.3-.5.1-.9z"
    />
  </svg>
</template>
