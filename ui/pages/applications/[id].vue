<template>
	<div class="px-4 py-3 sm:px-6 sm:py-4 space-y-6 sm:space-y-8">
		<!-- Loading State -->
		<div v-if="loading" class="text-center py-12">
			<Icon name="mdi:loading" size="32" class="animate-spin text-base-content/60 mx-auto mb-4" />
			<p class="text-base-content/60">Loading application details...</p>
		</div>

		<!-- Application Details -->
		<div v-else-if="application" class="space-y-6">
			<!-- Header -->
			<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
				<div class="flex items-center gap-4 min-w-0">
					<Icon :name="getAppIcon(application.provider)" size="40" class="text-base-content/70 flex-shrink-0" />
					<div class="min-w-0">
						<h1 class="text-2xl font-bold text-base-content truncate">{{ application.display_name }}</h1>
						<p class="text-base-content/60 truncate">{{ application.description }}</p>
					</div>
				</div>
				<div class="flex items-center gap-2 w-full sm:w-auto justify-between sm:justify-end border-t border-base-300/30 pt-3 sm:pt-0 sm:border-t-0">
					<span 
							:class="getStatusClass(application.status)"
							class="px-3 py-1 text-sm rounded-full capitalize font-medium"
					>
						{{ application.status }}
					</span>
					<a 
							v-if="application.domain"
							:href="`//${application.domain}`" 
							target="_blank" 
							class="px-3 py-1 text-sm btn btn-success rounded flex items-center gap-1 text-xs"
					>
						<Icon name="mdi:open-in-new" size="16" />
						Visit Site
					</a>
				</div>
			</div>

			<!-- Quick Actions -->
			<div class="flex flex-wrap gap-2">
				<button 
						@click="toggleAppStatus()"
						:disabled="statusLoading"
						:class="application.status === 'running' ? 'btn-warning' : 'btn-primary'"
						class="px-4 py-2 rounded flex items-center gap-2 disabled:opacity-50"
				>
					<Icon 
							:name="statusLoading ? 'mdi:loading' : (application.status === 'running' ? 'mdi:stop' : 'mdi:play')" 
							:size="16" 
							:class="statusLoading ? 'animate-spin' : ''" 
					/>
					{{ application.status === 'running' ? 'Stop' : 'Start' }}
				</button>
				
				<button 
						@click="restartApp()"
						:disabled="statusLoading"
						class="px-4 py-2 btn btn-secondary rounded flex items-center gap-2 disabled:opacity-50"
				>
					<Icon name="mdi:restart" size="16" />
					Restart
				</button>
				
				<button 
						@click="updateApp()"
						:disabled="updateLoading"
						class="px-4 py-2 btn btn-info rounded flex items-center gap-2 disabled:opacity-50"
				>
					<Icon 
							:name="updateLoading ? 'mdi:loading' : 'mdi:package-up'" 
							:size="16" 
							:class="updateLoading ? 'animate-spin' : ''" 
					/>
					Update
				</button>
				
				<button 
						@click="showDeleteDialog = true"
						class="px-4 py-2 btn btn-error rounded flex items-center gap-2"
				>
					<Icon name="mdi:delete" size="16" />
					Delete
				</button>
			</div>

			<!-- Application Info Grid -->
			<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
				<!-- Basic Info -->
				<div class="card p-4 bg-base-200">
					<h3 class="font-semibold text-base-content mb-3">Basic Information</h3>
					<div class="space-y-2 text-sm">
						<div class="flex justify-between">
							<span class="text-base-content/60">Provider:</span>
							<span class="text-base-content/80 capitalize">{{ application.provider }}</span>
						</div>
						<div class="flex justify-between">
							<span class="text-base-content/60">Version:</span>
							<span class="text-base-content/80">{{ application.version }}</span>
						</div>
						<div class="flex justify-between">
							<span class="text-base-content/60">Status:</span>
							<span :class="getStatusClass(application.status)" class="px-2 py-1 text-xs rounded-full capitalize">
								{{ application.status }}
							</span>
						</div>
						<div class="flex justify-between">
							<span class="text-base-content/60">Deploy Method:</span>
							<span class="text-base-content/80 capitalize">{{ application.deploy_method }}</span>
						</div>
					</div>
				</div>

				<!-- Technical Stack -->
				<div class="card p-4 bg-base-200">
					<h3 class="font-semibold text-base-content mb-3">Technical Stack</h3>
					<div class="space-y-2 text-sm">
						<div class="flex justify-between">
							<span class="text-base-content/60">Runtime:</span>
							<span class="text-base-content/80">{{ application.runtime }}</span>
						</div>
						<div class="flex justify-between">
							<span class="text-base-content/60">Database:</span>
							<span class="text-base-content/80">{{ application.database }}</span>
						</div>
						<div class="flex justify-between">
							<span class="text-base-content/60">Web Server:</span>
							<span class="text-base-content/80">{{ application.webserver }}</span>
						</div>
						<div class="flex justify-between">
							<span class="text-base-content/60">Process Manager:</span>
							<span class="text-base-content/80">{{ application.process_manager }}</span>
						</div>
					</div>
				</div>

				<!-- Path & Domain -->
				<div class="card p-4 bg-base-200">
					<h3 class="font-semibold text-base-content mb-3">Location</h3>
					<div class="space-y-2 text-sm">
						<div>
							<span class="text-base-content/60 block mb-1">Domain:</span>
							<p class="text-base-content/80 font-mono bg-base-300 p-2 rounded">
								{{ application.domain || 'Not configured' }}
							</p>
						</div>
						<div>
							<span class="text-base-content/60 block mb-1">Path:</span>
							<p class="text-base-content/80 font-mono text-xs bg-base-300 p-2 rounded">
								{{ application.path }}
							</p>
						</div>
					</div>
				</div>
			</div>

			<!-- Application Commands -->
			<div v-if="commands && Object.keys(commands).length" class="card p-4 bg-base-200">
				<h3 class="font-semibold text-base-content mb-4">Application Commands</h3>
				<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
					<div 
							v-for="(cmd, key) in commands" 
							:key="key"
							@click="executeCommand(key, cmd)"
							class="p-3 border border-base-300 rounded hover:bg-base-300 cursor-pointer transition-colors"
					>
						<div class="flex items-center gap-2 mb-2">
							<Icon :name="getCommandIcon(key)" size="20" class="text-primary-500" />
							<h4 class="font-medium text-base-content">{{ cmd.label }}</h4>
						</div>
						<p class="text-xs text-base-content/60">{{ cmd.help }}</p>
					</div>
				</div>
			</div>

			<!-- File Management -->
			<div v-if="editableFiles && editableFiles.length" class="card p-4 bg-base-200">
				<div class="flex items-center justify-between mb-4">
					<h3 class="font-semibold text-base-content">Configuration Files</h3>
					<button 
							@click="loadEditableFiles()"
							:disabled="filesLoading"
							class="px-3 py-1 text-xs btn btn-secondary rounded flex items-center gap-1 disabled:opacity-50"
					>
						<Icon 
								:name="filesLoading ? 'mdi:loading' : 'mdi:refresh'" 
								:size="14" 
								:class="filesLoading ? 'animate-spin' : ''" 
						/>
						Refresh
					</button>
				</div>
				<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
					<div 
							v-for="file in editableFiles" 
							:key="file.path"
							@click="openFileEditor(file)"
							class="p-3 border border-base-300 rounded hover:bg-base-300 cursor-pointer transition-colors"
					>
						<div class="flex items-center gap-2 mb-2">
							<Icon :name="getFileIcon(file.path)" size="20" class="text-primary-500" />
							<h4 class="font-medium text-base-content">{{ file.path }}</h4>
						</div>
						<p class="text-xs text-base-content/60">{{ file.description }}</p>
					</div>
				</div>
			</div>

			<!-- Status & Logs -->
			<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
				<!-- Application Status -->
				<div class="card p-4 bg-base-200">
					<h3 class="font-semibold text-base-content mb-3">Application Status</h3>
					<div v-if="statusLoading" class="text-center py-4">
						<Icon name="mdi:loading" size="24" class="animate-spin text-base-content/60 mx-auto mb-2" />
						<p class="text-base-content/60 text-sm">Checking status...</p>
					</div>
					<div v-else-if="appStatus" class="space-y-2 text-sm">
						<div class="flex justify-between">
							<span class="text-base-content/60">Status:</span>
							<span :class="appStatus.is_running ? 'text-green-500' : 'text-red-500'">
								{{ appStatus.is_running ? 'Running' : 'Stopped' }}
							</span>
						</div>
						<div class="flex justify-between">
							<span class="text-base-content/60">Healthy:</span>
							<span :class="appStatus.is_healthy ? 'text-green-500' : 'text-red-500'">
								{{ appStatus.is_healthy ? 'Yes' : 'No' }}
							</span>
						</div>
						<div class="flex justify-between">
							<span class="text-base-content/60">Last Checked:</span>
							<span class="text-base-content/80">{{ formatDate(appStatus.last_checked) }}</span>
						</div>
						<div v-if="appStatus.uptime" class="flex justify-between">
							<span class="text-base-content/60">Uptime:</span>
							<span class="text-base-content/80">{{ appStatus.uptime }}</span>
						</div>
					</div>
				</div>

				<!-- Recent Logs -->
				<div class="card p-4 bg-base-200">
					<div class="flex items-center justify-between mb-3">
						<h3 class="font-semibold text-base-content">Recent Logs</h3>
						<button 
								@click="loadLogs()"
								:disabled="logsLoading"
								class="px-3 py-1 text-xs btn btn-secondary rounded flex items-center gap-1 disabled:opacity-50"
						>
							<Icon 
									:name="logsLoading ? 'mdi:loading' : 'mdi:refresh'" 
									:size="14" 
									:class="logsLoading ? 'animate-spin' : ''" 
							/>
							Refresh
						</button>
					</div>
					<div v-if="logsLoading" class="text-center py-4">
						<Icon name="mdi:loading" size="20" class="animate-spin text-base-content/60 mx-auto mb-2" />
						<p class="text-base-content/60 text-sm">Loading logs...</p>
					</div>
					<div v-else-if="logs" class="space-y-1">
						<div 
								v-for="(log, index) in logs.slice(0, 10)" 
								:key="index"
								class="text-xs font-mono bg-base-300 p-2 rounded border-l-4"
								:class="getLogLevelClass(log)"
						>
							{{ log }}
						</div>
					</div>
					<div v-else class="text-center py-4">
						<p class="text-base-content/60 text-sm">No logs available</p>
					</div>
				</div>
			</div>
		</div>

		<!-- Error State -->
		<div v-else class="text-center py-12">
			<Icon name="mdi:alert-circle" size="64" class="text-red-500 mx-auto mb-4" />
			<h3 class="text-red-500 text-lg font-medium mb-2">Application Not Found</h3>
			<p class="text-base-content/60 mb-4">The application you're looking for doesn't exist or you don't have permission to view it.</p>
			<button @click="navigateTo('/applications')" class="px-6 py-2 btn btn-primary rounded-md">
				Back to Applications
			</button>
		</div>
	</div>

	<!-- Delete Confirmation Dialog -->
	<NativeDialog v-model="showDeleteDialog">
			<div class="space-y-4">
				<h2 class="text-base-content font-semibold text-lg">Delete Application</h2>
				<p class="text-base-content/60">Are you sure you want to delete this application? This action cannot be undone.</p>
				
				<div class="space-y-2">
					<label class="flex items-center gap-2">
						<input 
								type="checkbox" 
								v-model="deleteKeepData" 
								class="rounded border-base-300 text-primary-500 focus:ring-primary-500"
						>
						<span class="text-base-content/80">Keep database and configuration files</span>
					</label>
				</div>

				<div class="flex justify-end gap-2 pt-4">
					<button @click="showDeleteDialog = false" 
							class="px-4 py-2 text-base-content/70 hover:text-base-content transition-colors">
						Cancel
					</button>
					<button 
							@click="deleteApplication()"
							:disabled="deleteLoading"
							class="px-4 py-2 btn btn-error rounded flex items-center gap-2 disabled:opacity-50"
					>
						<Icon 
								:name="deleteLoading ? 'mdi:loading' : 'mdi:delete'" 
								:size="16" 
								:class="deleteLoading ? 'animate-spin' : ''" 
						/>
						Delete Application
					</button>
				</div>
			</div>
		</NativeDialog>

	<!-- Command Execution Dialog -->
	<NativeDialog v-model="showCommandDialog">
			<div class="space-y-4">
				<h2 class="text-base-content font-semibold text-lg">
					Execute: {{ currentCommand?.label }}
				</h2>
				<p class="text-base-content/60 text-sm">{{ currentCommand?.help }}</p>
				
				<!-- Input for arguments -->
				<div v-if="!commandOutput && !commandLoading && currentCommand?.args && currentCommand.args.length" class="space-y-4">
					<div v-for="argKey in currentCommand.args" :key="argKey" class="space-y-2 text-left">
						<label class="block text-sm font-medium text-base-content/80">
							{{ argKey }}
						</label>
						<input 
							type="text" 
							v-model="commandArgsMap[argKey]" 
							class="w-full p-2 bg-base-300 border border-base-300 rounded text-sm text-base-content focus:outline-none focus:ring-1 focus:ring-primary-500" 
							:placeholder="`Enter ${argKey}...`"
							@keyup.enter="runCommandExecution()"
						/>
					</div>
				</div>

				<div v-if="commandLoading" class="text-center py-4">
					<Icon name="mdi:loading" size="32" class="animate-spin text-base-content/60 mx-auto mb-4" />
					<p class="text-base-content/60">Executing command...</p>
				</div>
				
				<div v-else-if="commandOutput" class="space-y-3">
					<div>
						<h4 class="font-medium text-base-content mb-2">Output:</h4>
						<pre class="bg-base-300 p-3 rounded text-xs font-mono overflow-x-auto max-h-96">{{ commandOutput }}</pre>
					</div>
				</div>
			</div>
			
			<div class="flex justify-end pt-4 border-t border-base-300 gap-2">
				<button 
					v-if="currentCommand?.args && currentCommand.args.length && !commandOutput && !commandLoading"
					@click="runCommandExecution()" 
					class="px-4 py-2 btn btn-primary rounded"
				>
					Run Command
				</button>
				<button @click="showCommandDialog = false" 
						class="px-4 py-2 btn btn-secondary rounded">
					Close
				</button>
			</div>
		</NativeDialog>

	<!-- File Editor Dialog -->
	<NativeDialog v-model="showFileEditorDialog">
		<div class="space-y-4">
			<h2 class="text-base-content font-semibold text-lg">
				Edit: {{ currentFile?.path }}
			</h2>
			
			<div v-if="fileSaving" class="text-center py-4">
				<Icon name="mdi:loading" size="32" class="animate-spin text-base-content/60 mx-auto mb-4" />
				<p class="text-base-content/60">Saving file...</p>
			</div>
			
			<div v-else class="space-y-3">
				<div>
					<h4 class="font-medium text-base-content mb-2">File Content:</h4>
					<textarea
							v-model="fileContent"
							class="w-full h-96 p-3 bg-base-300 border border-base-300 rounded font-mono text-sm"
							:placeholder="`Editing ${currentFile?.path}...`"
					></textarea>
				</div>
			</div>
		</div>
		
		<div class="flex justify-end pt-4 border-t border-base-300 gap-2">
			<button @click="showFileEditorDialog = false" 
					class="px-4 py-2 text-base-content/70 hover:text-base-content transition-colors">
				Cancel
			</button>
			<button 
					@click="saveFile()"
					:disabled="fileSaving"
					class="px-4 py-2 btn btn-primary rounded flex items-center gap-2 disabled:opacity-50"
			>
				<Icon 
						:name="fileSaving ? 'mdi:loading' : 'mdi:content-save'" 
						:size="16" 
						:class="fileSaving ? 'animate-spin' : ''" 
				/>
				Save File
			</button>
		</div>
	</NativeDialog>
