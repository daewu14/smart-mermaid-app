<script setup lang="ts">
import { ref, watch, onMounted, nextTick } from 'vue';
import { Send, Save, CloudUpload, Code, ChevronDown, ZoomIn, ZoomOut } from 'lucide-vue-next';
import { config, currentDiagram, diagrams, toastMessage, showSettings, zoom, zoomIn, zoomOut, isGenerating } from './store';
import * as App from '../wailsjs/go/main/App.js';

import CustomSelect from './components/CustomSelect.vue';

const prompt = ref('');
const models = ref<string[]>([]);
const selectedModel = ref('');
const title = ref('');
const showCode = ref(false);
const textareaEl = ref<HTMLTextAreaElement | null>(null);

watch(currentDiagram, (val) => {
    if (val) {
        title.value = val.name;
        // Restore selected model if it exists in the diagram and in the available models
        if (val.model && models.value.includes(val.model)) {
            selectedModel.value = val.model;
        }
    }
}, { immediate: true, deep: true });

const fetchModels = async () => {
    if (!config.baseUrl) return;
    try {
        const res = await fetch(`http://127.0.0.1:45543/proxy`, {
            headers: {
                'X-Target-Url': `${config.baseUrl.endsWith('/') ? config.baseUrl.slice(0, -1) : config.baseUrl}/models`,
                ...(config.apiKey ? { 'Authorization': `Bearer ${config.apiKey}` } : {})
            }
        });
        const data = await res.json();
        
        let parsedModels: string[] = [];
        if (data.data) {
            parsedModels = data.data.map((m: any) => m.id || m.name);
        } else if (data.models) {
            parsedModels = data.models.map((m: any) => m.name || m.id);
        } else if (Array.isArray(data)) {
            parsedModels = data.map((m: any) => m.id || m.name);
        }
        
        if (parsedModels.length > 0) {
            models.value = parsedModels;
            if (currentDiagram.value?.model && models.value.includes(currentDiagram.value.model)) {
                selectedModel.value = currentDiagram.value.model;
            } else if (!selectedModel.value || !models.value.includes(selectedModel.value)) {
                selectedModel.value = models.value[0];
            }
        }
    } catch (e) {
        console.error('Failed to fetch models', e);
    }
};

onMounted(() => {
    fetchModels();
});

watch(() => [config.baseUrl, config.apiKey], () => {
    fetchModels();
});

const adjustHeight = () => {
    nextTick(() => {
        if (textareaEl.value) {
            textareaEl.value.style.height = 'auto';
            textareaEl.value.style.height = Math.min(textareaEl.value.scrollHeight, 150) + 'px';
        }
    });
};

const handleKeydown = (e: KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        generate();
    }
};

