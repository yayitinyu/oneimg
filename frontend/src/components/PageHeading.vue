<template>
  <header class="content-heading" :class="{ 'has-action': $slots.action }">
    <div class="content-heading-copy">
      <p v-if="eyebrow" class="content-eyebrow">{{ eyebrow }}</p>
      <h1>{{ title }}</h1>
      <p v-if="description" class="content-description">{{ description }}</p>
    </div>
    <div v-if="$slots.action" class="content-heading-action">
      <slot name="action" />
    </div>
  </header>
</template>

<script setup>
defineProps({
  eyebrow: { type: String, default: '' },
  title: { type: String, required: true },
  description: { type: String, default: '' },
})
</script>

<style scoped>
.content-heading {
  display: flex;
  min-height: 6.15rem;
  align-items: flex-end;
  justify-content: space-between;
  gap: 2rem;
  margin-bottom: 1.4rem;
}

.content-heading-copy {
  min-width: 0;
}

.content-eyebrow {
  color: #c65a73;
  font-size: .7rem;
  font-weight: 800;
  letter-spacing: .15em;
  line-height: 1.35;
  text-transform: uppercase;
}

h1 {
  margin: .12rem 0 .35rem;
  color: #303a44;
  font-size: clamp(2rem, 4vw, 3.4rem);
  font-weight: 700;
  letter-spacing: -.055em;
  line-height: 1;
  text-wrap: balance;
}

.dark h1 {
  color: #edf1f4;
}

.content-description {
  max-width: 42rem;
  color: #858f99;
  font-size: .84rem;
  line-height: 1.55;
  text-wrap: pretty;
}

.content-heading-action {
  flex: 0 0 auto;
}

@media (max-width: 640px) {
  .content-heading {
    position: relative;
    display: block;
    min-height: 0;
    margin-bottom: 1.15rem;
  }

  .content-heading.has-action .content-eyebrow,
  .content-heading.has-action h1 {
    padding-right: 7.5rem;
  }

  h1 {
    font-size: 2.25rem;
  }

  .content-description {
    font-size: .82rem;
  }

  .content-heading-action {
    position: absolute;
    top: 0;
    right: 0;
  }
}
</style>