</template>

<script setup>
const application = ref(null)
useHead({
  title: () => application.value?.display_name ? `App: ${application.value.display_name}` : 'Application Details'
})
const commands = ref(null)
const editableFiles = ref([])
const appStatus = ref(null)
const logs = ref([])
const loading = ref(true)
const statusLoading = ref(false)
const updateLoading = ref(false)
const logsLoading = ref(false)
const filesLoading = ref(false)
const deleteLoading = ref(false)
const commandLoading = ref(false)
const commandOutput = ref('')
const commandArgsMap = ref({})
const showDeleteDialog = ref(false)
const showCommandDialog = ref(false)
const showFileEditorDialog = ref(false)
const currentCommand = ref(null)
const currentFile = ref(null)
const fileContent = ref('')
const fileSaving = ref(false)
const deleteKeepData = ref(false)

const snackbar = useSnackbar()
const route = useRoute()
const router = useRouter()

// Load application details
const loadApplication = async () => {
	const appId = route.params.id
	if (!appId) return

	loading.value = true
	try {
		const response = await useNuxtApp().$axiosApi.get(`/applications/${appId}`)
		application.value = response.data.data
		commands.value = response.data.commands || null
	} catch (error) {
		console.error('Failed to load application:', error)
		snackbar.add({
			type: "error",
			text: "Failed to load application details"
		})
	} finally {
		loading.value = false
	}
}