const generate = async () => {
    if (!prompt.value.trim() || !config.baseUrl || !selectedModel.value || isGenerating.value) return;

    isGenerating.value = true;
    
    if (!currentDiagram.value.history) {
        currentDiagram.value.history = [];
    }

    const currentCode = currentDiagram.value.content;
    let contextPrompt = prompt.value;
    
    if (currentCode && currentCode.trim() !== 'graph TD\n    A-->B;') {
        contextPrompt = prompt.value + `\n\nCurrent Mermaid Code:\n\`\`\`mermaid\n${currentCode}\n\`\`\``;
    }

    const apiMessages = [
        { role: 'system', content: 'You are an expert Mermaid.js generator. Only output valid Mermaid code wrapped in ```mermaid ... ```. Do not include any other explanations, greetings, or text.' },
        ...currentDiagram.value.history,
        { role: 'user', content: contextPrompt }
    ];

    currentDiagram.value.content = '';
    const userOriginalPrompt = prompt.value;
    prompt.value = '';
    adjustHeight();

    try {
        const res = await fetch(`http://127.0.0.1:45543/proxy`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-Target-Url': `${config.baseUrl.endsWith('/') ? config.baseUrl.slice(0, -1) : config.baseUrl}/chat/completions`,
                ...(config.apiKey ? { 'Authorization': `Bearer ${config.apiKey}` } : {})
            },
            body: JSON.stringify({
                model: selectedModel.value,
                messages: apiMessages,
                stream: true
            })
        });

        if (!res.ok) {
            let errorMsg = `API Error (${res.status})`;
            try {
                const errData = await res.json();
                errorMsg = errData.error?.message || errData.message || errorMsg;
            } catch (e) {}
            throw new Error(errorMsg);
        }

        const reader = res.body?.getReader();
        const decoder = new TextDecoder();
        let fullText = '';
        let finalMermaidCode = '';

        if (reader) {
            while (true) {
                const { done, value } = await reader.read();
                if (done) break;
                
                const chunk = decoder.decode(value, { stream: true });
                const lines = chunk.split('\n');
                
                for (const line of lines) {
                    if (line.startsWith('data: ') && line !== 'data: [DONE]') {
                        try {
                            const data = JSON.parse(line.slice(6));
                            if (data.choices[0].delta.content) {
                                fullText += data.choices[0].delta.content;
                                
                                let content = fullText;
                                const match = content.match(/```mermaid([\s\S]*?)(?:```|$)/);
                                if (match) {
                                    content = match[1].trim();
                                } else {
                                    content = fullText.replace(/```.*/g, '').trim();
                                }
                                
                                currentDiagram.value.content = content;
                                finalMermaidCode = content;
                            }
                        } catch (e) {}
                    }
                }
            }
        }
        
        // Push to history after successful generation
        currentDiagram.value.history.push({ role: 'user', content: userOriginalPrompt });
        currentDiagram.value.history.push({ role: 'assistant', content: "```mermaid\n" + finalMermaidCode + "\n```" });
        
    } catch (e: any) {
        console.error(e);
        toastMessage.value = e.message || 'Generation failed.';
    } finally {
        isGenerating.value = false;
    }
};

const saveDiagram = async () => {
    if (!currentDiagram.value) return;
    currentDiagram.value.name = title.value;
    currentDiagram.value.model = selectedModel.value;
    
    await App.SaveDiagram(currentDiagram.value);
    const list = await App.GetDiagrams();
    diagrams.value = list || [];
    if (!currentDiagram.value.id) {
         const updated = (list || []).find((d: any) => d.name === title.value && d.content === currentDiagram.value.content);
         if (updated) currentDiagram.value = updated;
    }
    toastMessage.value = 'Diagram saved';
};

const isSyncing = ref(false);

const pushToGithub = async () => {
    if (!config.githubToken || !config.githubRepo) {
        toastMessage.value = 'Please configure GitHub OAuth & Repo in Settings.';
        showSettings.value = true;
        return;
    }
    
    await saveDiagram();
    
    isSyncing.value = true;
    try {
        await App.SyncToGitHub(config.githubToken, config.githubRepo, currentDiagram.value.id);
        const list = await App.GetDiagrams();
        diagrams.value = list || [];
        const updated = (list || []).find((d: any) => d.id === currentDiagram.value.id);
        if (updated) currentDiagram.value = updated;
        toastMessage.value = 'Synced to GitHub';
    } catch (e) {
        toastMessage.value = 'Sync failed: ' + e;
    } finally {
        isSyncing.value = false;
    }
};
</script>

<template>
<div class="flex flex-col gap-2 relative">
    
    <!-- Code Editor Drawer -->
    <transition name="slide-down">
        <div v-if="showCode" class="absolute bottom-full mb-2 w-full bg-white/95 dark:bg-[#161b22]/95 backdrop-blur-md border border-slate-200 dark:border-[#30363d] rounded-xl overflow-hidden shadow-2xl z-20">
            <div class="bg-slate-50 dark:bg-[#0d1117] px-3 py-1.5 text-[11px] font-mono text-slate-500 dark:text-slate-400 flex justify-between items-center border-b border-slate-200 dark:border-[#30363d]">
                Raw Mermaid Code
                <button @click="showCode = false" class="hover:text-slate-800 dark:hover:text-slate-200"><ChevronDown :size="14"/></button>
            </div>
            <textarea v-model="currentDiagram.content" class="w-full h-40 bg-transparent p-3 text-slate-800 dark:text-[#c9d1d9] font-mono text-[12px] focus:outline-none resize-y" spellcheck="false"></textarea>
        </div>
    </transition>

    <!-- Prompt Box -->
    <div class="w-full bg-white/70 dark:bg-[#131314]/70 backdrop-blur-3xl border border-slate-200 dark:border-white/5 rounded-2xl shadow-2xl flex flex-col transition-colors focus-within:border-slate-300 dark:focus-within:border-white/20 focus-within:bg-white/90 dark:focus-within:bg-[#18181a]/90 relative z-30">
        <!-- Top Toolbar -->
        <div class="flex items-center justify-between px-3 py-1.5 border-b border-slate-200 dark:border-white/5 bg-transparent rounded-t-2xl">
            <input type="text" v-model="title" placeholder="Untitled Diagram" class="bg-transparent border-none text-slate-500 dark:text-[#a1a1aa] text-[12px] font-medium focus:outline-none w-1/3 focus:text-slate-900 dark:focus:text-white transition-colors">
            
            <div class="flex items-center gap-1.5">
                <button @click="zoomOut" class="p-1 rounded text-slate-500 dark:text-[#a1a1aa] hover:text-slate-800 dark:hover:text-white hover:bg-black/5 dark:hover:bg-white/10 transition-colors" title="Zoom Out">
                    <ZoomOut :size="13" />
                </button>
                <span class="text-[11px] font-medium text-slate-700 dark:text-white w-9 text-center">{{ Math.round(zoom * 100) }}%</span>
                <button @click="zoomIn" class="p-1 rounded text-slate-500 dark:text-[#a1a1aa] hover:text-slate-800 dark:hover:text-white hover:bg-black/5 dark:hover:bg-white/10 transition-colors" title="Zoom In">
                    <ZoomIn :size="13" />
                </button>
                
                <div class="w-px h-3 bg-black/10 dark:bg-white/10 mx-1"></div>

                <button @click="showCode = !showCode" class="p-1 rounded text-slate-500 dark:text-[#a1a1aa] hover:text-slate-800 dark:hover:text-white hover:bg-black/5 dark:hover:bg-white/10 transition-colors" title="Code">
                    <Code :size="13" />
                </button>
                <div class="w-px h-3 bg-black/10 dark:bg-white/10 mx-0.5"></div>
                <button @click="saveDiagram" class="p-1 rounded text-slate-500 dark:text-[#a1a1aa] hover:text-slate-800 dark:hover:text-white hover:bg-black/5 dark:hover:bg-white/10 transition-colors" title="Save">
                    <Save :size="13" />
                </button>
                <button @click="pushToGithub" :disabled="isSyncing" class="px-1.5 py-0.5 rounded flex items-center gap-1 text-[#238636] hover:bg-[#238636]/10 transition-colors text-[11px] font-medium disabled:opacity-50" title="Sync">
                    <template v-if="isSyncing">
                        <div class="w-3 h-3 border-2 border-[#238636]/30 border-t-[#238636] rounded-full animate-spin"></div> Syncing...
                    </template>
                    <template v-else>
                        <CloudUpload :size="12" /> Sync
                    </template>
                </button>
            </div>
        </div>

        <!-- Input Area -->
        <div class="flex flex-col p-2">
            <div class="flex-1 flex flex-col">
                <textarea 
                    ref="textareaEl"
                    v-model="prompt" 
                    @keydown="handleKeydown"
                    @input="adjustHeight"
                    placeholder="Describe the diagram..." 
                    class="bg-transparent border-none px-2 py-1.5 text-[13px] text-slate-800 dark:text-[#c9d1d9] placeholder-slate-400 dark:placeholder-[#8b949e] focus:outline-none resize-none min-h-[36px] max-h-[150px] custom-scrollbar w-full"
                    rows="1"
                ></textarea>
            </div>
            
            <div class="px-2 pb-1 flex justify-between items-center mt-2">
                <CustomSelect v-model="selectedModel" :options="models" placeholder="Model..." direction="up" class="w-48" />
                
                <button 
                    @click="generate" 
                    :disabled="isGenerating || !selectedModel || !prompt.trim()" 
                    class="p-2.5 rounded-xl bg-black/5 dark:bg-white/10 hover:bg-black/10 dark:hover:bg-white/20 disabled:bg-black/5 dark:disabled:bg-white/5 disabled:text-slate-400 dark:disabled:text-[#484f58] disabled:cursor-not-allowed text-slate-700 dark:text-white transition-all flex items-center justify-center shrink-0 w-9 h-9"
                >
                    <template v-if="isGenerating">
                        <div class="w-4 h-4 border-2 border-slate-300 dark:border-white/30 border-t-slate-600 dark:border-t-white rounded-full animate-spin"></div>
                    </template>
                    <template v-else>
                        <Send :size="15" class="pr-0.5 pb-0.5" />
                    </template>
                </button>
            </div>
        </div>
    </div>
</div>
</template>
