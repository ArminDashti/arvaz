<script setup lang="ts">
import { onMounted, ref } from 'vue'

const mailtoHref = ref('#')
const photoFailed = ref(false)
const photoSrc = '/about-me/armin.png'

const parts = {
  user: 'arminonline71',
  at: '@',
  domain: 'gmail',
  dot: '.',
  tld: 'com',
}

onMounted(() => {
  mailtoHref.value = `mailto:${parts.user}${parts.at}${parts.domain}${parts.dot}${parts.tld}`
})
</script>

<template>
  <div class="mx-auto flex max-w-3xl flex-col gap-5 pb-8">
    <header class="space-y-2">
      <h1 class="text-3xl font-semibold tracking-tight sm:text-4xl">About Me</h1>
      <p class="text-sm text-muted-foreground">Armin Dashti — vibe coder, conductor of craft.</p>
    </header>

    <div class="grid gap-4 sm:grid-cols-[160px_1fr] sm:items-start">
      <div class="overflow-hidden rounded-2xl border bg-muted/30">
        <img
          v-show="!photoFailed"
          :src="photoSrc"
          alt="Armin Dashti"
          class="aspect-[4/5] w-full object-cover"
          @error="photoFailed = true"
        />
        <div
          v-if="photoFailed"
          class="flex aspect-[4/5] items-center justify-center px-3 text-center text-xs text-muted-foreground"
        >
          Photo placeholder — add public/about-me/armin.png
        </div>
      </div>

      <div class="space-y-4 text-[15px] leading-relaxed">
        <p>
          Armin Dashti is a software engineer and vibe coder—a conductor of craft more than a solitary
          typist. As a vibe coder, he shapes intent with clarity and constraint, then lets Cursor’s
          agents compose the implementation; he reviews, refines, and ships what remains worthy.
        </p>
        <p>
          This application was not stitched by hurried hands alone. It was written by AI
          agents—specifically Cursor—guided by Armin’s intent, then tempered by his review until it
          was fit to ship.
        </p>
      </div>
    </div>

    <section class="rounded-2xl border bg-card/60 p-4">
      <h2 class="mb-2 text-sm font-semibold uppercase tracking-wide text-muted-foreground">Interests</h2>
      <p class="text-sm leading-relaxed">
        Beyond the terminal, his curiosities roam widely: cars, movies, vibe coding, coding, politics,
        military, jet fighters, aircraft.
      </p>
    </section>

    <p class="text-sm">
      Contact:
      <a :href="mailtoHref" class="underline-offset-2 hover:underline">
        {{ parts.user }} [at] {{ parts.domain }} [dot] {{ parts.tld }}
      </a>
    </p>

    <div class="space-y-3">
      <blockquote class="border-l-2 border-primary pl-4 text-sm italic leading-relaxed text-foreground/90">
        “Beware the quiet man. For while others speak, he watches. While others act, he plans. And when
        they finally rest, he strikes.” — Anonymous
      </blockquote>
      <blockquote
        dir="rtl"
        lang="fa"
        class="border-r-2 border-primary pr-4 text-sm italic leading-relaxed text-foreground/90"
      >
        «از مرد خاموش برحذر باش. زیرا در حالی که دیگران سخن می‌گویند، او نظاره می‌کند. در حالی که دیگران
        عمل می‌کنند، او برنامه می‌ریزد. و هنگامی که سرانجام آرام می‌گیرند، او ضربه می‌زند.» — ناشناس
      </blockquote>
    </div>
  </div>
</template>
