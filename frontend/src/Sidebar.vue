<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { Plus, Settings, Cloud, CloudOff, Trash2 } from 'lucide-vue-next';
import { diagrams, currentDiagram, showSettings, isSidebarOpen, config, toastMessage } from './store';
import * as App from '../wailsjs/go/main/App.js';
import { WindowToggleMaximise } from '../wailsjs/runtime/runtime.js';

const isTestingConnection = ref(false);

const loadDiagrams = async () => {
    try {
        const list = await App.GetDiagrams();
        diagrams.value = list || [];
    } catch (e) {
        console.error(e);
    }
};

onMounted(() => {
    loadDiagrams();
});

const createNew = async () => {
    if (!config.baseUrl) {
        toastMessage.value = 'Please configure AI Base URL in Settings first.';
        showSettings.value = true;
        return;
    }

    isTestingConnection.value = true;
    try {
        const res = await App.TestAIConnection(config.baseUrl, config.apiKey);
        if (res !== 'success') throw new Error('API Error');
    } catch (e) {
        isTestingConnection.value = false;
        toastMessage.value = 'Failed to connect to AI server. Error: ' + e;
        showSettings.value = true;
        return;
    }
    isTestingConnection.value = false;

    currentDiagram.value = {
        id: '',
        name: 'New Diagram',
        content: 'graph TD\n    A-->B;',
        synced: false,
        updatedAt: ''
    };
};

const selectDiagram = (d: any) => {
    currentDiagram.value = d;
};

const deleteDiagram = async (e: Event, id: string) => {
    e.stopPropagation();
    await App.DeleteDiagram(id);
    if (currentDiagram.value && currentDiagram.value.id === id) {
        currentDiagram.value = null;
    }
    toastMessage.value = 'Diagram deleted';
    loadDiagrams();
};
</script>

<template>
<div class="h-full bg-[#131314] flex flex-col text-[#a1a1aa] border-r border-white/5">
    <!-- Header (Draggable for macOS) -->
    <div style="--wails-draggable: drag" class="h-[40px] pr-4 pl-[110px] border-b border-white/5 flex items-center justify-between shrink-0" @dblclick="WindowToggleMaximise">
        <h1 class="font-semibold text-[13px] text-[#a1a1aa] flex items-center gap-2">
            Diagrams
        </h1>
        <div style="--wails-draggable: no-drag" class="flex items-center gap-1">
            <button @click="showSettings = true" class="text-[#a1a1aa] hover:text-white p-1 rounded hover:bg-white/10 transition-colors" title="Settings">
                <Settings :size="14" />
            </button>
        </div>
    </div>

    <!-- Actions -->
    <div class="p-3 border-b border-white/5 shrink-0">
        <button @click="createNew" :disabled="isTestingConnection" class="w-full bg-white/5 hover:bg-white/10 text-white rounded-md py-1.5 px-2 text-[12px] font-medium flex items-center justify-center gap-1.5 transition-colors border border-white/5 disabled:opacity-50 disabled:cursor-not-allowed">
            <template v-if="isTestingConnection">
                <div class="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin"></div> Checking...
            </template>
            <template v-else>
                <Plus :size="14" /> New Diagram
            </template>
        </button>
    </div>

    <!-- List -->
    <div class="flex-1 overflow-y-auto p-3 space-y-1 custom-scrollbar">
        <div v-for="diagram in diagrams" :key="diagram.id"
            class="group w-full text-left px-3 py-2 rounded-md text-[12px] font-medium flex items-center justify-between cursor-pointer transition-colors"
            :class="currentDiagram?.id === diagram.id ? 'bg-[#1f6feb]/15 text-blue-400' : 'text-[#a1a1aa] hover:bg-white/5 hover:text-white'"
            @click="selectDiagram(diagram)"
        >
            <div class="truncate flex-1 pr-2">
                {{ diagram.name || 'Untitled' }}
            </div>
            <div class="flex items-center gap-1.5 opacity-0 group-hover:opacity-100 transition-opacity shrink-0">
                <button class="text-[#a1a1aa] hover:text-red-400 p-0.5" @click="deleteDiagram($event, diagram.id)">
                    <Trash2 :size="13" />
                </button>
                <Cloud v-if="diagram.synced" :size="13" class="text-[#238636]" title="Synced" />
                <CloudOff v-else :size="13" class="text-[#a1a1aa]/50" title="Local" />
            </div>
        </div>
        
        <div v-if="diagrams.length === 0" class="text-center text-[#a1a1aa]/70 text-[11px] mt-6">
            No diagrams yet
        </div>
    </div>
</div>
</template>
