<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import { ChevronDown, Check } from 'lucide-vue-next';

const props = defineProps({
  modelValue: {
    type: String,
    default: ''
  },
  options: {
    type: Array as () => string[],
    default: () => []
  },
  placeholder: {
    type: String,
    default: 'Select an option'
  },
  direction: {
    type: String,
    default: 'down' // 'up' or 'down'
  }
});

const emit = defineEmits(['update:modelValue', 'change']);

const isOpen = ref(false);
const dropdownRef = ref<HTMLElement | null>(null);

const toggleDropdown = () => {
  isOpen.value = !isOpen.value;
};

const selectOption = (option: string) => {
  emit('update:modelValue', option);
  emit('change', option);
  isOpen.value = false;
};

const closeOnOutsideClick = (e: MouseEvent) => {
  if (dropdownRef.value && !dropdownRef.value.contains(e.target as Node)) {
    isOpen.value = false;
  }
};

onMounted(() => {
  document.addEventListener('click', closeOnOutsideClick);
});

onUnmounted(() => {
  document.removeEventListener('click', closeOnOutsideClick);
});
</script>

<template>
  <div class="relative" ref="dropdownRef">
    <!-- Trigger Button -->
    <button 
      @click="toggleDropdown" 
      class="w-full bg-white/5 border border-white/5 hover:border-white/20 hover:bg-white/10 rounded-xl px-3 py-2 text-left text-[12px] text-white focus:outline-none transition-all flex items-center justify-between"
      type="button"
    >
      <span class="truncate" :class="{ 'text-[#a1a1aa]': !modelValue }">
        {{ modelValue || placeholder }}
      </span>
      <ChevronDown :size="14" class="text-[#a1a1aa] transition-transform duration-300" :class="{ 'rotate-180': isOpen }" />
    </button>

    <!-- Dropdown Menu -->
    <transition name="pop-scale">
      <div 
        v-if="isOpen" 
        class="absolute left-0 w-full bg-[#18181a]/95 backdrop-blur-3xl border border-white/10 rounded-xl shadow-2xl z-50 overflow-hidden py-1"
        :class="direction === 'up' ? 'bottom-full mb-2' : 'top-full mt-2'"
      >
        <div class="max-h-60 overflow-y-auto custom-scrollbar">
          <div v-if="options.length === 0" class="px-3 py-2 text-[12px] text-[#a1a1aa] text-center">
            No options available
          </div>
          <button 
            v-for="option in options" 
            :key="option"
            @click="selectOption(option)"
            class="w-full text-left px-3 py-2 text-[12px] flex items-center justify-between hover:bg-white/10 transition-colors"
            :class="modelValue === option ? 'text-white font-medium bg-white/10' : 'text-[#a1a1aa]'"
            type="button"
          >
            <span class="truncate pr-2">{{ option }}</span>
            <Check v-if="modelValue === option" :size="14" />
          </button>
        </div>
      </div>
    </transition>
  </div>
</template>
