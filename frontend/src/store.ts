import { reactive, ref } from 'vue';

export const config = reactive({
    baseUrl: '',
    apiKey: '',
    githubToken: '',
    githubRepo: ''
});

export const diagrams = ref<any[]>([]);
export const currentDiagram = ref<any>(null);

export const showSettings = ref(false);
export const isSidebarOpen = ref(true);
export const isFullscreen = ref(false);
export const toastMessage = ref('');

export const zoom = ref(1);
export const zoomIn = () => {
    zoom.value = Math.min(zoom.value * 1.2, 300);
};
export const zoomOut = () => {
    zoom.value = Math.max(zoom.value / 1.2, 0.1);
};
