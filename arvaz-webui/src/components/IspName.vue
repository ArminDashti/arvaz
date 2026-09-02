<script setup lang="ts">
import { computed } from 'vue'
import { ispLogoSrc } from '@/lib/ispLogo'
import { emptyGridValue } from '@/lib/utils'

const props = defineProps<{
  name?: string
  logoKey?: string
}>()

const lookupFailed = computed(() => (props.name || '').trim().toLowerCase() === 'error')
const src = computed(() => (lookupFailed.value ? undefined : ispLogoSrc(props.name, props.logoKey)))
const label = computed(() => {
  if (lookupFailed.value) return 'ISP lookup failed'
  return props.name?.trim() || emptyGridValue
})
const tip = computed(() => {
  if (lookupFailed.value) return 'Lookup service returned an error for this IP'
  return props.name?.trim() || undefined
})
</script>

<template>
  <span class="isp-name" :title="tip">
    <img v-if="src" class="isp-name__logo" :src="src" alt="" />
    <span class="isp-name__text">{{ label }}</span>
  </span>
</template>

<style scoped>
.isp-name {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  min-width: 0;
  max-width: 28rem;
}

.isp-name__logo {
  width: 24px;
  height: 24px;
  flex-shrink: 0;
  object-fit: contain;
}

.isp-name__text {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
