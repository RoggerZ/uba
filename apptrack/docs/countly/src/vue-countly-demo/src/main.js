import { createApp } from 'vue';
import App from './App.vue';
import countlyPlugin from './plugins/countly';

const app = createApp(App);

// Use Countly plugin
// IMPORTANT: Replace these values with your actual Countly configuration
app.use(countlyPlugin, {
  appKey: '54fd10316e1fd7b4fc17a8bdef99759de8ba6262',  // Replace with your App Key
  url: 'http://local.count.ly:8380',  // Replace with your Countly server URL
  debug: true  // Enable debug mode for development (set to false in production)
});

app.mount('#app');