// Load editable files
const loadEditableFiles = async () => {
	if (!application.value) return

	filesLoading.value = true
	try {
		const response = await useNuxtApp().$axiosApi.get(`/applications/${application.value.id}/files`)
		editableFiles.value = response.data.files || []
	} catch (error) {
		console.error('Failed to load editable files:', error)
		snackbar.add({
			type: "error",
			text: "Failed to load editable files"
		})
	} finally {
		filesLoading.value = false
	}
}

// Open file editor
const openFileEditor = async (file) => {
	if (!application.value || !file) return

	currentFile.value = file
	fileContent.value = ''
	
	// Construct full file path
	const fullPath = `${application.value.path}/${file.path}`
	
	try {
		const response = await useNuxtApp().$axiosApi.post('/filemanager/read', {
			path: fullPath
		})
		fileContent.value = response.data || ''
		showFileEditorDialog.value = true
	} catch (error) {
		console.error('Failed to load file:', error)
		snackbar.add({
			type: "error",
			text: `Failed to load ${file.path}`
		})
	}
}

// Save file
const saveFile = async () => {
	if (!application.value || !currentFile.value) return

	fileSaving.value = true
	
	// Construct full file path
	const fullPath = `${application.value.path}/${currentFile.value.path}`
	
	try {
		const response = await useNuxtApp().$axiosApi.post('/filemanager/write', {
			path: fullPath,
			content: fileContent.value
		})
		
		if (response.data.message) {
			snackbar.add({
				type: "success",
				text: `${currentFile.value.path} saved successfully`
			})
			showFileEditorDialog.value = false
		} else {
			throw new Error(response.data.error || "Failed to save file")
		}
	} catch (error) {
		console.error('Failed to save file:', error)
		snackbar.add({
			type: "error",
			text: `Failed to save ${currentFile.value.path}`
		})
	} finally {
		fileSaving.value = false
	}
}

