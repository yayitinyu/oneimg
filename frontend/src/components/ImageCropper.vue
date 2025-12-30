<template>
  <div class="fixed inset-0 z-[60] flex items-center justify-center bg-black/70 backdrop-blur-sm transition-opacity" v-if="visible">
    <div class="bg-white dark:bg-dark-300 rounded-xl shadow-2xl w-full max-w-2xl overflow-hidden flex flex-col max-h-[90vh] mx-4 md:mx-0">
      <!-- Header -->
      <div class="px-4 py-3 md:px-6 md:py-4 border-b border-gray-100 dark:border-dark-100 flex justify-between items-center">
        <h3 class="text-lg font-bold text-gray-800 dark:text-white">裁剪图片</h3>
        <button @click="handleCancel" class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 transition-colors">
          <i class="ri-close-line text-2xl"></i>
        </button>
      </div>

      <!-- Content -->
      <div class="p-4 md:p-6 flex-1 overflow-hidden bg-gray-50 dark:bg-dark-400 relative">
        <div class="h-[50vh] md:h-[400px] w-full">
            <img ref="imageRef" :src="imageSrc" alt="Source Image" class="max-w-full block" />
        </div>
      </div>

      <!-- Footer -->
      <div class="px-4 py-3 md:px-6 md:py-4 border-t border-gray-100 dark:border-dark-100 flex flex-col-reverse md:flex-row justify-end gap-3 bg-white dark:bg-dark-300">
        <button @click="handleCancel" class="w-full md:w-auto px-5 py-2 rounded-lg text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-dark-200 transition-colors font-medium">
          取消
        </button>
        <button @click="handleConfirm" class="w-full md:w-auto px-5 py-2 rounded-lg bg-pink-500 hover:bg-pink-600 text-white shadow-lg shadow-pink-500/30 transition-all font-medium flex items-center justify-center gap-2">
          <i class="ri-check-line"></i>
          确认裁剪
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue';
import Cropper from 'cropperjs';
import 'cropperjs/dist/cropper.css';

const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  },
  imageSrc: {
    type: String,
    required: true
  },
  aspectRatio: {
    type: Number,
    default: 1
  }
});

const emit = defineEmits(['update:visible', 'info', 'cropped']);

const imageRef = ref(null);
let cropper = null;

const initCropper = () => {
    if (cropper) {
        cropper.destroy();
    }
    if (imageRef.value) {
        cropper = new Cropper(imageRef.value, {
            aspectRatio: props.aspectRatio,
            viewMode: 1,
            dragMode: 'move',
            autoCropArea: 1,
            restore: false,
            guides: true,
            center: true,
            highlight: false,
            cropBoxMovable: true,
            cropBoxResizable: true,
            toggleDragModeOnDblclick: false,
            background: false, // Turn off default checkboard
        });
    }
};

watch(() => props.visible, (newVal) => {
    if (newVal) {
        nextTick(() => {
            initCropper();
        });
    } else {
        if (cropper) {
            cropper.destroy();
            cropper = null;
        }
    }
});

const handleCancel = () => {
  emit('update:visible', false);
};

const handleConfirm = () => {
  if (!cropper) return;
  
  cropper.getCroppedCanvas({
      width: 400, // Reasonable default for avatar/logo
      height: 400,
      imageSmoothingQuality: 'high',
  }).toBlob((blob) => {
      emit('cropped', blob);
      emit('update:visible', false);
  }, 'image/png');
};

onUnmounted(() => {
    if (cropper) {
        cropper.destroy();
    }
});
</script>

<style scoped>
/* Override some cropper styles to match theme if needed */
:deep(.cropper-view-box),
:deep(.cropper-face) {
  border-radius: 0; 
}
/* If circle crop is needed visually, we can add border-radius: 50% to view-box but actual crop is rect */
</style>
