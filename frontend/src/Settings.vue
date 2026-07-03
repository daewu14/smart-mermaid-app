<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted } from 'vue';
import { X, Github, Database, CheckCircle2, ChevronRight, Check } from 'lucide-vue-next';
import { showSettings, config, toastMessage } from './store';
import * as App from '../wailsjs/go/main/App.js';
import { BrowserOpenURL } from '../wailsjs/runtime/runtime.js';
import CustomSelect from './components/CustomSelect.vue';

const activeTab = ref('ai');
// ... other refs ...
const isConnectingGithub = ref(false);
const repos = ref<string[]>([]);
const isFetchingRepos = ref(false);
const githubStep = ref(0); // 0: initial, 1: connecting
const deviceFlowRes = ref<any>(null);
const pollInterval = ref<any>(null);

const localConfig = ref({
    baseUrl: config.baseUrl,
    apiKey: config.apiKey,
    githubToken: config.githubToken,
    githubRepo: config.githubRepo
});

watch(showSettings, (val) => {
    if (val) {
        localConfig.value = {
            baseUrl: config.baseUrl,
            apiKey: config.apiKey,
            githubToken: config.githubToken,
            githubRepo: config.githubRepo
        };
    }
});

const isTestingAI = ref(false);

const testConnection = async () => {
    if (!localConfig.value.baseUrl) {
        toastMessage.value = 'Please enter a Base URL';
        return;
    }
    isTestingAI.value = true;
    try {
        const res = await App.TestAIConnection(localConfig.value.baseUrl, localConfig.value.apiKey);
        if (res === 'success') {
            toastMessage.value = 'Connection successful!';
        } else {
            toastMessage.value = 'Test failed: ' + res;
        }
    } catch (e) {
        toastMessage.value = 'Connection failed: ' + e;
    } finally {
        isTestingAI.value = false;
    }
};

// Sync changes back to global state
const saveConfig = () => {
    config.baseUrl = localConfig.value.baseUrl;
    config.apiKey = localConfig.value.apiKey;
    config.githubToken = localConfig.value.githubToken;
    config.githubRepo = localConfig.value.githubRepo;
    localStorage.setItem('smart-mermaid-config', JSON.stringify(config));
    toastMessage.value = 'Settings saved';
};

const handleKeydown = (e: KeyboardEvent) => {
    if (e.key === 'Escape') {
        showSettings.value = false;
    }
};

onMounted(() => {
    document.addEventListener('keydown', handleKeydown);
    if (localConfig.value.githubToken) {
        fetchRepos(localConfig.value.githubToken);
    }
});

onUnmounted(() => {
    document.removeEventListener('keydown', handleKeydown);
    if (pollInterval.value) clearInterval(pollInterval.value);
});

const fetchRepos = async (token: string) => {
    isFetchingRepos.value = true;
    try {
        const list = await App.GetGitHubRepos(token);
        repos.value = list || [];
    } catch (e) {
        toastMessage.value = 'Failed to fetch repos: ' + e;
    } finally {
        isFetchingRepos.value = false;
    }
};

const startGithubAuth = async () => {
    isConnectingGithub.value = true;
    try {
        const clientId = "Ov23likD4E9U5f7L828c"; // Standard GitHub OAuth App Client ID
        const res = await App.StartGitHubDeviceFlow(clientId);
        if (res && res.device_code) {
            deviceFlowRes.value = res;
            githubStep.value = 1;
            BrowserOpenURL(res.verification_uri);

            // Start polling
            let expires = res.expires_in;
            pollInterval.value = setInterval(async () => {
                expires -= res.interval;
                if (expires <= 0) {
                    clearInterval(pollInterval.value);
                    toastMessage.value = 'GitHub Login Timeout';
                    githubStep.value = 0;
                    return;
                }
                try {
                    const tokenRes = await App.PollGitHubDeviceFlow(clientId, res.device_code);
                    if (tokenRes && tokenRes.access_token) {
                        clearInterval(pollInterval.value);
                        localConfig.value.githubToken = tokenRes.access_token;
                        saveConfig();
                        fetchRepos(tokenRes.access_token);
                        githubStep.value = 0;
                        toastMessage.value = 'GitHub connected successfully';
                    }
                } catch (e) {
                    // ignore pending errors
                }
            }, res.interval * 1000);
        } else {
            toastMessage.value = 'Failed to start OAuth flow.';
        }
    } catch (e) {
        toastMessage.value = String(e);
    } finally {
        isConnectingGithub.value = false;
    }
};

