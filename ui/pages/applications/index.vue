<template>
	<div class="px-4 py-3 sm:px-6 sm:py-4 space-y-6 sm:space-y-8">
		<div class="space-y-2">
			<h1 class="text-xl font-semibold text-base-content">Applications</h1>
			<div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
				<p class="text-base-content/60 text-sm">Manage your web applications and create new installations</p>
				<div class="flex flex-wrap gap-2 w-full md:w-auto items-center">
					<!-- Search -->
					<div class="relative w-full sm:w-auto flex-1 sm:flex-none">
						<input 
							v-model="searchQuery"
							type="text" 
							placeholder="Search applications..." 
							class="pl-8 pr-4 py-2 w-full sm:w-48 bg-base-200 border border-base-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
						>
						<Icon name="mdi:magnify" size="16" class="absolute left-2.5 top-2.5 text-base-content/40" />
					</div>
					
					<!-- Filter -->
					<select 
							v-model="statusFilter"
							class="px-3 py-2 bg-base-200 border border-base-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-primary-500 flex-1 sm:flex-none"
					>
						<option value="">All Status</option>
						<option value="installed">Installed</option>
						<option value="running">Running</option>
						<option value="stopped">Stopped</option>
						<option value="updated">Updated</option>
					</select>
					
					<!-- Create Button -->
					<button @click="navigateTo('/registry')" 
							class="px-4 py-2 btn btn-primary rounded-md transition-colors flex items-center gap-2 w-full sm:w-auto justify-center text-sm">
						<Icon name="mdi:plus" size="20" />
						Install Application
					</button>
				</div>
			</div>
		</div>

		<!-- Applications Grid -->
		<div v-if="filteredApplications?.length" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
			<div 
					v-for="app in filteredApplications" 
					:key="app.id"
					class="card bg-base-200 hover:bg-base-300 transition-all duration-200 cursor-pointer"
					@click="openApplication(app)"
			>
				<!-- Header -->
				<div class="p-4 border-b border-base-300">
					<div class="flex items-center justify-between mb-3">
						<div class="flex items-center gap-3">
							<Icon :name="getAppIcon(app)" size="32" class="text-base-content/70" />
							<div>
								<h3 class="font-bold text-base-content">{{ app.display_name }}</h3>
								<p class="text-base-content/60 text-sm">{{ app.description }}</p>
							</div>
						</div>
						<div class="flex items-center gap-2">
							<span 
									:class="getStatusClass(app.status)"
									class="px-2 py-1 text-xs rounded-full capitalize"
							>
								{{ app.status || "Installed" }}
							</span>
						</div>
					</div>
					
					<!-- Quick Actions -->
					<div class="flex gap-2">
						<a 
								v-if="app.domain"
								:href="`//${app.domain}`" 
								target="_blank" 
								@click.stop
								class="px-3 py-1 text-xs btn btn-success rounded flex items-center gap-1"
						>
							<Icon name="mdi:open-in-new" size="14" />
							Visit
						</a>
						
						<button 
								@click.stop="toggleAppStatus(app)"
								:disabled="statusLoading === app.id"
								:class="app.status === 'running' ? 'btn-warning' : 'btn-primary'"
								class="px-3 py-1 text-xs rounded flex items-center gap-1 disabled:opacity-50"
						>
							<Icon 
									:name="statusLoading === app.id ? 'mdi:loading' : (app.status === 'running' ? 'mdi:stop' : 'mdi:play')" 
									:size="14" 
									:class="statusLoading === app.id ? 'animate-spin' : ''" 
							/>
							{{ app.status === 'running' ? 'Stop' : 'Start' }}
						</button>
					</div>
				</div>

				<!-- Details -->
				<div class="p-4 space-y-3">
					<div class="grid grid-cols-2 gap-4 text-sm">
						<div class="space-y-1">
							<span class="text-base-content/60">Provider:</span>
							<p class="text-base-content/80 capitalize">{{ app.provider }}</p>
						</div>
						<div class="space-y-1">
							<span class="text-base-content/60">Domain:</span>
							<p class="text-base-content/80 font-mono">{{ app.domain || 'Not set' }}</p>
						</div>
						<div class="space-y-1">
							<span class="text-base-content/60">Runtime:</span>
							<p class="text-base-content/80">{{ app.runtime }}</p>
						</div>
						<div class="space-y-1">
							<span class="text-base-content/60">Database:</span>
							<p class="text-base-content/80">{{ app.database }}</p>
						</div>
					</div>
					
					<!-- Path -->
					<div class="space-y-1">
						<span class="text-base-content/60">Path:</span>
						<p class="text-base-content/80 font-mono text-xs bg-base-300 p-2 rounded">{{ app.path }}</p>
					</div>
					
					<!-- Deployment Method -->
					<div class="flex items-center gap-2 text-sm">
						<span class="text-base-content/60">Deployed via:</span>
						<span class="px-2 py-1 bg-base-300 rounded text-xs capitalize">{{ app.deploy_method }}</span>
					</div>
				</div>
			</div>
		</div>

		<!-- Empty State -->
		<div v-else-if="!loading" class="text-center py-12">
			<Icon name="mdi:application-outline" size="64" class="text-slate-600 mx-auto mb-4" />
			<h3 class="text-slate-300 text-lg font-medium mb-2">
				{{ searchQuery || statusFilter ? 'No Applications Found' : 'No Applications Installed' }}
			</h3>
			<p class="text-base-content/60 mb-4">
				{{ searchQuery || statusFilter ? 'Try adjusting your search or filters' : 'Get started by installing your first application' }}
			</p>
			<button 
					v-if="!searchQuery && !statusFilter"
					@click="navigateTo('/registry')"
					class="px-6 py-2 btn btn-primary rounded-md flex items-center gap-2 mx-auto"
			>
				<Icon name="mdi:plus" size="20" />
				Install Your First App
			</button>
		</div>

		<!-- Loading State -->
		<div v-if="loading" class="text-center py-12">
			<Icon name="mdi:loading" size="32" class="animate-spin text-base-content/60 mx-auto mb-4" />
			<p class="text-base-content/60">Loading applications...</p>
		</div>
	</div>
