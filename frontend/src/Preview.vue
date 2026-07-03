<script setup lang="ts">
import { ref, watch, onMounted, nextTick } from 'vue';
import mermaid from 'mermaid';
import { Maximize2, Minimize2, ZoomIn, ZoomOut, Download, Navigation, Expand, Minimize } from 'lucide-vue-next';
import { currentDiagram, isSidebarOpen, isFullscreen, zoom } from './store';
import { WindowFullscreen, WindowUnfullscreen } from '../wailsjs/runtime/runtime.js';

const toggleFullscreen = () => {
    isFullscreen.value = !isFullscreen.value;
    if (isFullscreen.value) {
        WindowFullscreen();
    } else {
        WindowUnfullscreen();
    }
};

const container = ref<HTMLElement | null>(null);
const panX = ref(0);
const panY = ref(0);
const isDragging = ref(false);

let startX = 0;
let startY = 0;

const startPan = (e: MouseEvent) => {
    isDragging.value = true;
    startX = e.clientX - panX.value;
    startY = e.clientY - panY.value;
};

const doPan = (e: MouseEvent) => {
    if (!isDragging.value) return;
    panX.value = e.clientX - startX;
    panY.value = e.clientY - startY;
};

const endPan = () => {
    isDragging.value = false;
};

const handleWheel = (e: WheelEvent) => {
    if (e.ctrlKey || e.metaKey) {
        // Pinch to zoom or Ctrl+Scroll
        const delta = e.deltaY * -0.01;
        zoom.value = Math.min(Math.max(0.1, zoom.value * (1 + delta)), 300);
    } else {
        // Pan
        panX.value -= e.deltaX;
        panY.value -= e.deltaY;
    }
};

const resetView = () => {
    zoom.value = 1;
    panX.value = 0;
    panY.value = 0;
};

onMounted(() => {
    mermaid.initialize({ 
        startOnLoad: false, 
        theme: 'dark',
        securityLevel: 'loose'
    });
});

const renderDiagram = async () => {
    if (!container.value || !currentDiagram.value?.content) return;
    
    try {
        // Clear container first so no duplicate IDs exist in the DOM
        container.value.innerHTML = '';
        
        // Use dynamic ID to prevent any DOM clashing in Mermaid
        const id = `mermaid-${Date.now()}`;
        
        // In Mermaid v11, parse() returns boolean or throws
        await mermaid.parse(currentDiagram.value.content);
        
        const { svg } = await mermaid.render(id, currentDiagram.value.content);
        container.value.innerHTML = svg;
    } catch (e: any) {
        // Syntax error in mermaid code is expected while typing
        // Strip the error message to be cleaner
        let errMsg = e.message || String(e);
        // Sometimes the error starts with "Error: " or "Parse error". We can just show it.
        container.value.innerHTML = `
            <div class="flex flex-col items-center justify-center w-full max-w-2xl mx-auto mt-10">
                <div class="bg-red-500/10 border border-red-500/20 rounded-xl p-5 w-full shadow-lg">
                    <div class="flex items-center gap-2 text-red-400 mb-3">
                        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z"/><path d="M12 9v4"/><path d="M12 17h.01"/></svg>
                        <span class="font-semibold text-[13px]">Mermaid Syntax Error</span>
                    </div>
                    <div class="text-red-300/80 font-mono text-[11px] whitespace-pre-wrap leading-relaxed">${errMsg}</div>
                    <div class="mt-4 pt-3 border-t border-red-500/10 text-[11px] text-red-400/60">
                        Please check your Raw Mermaid Code. The diagram will render automatically once the syntax is fixed.
                    </div>
                </div>
            </div>`;
        
        // Clean up any bombs Mermaid might have dropped
        document.querySelectorAll('svg[id^="dmermaid-"]').forEach(el => el.remove());
        const errorNodes = document.querySelectorAll(`[id^="error-"]`);
        errorNodes.forEach(el => el.remove());
    }
};

watch(() => currentDiagram.value?.content, () => {
    nextTick(() => {
        renderDiagram();
    });
}, { immediate: true });

const downloadSvg = () => {
    if (!container.value) return;
    const svgEl = container.value.querySelector('svg');
    if (!svgEl) return;
    
    const svgData = new XMLSerializer().serializeToString(svgEl);
    const blob = new Blob([svgData], { type: 'image/svg+xml;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = `${currentDiagram.value?.name || 'diagram'}.svg`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
};
</script>

<template>
<div 
    class="h-full w-full bg-[#0a0c10] overflow-hidden flex items-center justify-center relative select-none"
    :class="isDragging ? 'cursor-grabbing' : 'cursor-grab'"
    @mousedown="startPan"
    @mousemove="doPan"
    @mouseup="endPan"
    @mouseleave="endPan"
    @wheel.prevent="handleWheel"
>
    
    <!-- Pan Container (Instant) -->
    <div class="flex items-center justify-center origin-center" :style="`transform: translate(${panX}px, ${panY}px)`">
        <!-- Zoom Container (Instant for trackpad) -->
        <div 
            ref="container" 
            class="origin-center flex items-center justify-center mermaid-container" 
            :style="`transform: scale(${zoom})`"
        >
            <div class="text-slate-500 animate-pulse text-[12px]">Rendering...</div>
        </div>
    </div>
</div>
</template>

<style scoped>
/* Prevent Mermaid from shrinking large diagrams to fit the screen */
:deep(svg) {
    max-width: none !important;
    height: auto !important;
}
</style>
