import { createApp } from 'vue';
import App from './App.vue';
import { naive } from 'naive-ui';

// const naive = create({
//   components: [NButton, NConfigProvider, NSlider, NSpace]
// });

const app = createApp(App);
app.use(naive);
app.mount('#app');
