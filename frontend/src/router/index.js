import { createRouter, createWebHistory } from 'vue-router'
import Home from '../views/Home.vue'
import { useAuthStore } from '../stores/auth'

const router = createRouter({
	history: createWebHistory(),
	routes: [
		{
			path: '/',
			component: Home
		},
		{
			path: '/portforward',
			component: () => import('../views/Portforward.vue')
		},
		{
			path: '/settings',
			component: () => import('../views/Settings.vue')
		},
		{
			path: '/login',
			component: () => import('../views/Login.vue')
		},
		{
			path: '/setup',
			component: () => import('../views/Setup.vue')
		},
	],
})

// Guards every navigation: first-run visitors are forced through /setup to
// create the single admin account; everyone else needs a valid session
// (silently restored from the httpOnly refresh cookie when possible) before
// reaching any page other than /login.
router.beforeEach(async (to) => {
	const auth = useAuthStore()
	await auth.init()

	if (auth.setupRequired) {
		return to.path === '/setup' ? true : '/setup'
	}
	if (to.path === '/setup') {
		return '/login'
	}
	if (!auth.isAuthenticated) {
		return to.path === '/login' ? true : '/login'
	}
	if (to.path === '/login') {
		return '/'
	}
	return true
})

export default router
