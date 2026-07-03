<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { Menu, X, PanelLeft } from 'lucide-vue-next';
import Sidebar from './Sidebar.vue';
import Settings from './Settings.vue';
import Editor from './Editor.vue';
import Preview from './Preview.vue';
import { currentDiagram, showSettings, isSidebarOpen, toastMessage, config } from './store';
import * as App from '../wailsjs/go/main/App.js';
import { WindowToggleMaximise } from '../wailsjs/runtime/runtime.js';

let hideToastTimeout: any = null;
const setToast = (msg: string) => {
    toastMessage.value = msg;
    if (hideToastTimeout) clearTimeout(hideToastTimeout);
    hideToastTimeout = setTimeout(() => { toastMessage.value = ''; }, 4000);
};

onMounted(async () => {
    // Load config from localStorage
    const saved = localStorage.getItem('smart-mermaid-config');
    if (saved) {
        const parsed = JSON.parse(saved);
        config.baseUrl = parsed.baseUrl || '';
        config.apiKey = parsed.apiKey || '';
        config.githubToken = parsed.githubToken || '';
        config.githubRepo = parsed.githubRepo || '';
    }
});

// Watch for toast changes
import { watch } from 'vue';
watch(toastMessage, (val) => {
    if (val) setToast(val);
});
</script>

<template>
  <main class="w-full h-full flex bg-[#0e0e11] text-slate-300 font-sans overflow-hidden select-none">
    
    <!-- Global Sidebar Toggle (Antigravity style) -->
    <div style="--wails-draggable: no-drag" class="absolute top-[8px] left-[76px] z-50">
        <button @click="isSidebarOpen = !isSidebarOpen" class="p-1 rounded text-[#a1a1aa] hover:text-white hover:bg-white/10 transition-colors flex items-center justify-center" title="Toggle Sidebar">
            <PanelLeft :size="16" stroke-width="2.2" />
        </button>
    </div>

    <!-- Sidebar with smooth width transition -->
    <div :class="isSidebarOpen ? 'w-64 border-r border-white/5' : 'w-0 border-r-0'" class="shrink-0 transition-all duration-300 ease-[cubic-bezier(0.2,0,0,1)] z-40 relative overflow-hidden bg-[#131314]">
        <div class="w-64 h-full">
            <Sidebar />
        </div>
    </div>

    <!-- Main Content Area -->
    <div class="flex-1 relative flex flex-col overflow-hidden bg-[#0e0e11]">
        
        <!-- Draggable Title Bar Area for entire app -->
        <div style="--wails-draggable: drag" class="absolute top-0 left-0 w-full h-10 z-10" @dblclick="WindowToggleMaximise"></div>

        <template v-if="currentDiagram">
            <!-- Full size preview -->
            <div class="absolute inset-0">
                <Preview />
            </div>

            <!-- Floating UI Overlay (Compact) -->
            <div style="--wails-draggable: no-drag" class="absolute bottom-8 left-1/2 -translate-x-1/2 w-full max-w-4xl px-4 z-30">
                <Editor />
            </div>
        </template>
        <template v-else>
            <div class="w-full h-full flex flex-col items-center justify-center text-slate-500 z-10 relative">
                <div class="w-16 h-16 mb-4 rounded-full bg-[#21262d] flex items-center justify-center border border-[#30363d] shadow-lg">
                    <span class="text-2xl text-blue-500 font-bold">M</span>
                </div>
                <h1 class="text-lg font-semibold text-slate-300 mb-1">Smart Mermaid</h1>
                <p class="text-xs text-slate-500">Select or create a diagram to begin</p>
            </div>
        </template>
    </div>

    <!-- Modals -->
    <Settings />

    <!-- Toast Notification -->
    <transition name="slide-up">
        <div v-if="toastMessage" class="absolute bottom-6 right-6 bg-[#1f6feb] text-white px-4 py-2 rounded-lg shadow-xl text-[13px] font-medium z-[9999] flex items-center gap-2 max-w-sm">
            <span>{{ toastMessage }}</span>
            <button @click="toastMessage = ''" class="p-0.5 hover:bg-white/20 rounded"><X :size="14"/></button>
        </div>
    </transition>
  </main>
</template>