const disconnectGithub = () => {
    localConfig.value.githubToken = '';
    localConfig.value.githubRepo = '';
    repos.value = [];
    saveConfig();
    toastMessage.value = 'GitHub disconnected';
};
</script>

<template>
<transition name="modal-fade">
    <div v-if="showSettings" class="fixed top-0 left-0 w-full h-full bg-[#0a0c10]/70 backdrop-blur-sm flex items-center justify-center z-[9999] p-6">
        
        <!-- Modal Container -->
        <div class="modal-content bg-[#0e0e11]/90 backdrop-blur-3xl border border-white/5 rounded-2xl w-full max-w-3xl h-[600px] shadow-2xl flex overflow-hidden">
            
            <!-- Sidebar -->
            <div class="w-48 bg-[#131314]/80 border-r border-white/5 p-3 flex flex-col gap-1 shrink-0">
                <h2 class="text-[#8b949e] text-[11px] font-semibold uppercase tracking-wider mb-2 px-2">Settings</h2>
                
                <button @click="activeTab = 'ai'" class="w-full text-left px-3 py-2 rounded-lg text-[13px] font-medium flex items-center gap-2 transition-colors" :class="activeTab === 'ai' ? 'bg-[#1f6feb]/15 text-blue-400' : 'text-[#a1a1aa] hover:bg-white/10'">
                    <Database :size="16" /> AI Configuration
                </button>
                
                <button @click="activeTab = 'github'" class="w-full text-left px-3 py-2 rounded-lg text-[13px] font-medium flex items-center gap-2 transition-colors" :class="activeTab === 'github' ? 'bg-[#1f6feb]/15 text-blue-400' : 'text-[#a1a1aa] hover:bg-white/10'">
                    <Github :size="16" /> GitHub Sync
                </button>
            </div>

            <!-- Content -->
            <div class="flex-1 flex flex-col relative bg-transparent">
                <button @click="showSettings = false" class="absolute top-4 right-4 text-[#a1a1aa] hover:text-white p-1 rounded-lg hover:bg-white/10 transition-colors z-10">
                    <X :size="18" />
                </button>

                <div class="p-8 flex-1 overflow-y-auto">
                    <div v-if="activeTab === 'ai'" class="max-w-md">
                        <h1 class="text-xl font-semibold text-white mb-6">AI Configuration</h1>
                        
                        <div class="space-y-4">
                            <div class="space-y-1.5">
                                <label class="block text-[13px] font-medium text-slate-300">Base URL</label>
                                <input type="text" v-model="localConfig.baseUrl" placeholder="http://127.0.0.1:11434/v1" class="w-full bg-[#0d1117] border border-[#30363d] rounded-lg px-3 py-2 text-[13px] text-slate-200 focus:outline-none focus:border-[#58a6ff] focus:ring-1 focus:ring-[#58a6ff] transition-all">
                                <p class="text-[11px] text-slate-500">The OpenAI-compatible API endpoint (e.g. Ollama, LMStudio, Custom).</p>
                            </div>

                            <div class="space-y-1.5">
                                <label class="block text-[13px] font-medium text-slate-300">API Key (Optional)</label>
                                <input type="password" v-model="localConfig.apiKey" placeholder="sk-..." class="w-full bg-[#0d1117] border border-[#30363d] rounded-lg px-3 py-2 text-[13px] text-slate-200 focus:outline-none focus:border-[#58a6ff] focus:ring-1 focus:ring-[#58a6ff] transition-all">
                            </div>

                            <div class="pt-4 border-t border-[#30363d] flex items-center gap-2">
                                <button @click="testConnection" :disabled="isTestingAI" class="bg-[#21262d] hover:bg-[#30363d] text-slate-300 px-4 py-2 rounded-lg text-[13px] font-medium transition-colors border border-[#30363d] disabled:opacity-50 flex items-center gap-2">
                                    <template v-if="isTestingAI">
                                        <div class="w-3.5 h-3.5 border-2 border-slate-400/30 border-t-slate-400 rounded-full animate-spin"></div> Testing...
                                    </template>
                                    <template v-else>
                                        Test Connection
                                    </template>
                                </button>
                                <button @click="saveConfig" class="bg-[#238636] hover:bg-[#2ea043] text-white px-4 py-2 rounded-lg text-[13px] font-medium transition-colors">
                                    Save AI Configuration
                                </button>
                            </div>
                        </div>
                    </div>

                    <div v-if="activeTab === 'github'" class="max-w-md">
                        <h1 class="text-xl font-semibold text-white mb-6">GitHub Sync</h1>

                        <div v-if="localConfig.githubToken" class="space-y-6">
                            <div class="bg-[#0d1117] border border-[#238636]/30 rounded-xl p-4 flex items-center justify-between">
                                <div class="flex items-center gap-3">
                                    <div class="w-10 h-10 rounded-full bg-[#238636]/10 flex items-center justify-center text-[#238636]">
                                        <CheckCircle2 :size="20" />
                                    </div>
                                    <div>
                                        <div class="text-[14px] font-medium text-slate-200">Connected to GitHub</div>
                                        <div class="text-[12px] text-slate-500">You can now sync diagrams to your repos.</div>
                                    </div>
                                </div>
                                <button @click="disconnectGithub" class="text-[12px] text-red-400 hover:text-red-300 font-medium px-3 py-1.5 rounded hover:bg-red-400/10 transition-colors">
                                    Disconnect
                                </button>
                            </div>

                            <div class="space-y-1.5 relative z-20">
                                <label class="block text-[13px] font-medium text-slate-300">Target Repository</label>
                                <CustomSelect 
                                    v-model="localConfig.githubRepo" 
                                    :options="repos" 
                                    placeholder="Select a repository..." 
                                    direction="down" 
                                    @change="saveConfig"
                                />
                                <p v-if="isFetchingRepos" class="text-[11px] text-slate-500 animate-pulse mt-1">Loading repositories...</p>
                            </div>
                        </div>

                        <div v-else class="space-y-6">
                            <div class="bg-[#0d1117] border border-[#30363d] rounded-xl p-6 text-center space-y-4">
                                <div class="w-12 h-12 rounded-full bg-[#21262d] flex items-center justify-center mx-auto text-slate-300 mb-2">
                                    <Github :size="24" />
                                </div>
                                <h3 class="text-[15px] font-medium text-slate-200">Connect your GitHub Account</h3>
                                <p class="text-[13px] text-slate-500 max-w-sm mx-auto leading-relaxed">
                                    Connect GitHub to sync your Mermaid diagrams directly to your repositories.
                                </p>
                                
                                <div v-if="githubStep === 0" class="pt-2">
                                    <button @click="startGithubAuth" :disabled="isConnectingGithub" class="bg-white text-black hover:bg-slate-200 px-5 py-2.5 rounded-lg text-[13px] font-semibold transition-colors flex items-center gap-2 mx-auto disabled:opacity-50">
                                        <template v-if="isConnectingGithub">
                                            <div class="w-3.5 h-3.5 border-2 border-black/30 border-t-black rounded-full animate-spin"></div> Connecting...
                                        </template>
                                        <template v-else>
                                            <Github :size="16" /> Authenticate with GitHub
                                        </template>
                                    </button>
                                </div>

                                <div v-if="githubStep === 1" class="pt-4 border-t border-[#30363d] text-left space-y-4">
                                    <p class="text-[13px] text-slate-300">
                                        Please enter this code in the browser window that just opened:
                                    </p>
                                    <div class="text-3xl font-mono tracking-widest text-white text-center font-bold py-4 bg-[#21262d] rounded-lg">
                                        {{ deviceFlowRes?.user_code }}
                                    </div>
                                    <p class="text-[11px] text-slate-500 text-center flex items-center justify-center gap-2">
                                        <div class="w-3 h-3 border-[1.5px] border-slate-500/30 border-t-slate-500 rounded-full animate-spin"></div> Waiting for approval...
                                    </p>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</transition>
</template>
