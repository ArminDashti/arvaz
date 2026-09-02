<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    countryCode?: string
    countryName?: string
    size?: 'sm' | 'lg'
  }>(),
  { size: 'sm' },
)

const code = computed(() => {
  const cc = props.countryCode?.trim().toLowerCase()
  if (!cc || cc.length !== 2) return ''
  return cc
})

const pixelWidth = computed(() => (props.size === 'lg' ? 80 : 40))

const src = computed(() => {
  if (!code.value) return ''
  return `https://flagcdn.com/w${pixelWidth.value}/${code.value}.png`
})

const alt = computed(() => props.countryName?.trim() || (code.value ? code.value.toUpperCase() : 'Country'))
</script>

<template>
  <img
    v-if="src"
    class="country-flag"
    :class="size === 'lg' ? 'country-flag--lg' : 'country-flag--sm'"
    :src="src"
    :alt="alt"
    :title="alt"
  />
</template>

<style scoped>
.country-flag {
  display: inline-block;
  object-fit: cover;
  flex-shrink: 0;
  border-radius: 2px;
}

.country-flag--sm {
  width: 28px;
  height: 21px;
}

.country-flag--lg {
  width: 80px;
  height: 60px;
}
</style>
