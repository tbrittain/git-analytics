import { createRouter, createWebHashHistory } from 'vue-router'
import ContributorsPage from './pages/ContributorsPage.vue'
import CouplingPage from './pages/CouplingPage.vue'
import HomePage from './pages/HomePage.vue'
import HotspotsPage from './pages/HotspotsPage.vue'
import OwnershipPage from './pages/OwnershipPage.vue'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: '/', component: HomePage },
    { path: '/hotspots', component: HotspotsPage },
    { path: '/contributors', component: ContributorsPage },
    { path: '/ownership', component: OwnershipPage },
    { path: '/coupling', component: CouplingPage },
  ],
})

export default router