</template>

<script setup>
const applications = ref([])
const loading = ref(true)
const searchQuery = ref('')
const statusFilter = ref('')
const statusLoading = ref(null)

const snackbar = useSnackbar()
const route = useRoute()
const router = useRouter()

// Computed property for filtered applications
const filteredApplications = computed(() => {
	let filtered = applications.value

	// Apply search filter
	if (searchQuery.value) {
		const query = searchQuery.value.toLowerCase()
		filtered = filtered.filter(app => 
			app.display_name.toLowerCase().includes(query) ||
			app.provider.toLowerCase().includes(query) ||
			app.domain?.toLowerCase().includes(query) ||
			app.path.toLowerCase().includes(query)
		)
	}

	// Apply status filter
	if (statusFilter.value) {
		filtered = filtered.filter(app => app.status === statusFilter.value)
	}

	return filtered
})

// Load applications
const loadApplications = async () => {
	loading.value = true
	try {
		const response = await useNuxtApp().$axiosApi.get('/applications')
		applications.value = response.data.data || []
	} catch (error) {
		console.error('Failed to load applications:', error)
		snackbar.add({
			type: "error",
			text: "Failed to load applications"
		})
	} finally {
		loading.value = false
	}
}

// Application actions
const openApplication = (app) => {
	router.push(`/applications/${app.id}`)
}

const toggleAppStatus = async (app) => {
	if (statusLoading.value) return

	statusLoading.value = app.id
	const action = app.status === 'running' ? 'stop' : 'start'

	try {
		const response = await useNuxtApp().$axiosApi.post(`/applications/${app.id}/${action}`)
		
		if (response.data.success) {
			// Update local status
			const index = applications.value.findIndex(a => a.id === app.id)
			if (index !== -1) {
				applications.value[index].status = action === 'start' ? 'running' : 'stopped'
			}
			
			snackbar.add({
				type: "success",
				text: `Application ${action}ed successfully`
			})
		} else {
			throw new Error(response.data.message || `Failed to ${action} application`)
		}
	} catch (error) {
		console.error(`Failed to ${action} application:`, error)
		snackbar.add({
			type: "error",
			text: `Failed to ${action} application`
		})
	} finally {
		statusLoading.value = null
	}
}

// Utility functions
const getAppIcon = (app) => {
	// Use icon from manifest, fallback to default
	return app.icon || 'mdi:application-outline'
}

const getStatusClass = (status) => {
	const statusClasses = {
		'running': 'bg-green-500 text-white',
		'stopped': 'bg-red-500 text-white',
		'installed': 'bg-blue-500 text-white',
		'updated': 'bg-yellow-500 text-white'
	}
	return statusClasses[status] || 'bg-gray-500 text-white'
}

// Load on mount
onMounted(() => {
	loadApplications()
})

// Watch for query parameters
watch(() => route.query, (newQuery) => {
	if (newQuery.search) {
		searchQuery.value = newQuery.search
	}
	if (newQuery.status) {
		statusFilter.value = newQuery.status
	}
}, { immediate: true })

// Update URL when filters change
watch([searchQuery, statusFilter], () => {
	router.push({
		query: {
			...route.query,
			search: searchQuery.value || undefined,
			status: statusFilter.value || undefined
		}
	})
})
</script>

<style scoped>
.card {
	transition: all 0.2s ease;
}

.card:hover {
	box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
	transform: scale(1.02);
}

.animate-spin {
	animation: spin 1s linear infinite;
}

@keyframes spin {
	from { transform: rotate(0deg); }
	to { transform: rotate(360deg); }
}
</style>