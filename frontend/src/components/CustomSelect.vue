<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue';
import { ChevronDown, Check, Search } from 'lucide-vue-next';

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
const searchInput = ref<HTMLInputElement | null>(null);
const searchQuery = ref('');

const filteredOptions = computed(() => {
  if (!searchQuery.value) return props.options;
  return props.options.filter(opt => opt.toLowerCase().includes(searchQuery.value.toLowerCase()));
});

const toggleDropdown = () => {
  isOpen.value = !isOpen.value;
  if (isOpen.value) {
    searchQuery.value = '';
    nextTick(() => {
      searchInput.value?.focus();
    });
  }
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
      class="w-full bg-black/5 dark:bg-white/5 border border-slate-200 dark:border-white/5 hover:border-slate-300 dark:hover:border-white/20 hover:bg-black/10 dark:hover:bg-white/10 rounded-xl px-3 py-2 text-left text-[12px] text-slate-800 dark:text-white focus:outline-none transition-all flex items-center justify-between"
      type="button"
    >
      <span class="truncate" :class="{ 'text-slate-500 dark:text-[#a1a1aa]': !modelValue }">
        {{ modelValue || placeholder }}
      </span>
      <ChevronDown :size="14" class="text-slate-500 dark:text-[#a1a1aa] transition-transform duration-300" :class="{ 'rotate-180': isOpen }" />
    </button>

    <!-- Dropdown Menu -->
    <transition name="pop-scale">
      <div 
        v-if="isOpen" 
        class="absolute left-0 w-full bg-white/95 dark:bg-[#18181a]/95 backdrop-blur-3xl border border-slate-200 dark:border-white/10 rounded-xl shadow-2xl z-50 overflow-hidden py-1 flex flex-col"
        :class="direction === 'up' ? 'bottom-full mb-2' : 'top-full mt-2'"
      >
        <div v-if="options.length > 3" class="px-2 pb-1 border-b border-slate-200 dark:border-white/5 shrink-0 pt-1">
          <div class="relative flex items-center">
             <Search :size="12" class="absolute left-3 text-slate-400 dark:text-[#a1a1aa]" />
             <input 
               ref="searchInput"
               type="text"
               v-model="searchQuery"
               placeholder="Search..."
               class="w-full bg-slate-100 dark:bg-black/20 border border-slate-200 dark:border-white/5 rounded-lg pl-8 pr-3 py-1.5 text-[12px] text-slate-800 dark:text-white focus:outline-none focus:border-slate-300 dark:focus:border-white/20 transition-all placeholder:text-slate-400 dark:placeholder:text-[#a1a1aa]/50"
             >
          </div>
        </div>

        <div class="max-h-60 overflow-y-auto custom-scrollbar">
          <div v-if="filteredOptions.length === 0" class="px-3 py-2 text-[12px] text-slate-500 dark:text-[#a1a1aa] text-center">
            No options found
          </div>
          <button 
            v-for="option in filteredOptions" 
            :key="option"
            @click="selectOption(option)"
            class="w-full text-left px-3 py-2 text-[12px] flex items-center justify-between hover:bg-slate-100 dark:hover:bg-white/10 transition-colors"
            :class="modelValue === option ? 'text-blue-600 dark:text-white font-medium bg-blue-50 dark:bg-white/10' : 'text-slate-600 dark:text-[#a1a1aa]'"
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
