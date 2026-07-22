<template>
  <div ref="root" class="app-select" :class="{ open }">
    <button
      ref="trigger"
      class="app-select-trigger"
      type="button"
      :aria-controls="menuId"
      :aria-expanded="open"
      aria-haspopup="listbox"
      @click="toggleMenu"
      @keydown="handleTriggerKeydown"
    >
      <span class="app-select-value">{{ selectedOption?.label || placeholder }}</span>
      <i class="mgc_down_small_line app-select-chevron" aria-hidden="true"></i>
    </button>

    <Transition name="select-menu">
      <div
        v-if="open"
        :id="menuId"
        class="app-select-menu"
        role="listbox"
        :aria-label="ariaLabel"
      >
        <button
          v-for="(option, index) in options"
          :key="option.value"
          class="app-select-option"
          :class="{ selected: option.value === modelValue }"
          :data-index="index"
          type="button"
          role="option"
          :aria-selected="option.value === modelValue"
          @click="selectOption(option, index)"
          @keydown.down.prevent="moveActive(1)"
          @keydown.up.prevent="moveActive(-1)"
          @keydown.home.prevent="focusIndex(0)"
          @keydown.end.prevent="focusIndex(options.length - 1)"
          @keydown.esc.prevent="closeMenu(true)"
        >
          <span>{{ option.label }}</span>
          <i v-if="option.value === modelValue" class="mgc_check_line" aria-hidden="true"></i>
        </button>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { computed, nextTick, onMounted, onUnmounted, ref, useId } from 'vue'

const props = defineProps({
  modelValue: { type: [String, Number], default: '' },
  options: { type: Array, required: true },
  placeholder: { type: String, default: '请选择' },
  ariaLabel: { type: String, default: '选择选项' },
})

const emit = defineEmits(['update:modelValue', 'change'])
const root = ref(null)
const trigger = ref(null)
const open = ref(false)
const activeIndex = ref(0)
const menuId = `app-select-${useId()}`

const selectedIndex = computed(() => Math.max(0, props.options.findIndex((option) => option.value === props.modelValue)))
const selectedOption = computed(() => props.options.find((option) => option.value === props.modelValue))

const focusActiveOption = () => {
  root.value?.querySelector(`[data-index="${activeIndex.value}"]`)?.focus()
}

const openMenu = (offset = 0) => {
  if (!props.options.length) return
  activeIndex.value = Math.min(props.options.length - 1, Math.max(0, selectedIndex.value + offset))
  open.value = true
  nextTick(focusActiveOption)
}

const closeMenu = (restoreFocus = false) => {
  open.value = false
  if (restoreFocus) nextTick(() => trigger.value?.focus())
}

const toggleMenu = () => {
  if (open.value) closeMenu()
  else openMenu()
}

const handleTriggerKeydown = (event) => {
  if (event.key === 'ArrowDown') {
    event.preventDefault()
    openMenu(0)
  } else if (event.key === 'ArrowUp') {
    event.preventDefault()
    openMenu(-1)
  } else if (event.key === 'Escape') {
    closeMenu()
  }
}

const focusIndex = (index) => {
  if (!props.options.length) return
  activeIndex.value = Math.min(props.options.length - 1, Math.max(0, index))
  nextTick(focusActiveOption)
}

const moveActive = (direction) => {
  const nextIndex = (activeIndex.value + direction + props.options.length) % props.options.length
  focusIndex(nextIndex)
}

const selectOption = (option, index) => {
  activeIndex.value = index
  if (option.value !== props.modelValue) {
    emit('update:modelValue', option.value)
    emit('change', option.value)
  }
  closeMenu(true)
}

const handleOutside = (event) => {
  if (!root.value?.contains(event.target)) closeMenu()
}

onMounted(() => document.addEventListener('pointerdown', handleOutside))
onUnmounted(() => document.removeEventListener('pointerdown', handleOutside))
</script>

<style scoped>
.app-select { position:relative; display:inline-block; min-width:0; }
.app-select-trigger { display:flex; align-items:center; justify-content:space-between; gap:.65rem; width:100%; min-height:2.55rem; padding:.5rem .62rem .5rem .75rem; border:1px solid #e8dfe2; border-radius:.72rem; color:#59636d; background:rgba(255,250,251,.96); box-shadow:0 5px 15px rgba(78,48,57,.035); font-size:.75rem; font-weight:700; text-align:left; transition:border-color .2s ease,box-shadow .2s ease,background .2s ease; }
.app-select-trigger:hover,.app-select.open .app-select-trigger { border-color:rgba(207,91,118,.48); background:#fff; box-shadow:0 0 0 3px rgba(207,91,118,.075); }
.app-select-value { min-width:0; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.app-select-chevron { flex:0 0 auto; color:#9a858b; font-size:1rem; transition:transform .2s ease; }
.app-select.open .app-select-chevron { transform:rotate(180deg); }
.app-select-menu { position:absolute; z-index:35; top:calc(100% + .45rem); right:0; min-width:max(100%,11rem); max-width:min(18rem,calc(100vw - 2rem)); max-height:17rem; overflow:auto; padding:.38rem; border:1px solid rgba(211,122,143,.16); border-radius:.9rem; background:rgba(255,255,255,.98); box-shadow:0 18px 46px rgba(83,50,60,.16),inset 0 1px 0 rgba(255,255,255,.8); backdrop-filter:blur(18px); }
.app-select-option { display:flex; align-items:center; justify-content:space-between; gap:1rem; width:100%; min-height:2.45rem; padding:.55rem .68rem; border-radius:.65rem; color:#65707a; font-size:.75rem; font-weight:650; text-align:left; transition:background .16s ease,color .16s ease; }
.app-select-option:hover,.app-select-option:focus-visible { color:#af4860; background:#fff3f5; outline:none; }
.app-select-option.selected { color:#b34b63; background:#fff0f3; }
.app-select-option i { color:#d25d77; font-size:.95rem; }
.dark .app-select-trigger { color:#e2e7ea; background:#272d35; border-color:#414955; box-shadow:none; }
.dark .app-select-trigger:hover,.dark .app-select.open .app-select-trigger { border-color:rgba(244,139,163,.45); background:#2c323b; box-shadow:0 0 0 3px rgba(244,139,163,.08); }
.dark .app-select-chevron { color:#b9aeb2; }
.dark .app-select-menu { background:rgba(35,40,48,.98); border-color:rgba(255,255,255,.08); box-shadow:0 20px 48px rgba(2,7,12,.4); }
.dark .app-select-option { color:#c8cfd4; }
.dark .app-select-option:hover,.dark .app-select-option:focus-visible,.dark .app-select-option.selected { color:#ff9bb0; background:#3b3037; }
.select-menu-enter-active,.select-menu-leave-active { transition:opacity .16s ease,transform .16s ease; transform-origin:top right; }
.select-menu-enter-from,.select-menu-leave-to { opacity:0; transform:translateY(-5px) scale(.985); }
@media(prefers-reduced-motion:reduce){.app-select-trigger,.app-select-chevron,.app-select-option,.select-menu-enter-active,.select-menu-leave-active{transition:none}}
</style>
