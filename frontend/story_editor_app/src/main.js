import { createApp } from 'vue';
import StoryEditor from './components/StoryEditor.vue';

// Create and mount the Vue application.
const app = createApp(StoryEditor);
//app.component('story-editor', StoryEditor); // Register the component
app.mount('#app');

