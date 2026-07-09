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
        updatedAt: '',
        history: []
    };
};

const selectDiagram = (d: any) => {
    currentDiagram.value = d;
};

const showDeleteConfirm = ref(false);
const diagramToDelete = ref<any>(null);
const deleteFromGithubAlso = ref(false);
const isDeleting = ref(false);

const confirmDelete = (e: Event, diagram: any) => {
    e.stopPropagation();
    diagramToDelete.value = diagram;
    deleteFromGithubAlso.value = false;
    showDeleteConfirm.value = true;
};

const executeDelete = async () => {
    if (!diagramToDelete.value) return;
    isDeleting.value = true;
    try {
        if (deleteFromGithubAlso.value && diagramToDelete.value.synced && diagramToDelete.value.sha) {
            if (!config.githubToken || !config.githubRepo) {
                throw new Error("GitHub not connected or repo not selected");
            }
            await App.DeleteFromGitHub(config.githubToken, config.githubRepo, diagramToDelete.value.name, diagramToDelete.value.sha);
        }
        await App.DeleteDiagram(diagramToDelete.value.id);
        
        if (currentDiagram.value && currentDiagram.value.id === diagramToDelete.value.id) {
            currentDiagram.value = null;
        }
        toastMessage.value = 'Diagram deleted successfully';
        loadDiagrams();
    } catch (e) {
        toastMessage.value = 'Failed to delete: ' + String(e);
    } finally {
        isDeleting.value = false;
        showDeleteConfirm.value = false;
        diagramToDelete.value = null;
    }
};

</script>

<template>
<div class="h-full bg-slate-100 dark:bg-[#131314] flex flex-col text-slate-500 dark:text-[#a1a1aa] border-r border-slate-200 dark:border-white/5 transition-colors duration-300">
    <!-- Header (Draggable for macOS) -->
    <div style="--wails-draggable: drag" class="h-[40px] pr-4 pl-[110px] border-b border-slate-200 dark:border-white/5 flex items-center justify-between shrink-0" @dblclick="WindowToggleMaximise">
        <h1 class="font-semibold text-[13px] text-slate-500 dark:text-[#a1a1aa] flex items-center gap-2">
            Diagrams
        </h1>
        <div style="--wails-draggable: no-drag" class="flex items-center gap-1">
            <button @click="showSettings = true" class="text-slate-500 dark:text-[#a1a1aa] hover:text-slate-800 dark:hover:text-white p-1 rounded hover:bg-black/5 dark:hover:bg-white/10 transition-colors" title="Settings">
                <Settings :size="14" />
            </button>
        </div>
    </div>

    <!-- Actions -->
    <div class="p-3 border-b border-slate-200 dark:border-white/5 shrink-0">
        <button @click="createNew" :disabled="isTestingConnection" class="w-full bg-white dark:bg-white/5 hover:bg-slate-50 dark:hover:bg-white/10 text-slate-700 dark:text-white rounded-md py-1.5 px-2 text-[12px] font-medium flex items-center justify-center gap-1.5 transition-colors border border-slate-200 dark:border-white/5 shadow-sm dark:shadow-none disabled:opacity-50 disabled:cursor-not-allowed">
            <template v-if="isTestingConnection">
                <div class="w-3.5 h-3.5 border-2 border-slate-400/30 dark:border-white/30 border-t-slate-500 dark:border-t-white rounded-full animate-spin"></div> Checking...
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
            :class="currentDiagram?.id === diagram.id ? 'bg-blue-50 dark:bg-[#1f6feb]/15 text-blue-600 dark:text-blue-400' : 'text-slate-600 dark:text-[#a1a1aa] hover:bg-black/5 dark:hover:bg-white/5 hover:text-slate-900 dark:hover:text-white'"
            @click="selectDiagram(diagram)"
        >
            <div class="truncate flex-1 pr-2">
                {{ diagram.name || 'Untitled' }}
            </div>
            <div class="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                <Cloud v-if="diagram.synced" :size="12" class="text-blue-500 dark:text-blue-400" title="Synced to GitHub" />
                <CloudOff v-else :size="12" class="text-slate-400 dark:text-slate-500" title="Not synced" />
                <button @click="(e) => confirmDelete(e, diagram)" class="p-1 hover:bg-black/5 dark:hover:bg-white/10 rounded text-slate-400 hover:text-red-500 dark:hover:text-red-400 transition-colors ml-1" title="Delete">
                    <Trash2 :size="12" />
                </button>
            </div>
        </div>
        
        <div v-if="diagrams.length === 0" class="text-center py-8 px-4 text-slate-500 dark:text-[#a1a1aa] text-[12px] border border-dashed border-slate-300 dark:border-white/10 rounded-lg">
            No diagrams yet.<br>Create one to get started!
        </div>
    </div>

    <!-- Delete Confirmation Modal -->
    <transition name="modal-fade">
        <div v-if="showDeleteConfirm" class="fixed top-0 left-0 w-full h-full bg-slate-900/20 dark:bg-[#0a0c10]/70 backdrop-blur-sm flex items-center justify-center z-[9999] p-6">
            <div class="modal-content bg-white dark:bg-[#0e0e11]/90 backdrop-blur-3xl border border-slate-200 dark:border-white/5 rounded-2xl w-full max-w-sm shadow-2xl overflow-hidden" @click.stop>
                <div class="p-6">
                    <h3 class="text-lg font-semibold text-slate-800 dark:text-white mb-2">Delete Diagram</h3>
                    <p class="text-[13px] text-slate-500 dark:text-slate-400 mb-5">
                        Are you sure you want to delete <b>{{ diagramToDelete?.name || 'Untitled' }}</b>? This action cannot be undone.
                    </p>
                    
                    <label v-if="diagramToDelete?.synced && diagramToDelete?.sha" class="flex items-center gap-2 text-[12px] text-slate-600 dark:text-slate-300 mb-6 cursor-pointer bg-slate-50 dark:bg-white/5 p-3 rounded-lg border border-slate-200 dark:border-white/5 hover:bg-slate-100 dark:hover:bg-white/10 transition-colors">
                        <input type="checkbox" v-model="deleteFromGithubAlso" class="rounded bg-white dark:bg-black/50 border-slate-300 dark:border-white/20 text-red-500 focus:ring-red-500/30 w-4 h-4">
                        Also delete from GitHub repository
                    </label>
                    
                    <div class="flex justify-end gap-2">
                        <button @click="showDeleteConfirm = false" :disabled="isDeleting" class="px-4 py-2 rounded-lg text-[13px] font-medium text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-white/10 transition-colors disabled:opacity-50">
                            Cancel
                        </button>
                        <button @click="executeDelete" :disabled="isDeleting" class="px-4 py-2 rounded-lg text-[13px] font-medium bg-red-500/20 text-red-400 hover:bg-red-500/30 transition-colors flex items-center gap-2 disabled:opacity-50">
                            <template v-if="isDeleting">
                                <div class="w-3.5 h-3.5 border-2 border-red-400/30 border-t-red-400 rounded-full animate-spin"></div> Deleting...
                            </template>
                            <template v-else>
                                Delete
                            </template>
                        </button>
                    </div>
                </div>
            </div>
        </div>
    </transition>
</div>
</template>