// Load application status
const loadStatus = async () => {
	if (!application.value) return

	statusLoading.value = true
	try {
		const response = await useNuxtApp().$axiosApi.get(`/applications/${application.value.id}/status`)
		appStatus.value = response.data.data
	} catch (error) {
		console.error('Failed to load status:', error)
	} finally {
		statusLoading.value = false
	}
}

// Load application logs
const loadLogs = async () => {
	if (!application.value) return

	logsLoading.value = true
	try {
		const response = await useNuxtApp().$axiosApi.get(`/applications/${application.value.id}/logs`)
		logs.value = response.data.logs ? response.data.logs.split('\n') : []
	} catch (error) {
		console.error('Failed to load logs:', error)
		snackbar.add({
			type: "error",
			text: "Failed to load logs"
		})
	} finally {
		logsLoading.value = false
	}
}

// Application actions
const toggleAppStatus = async () => {
	if (!application.value || statusLoading.value) return

	statusLoading.value = true
	const action = application.value.status === 'running' ? 'stop' : 'start'

	try {
		const response = await useNuxtApp().$axiosApi.post(`/applications/${application.value.id}/${action}`)
		
		if (response.data.success) {
			application.value.status = action === 'start' ? 'running' : 'stopped'
			await loadStatus() // Refresh status
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
		statusLoading.value = false
	}
}

const restartApp = async () => {
	if (!application.value || statusLoading.value) return

	statusLoading.value = true
	try {
		const response = await useNuxtApp().$axiosApi.post(`/applications/${application.value.id}/restart`)
		
		if (response.data.success) {
			application.value.status = 'running'
			await loadStatus() // Refresh status
			snackbar.add({
				type: "success",
				text: "Application restarted successfully"
			})
		} else {
			throw new Error(response.data.message || "Failed to restart application")
		}
	} catch (error) {
		console.error('Failed to restart application:', error)
		snackbar.add({
			type: "error",
			text: "Failed to restart application"
		})
	} finally {
		statusLoading.value = false
	}
}

const updateApp = async () => {
	if (!application.value || updateLoading.value) return

	updateLoading.value = true
	try {
		const response = await useNuxtApp().$axiosApi.post(`/applications/${application.value.id}/update`)
		
		if (response.data.success) {
			application.value.status = 'updated'
			snackbar.add({
				type: "success",
				text: "Application updated successfully"
			})
		} else {
			throw new Error(response.data.message || "Failed to update application")
		}
	} catch (error) {
		console.error('Failed to update application:', error)
		snackbar.add({
			type: "error",
			text: "Failed to update application"
		})
	} finally {
		updateLoading.value = false
	}
}

const deleteApplication = async () => {
	if (!application.value || deleteLoading.value) return

	deleteLoading.value = true
	try {
		const response = await useNuxtApp().$axiosApi.delete(`/applications/${application.value.id}`)
		
		if (response.data.success) {
			snackbar.add({
				type: "success",
				text: "Application deleted successfully"
			})
			navigateTo('/applications')
		} else {
			throw new Error(response.data.message || "Failed to delete application")
		}
	} catch (error) {
		console.error('Failed to delete application:', error)
		snackbar.add({
			type: "error",
			text: "Failed to delete application"
		})
	} finally {
		deleteLoading.value = false
		showDeleteDialog.value = false
	}
}

const executeCommand = async (commandKey, command) => {
	if (!application.value || commandLoading.value) return

	currentCommand.value = { ...command, key: commandKey }
	commandOutput.value = ''
	commandArgsMap.value = {}

	// Initialize keys in commandArgsMap
	if (command.args && command.args.length) {
		command.args.forEach(key => {
			commandArgsMap.value[key] = ''
		})
		showCommandDialog.value = true
	} else {
		showCommandDialog.value = true // Show dialog to display loading state
		await runCommandExecution()
	}
}

const runCommandExecution = async () => {
	if (!application.value || !currentCommand.value || commandLoading.value) return

	// Validation check for required array args
	if (currentCommand.value.args && currentCommand.value.args.length) {
		for (const key of currentCommand.value.args) {
			if (!commandArgsMap.value[key]) {
				snackbar.add({
					type: "error",
					text: `"${key}" is required.`
				})
				return
			}
		}
	}

	commandLoading.value = true
	try {
		const payload = {
			command: currentCommand.value.key,
			args: commandArgsMap.value
		}
		const response = await useNuxtApp().$axiosApi.post(`/applications/${application.value.id}/command`, payload)
		
		if (response.data.success) {
			commandOutput.value = response.data.output
			snackbar.add({
				type: "success",
				text: `Command "${currentCommand.value.label}" executed successfully`
			})
		} else {
			throw new Error(response.data.message || `Failed to execute command`)
		}
	} catch (error) {
		console.error('Failed to execute command:', error)
		commandOutput.value = error.response?.data?.output || error.response?.data?.details || error.response?.data?.error || error.message
		snackbar.add({
			type: "error",
			text: `Failed to execute command`
		})
	} finally {
		commandLoading.value = false
	}
}

// Utility functions
const getAppIcon = (provider) => {
	const iconMap = {
		'wordpress': 'skill-icons:wordpress',
		'laravel': 'skill-icons:laravel-dark',
		'php': 'skill-icons:php-dark',
		'nodejs': 'skill-icons:nodejs-dark',
		'golang': 'skill-icons:golang',
		'python': 'skill-icons:python-dark',
		'ghost': 'skill-icons:ghost',
		'static': 'mdi:language-html5',
		'umami': 'skill-icons:javascript',
		'erpnext': 'skill-icons:javascript'
	}
	return iconMap[provider?.toLowerCase()] || 'mdi:application-outline'
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

const getCommandIcon = (commandKey) => {
	const iconMap = {
		'backup': 'mdi:database-export',
		'optimize': 'mdi:tune',
		'cache_clear': 'mdi:broom',
		'migrate': 'mdi:database-sync',
		'deploy': 'mdi:rocket-launch',
		'build': 'mdi:hammer',
		'test': 'mdi:test-tube',
		'restart': 'mdi:restart',
		'stop': 'mdi:stop',
		'start': 'mdi:play'
	}
	return iconMap[commandKey] || 'mdi:cog'
}

// Use default file icon for all files (no specific mapping)
const getFileIcon = () => {
	return 'mdi:file-document'
}

const getLogLevelClass = (log) => {
	const lowerLog = log.toLowerCase()
	if (lowerLog.includes('error')) return 'border-red-500'
	if (lowerLog.includes('warn')) return 'border-yellow-500'
	if (lowerLog.includes('info')) return 'border-blue-500'
	return 'border-gray-500'
}

const formatDate = (dateString) => {
	if (!dateString) return 'Never'
	return new Date(dateString).toLocaleString()
}

// Load on mount
onMounted(() => {
	loadApplication()
	loadStatus()
	loadLogs()
	loadEditableFiles()
})
</script>

<style scoped>
.card {
	transition: all 0.2s ease;
}

.card:hover {
	box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
}

.animate-spin {
	animation: spin 1s linear infinite;
}

@keyframes spin {
	from { transform: rotate(0deg); }
	to { transform: rotate(360deg); }
}
</style>
