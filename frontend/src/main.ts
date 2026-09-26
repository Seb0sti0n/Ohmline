import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import { setUnauthorizedHandler } from './services/api'
import { useAuthStore } from './stores/auth'
import { useAnalysisStore } from './stores/analysis'
import './assets/main.css'

const app = createApp(App)
app.use(createPinia()).use(router)

// An expired session (401) clears the state and returns to the login screen.
setUnauthorizedHandler(() => {
  useAnalysisStore().stop()
  useAuthStore().logout()
  void router.push({ name: 'login' })
})

app.mount('#app')
