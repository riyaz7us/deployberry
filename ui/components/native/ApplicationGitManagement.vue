<template>
  <div class="bg-base-200 border border-base-300 rounded-lg p-6">
    <div class="flex items-center justify-between mb-6">
      <div class="flex items-center gap-3">
        <Icon name="mdi:git" size="24" class="text-orange-500" />
        <h3 class="text-lg font-semibold text-base-content">Version Control</h3>
      </div>
      <NativeButton
        @click="refreshStatus"
        variant="secondary"
        size="sm"
        :loading="loading"
        icon="mdi:refresh"
      >
        Refresh
      </NativeButton>
    </div>

    <!-- Git Status Overview -->
    <div v-if="gitStatus.isGitRepo" class="mb-6">
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-4">
        <div class="bg-base-100 p-4 rounded-lg border border-base-300">
          <div class="flex items-center gap-2 mb-2">
            <Icon name="mdi:source-branch" class="text-primary" size="16" />
            <span class="text-sm text-base-content/70">Branch</span>
          </div>
          <p class="text-base-content font-medium">{{ gitStatus.currentBranch || 'Unknown' }}</p>
        </div>

        <div class="bg-base-100 p-4 rounded-lg border border-base-300">
          <div class="flex items-center gap-2 mb-2">
            <Icon name="mdi:git" class="text-success" size="16" />
            <span class="text-sm text-base-content/70">Remote</span>
          </div>
          <p class="text-base-content font-medium text-xs">{{ gitStatus.remoteURL || 'No remote' }}</p>
        </div>

        <div class="bg-base-100 p-4 rounded-lg border border-base-300">
          <div class="flex items-center gap-2 mb-2">
            <Icon name="mdi:arrow-up" class="text-warning" size="16" />
            <span class="text-sm text-base-content/70">Ahead</span>
          </div>
          <p class="text-base-content font-medium">{{ gitStatus.ahead }}</p>
        </div>

        <div class="bg-base-100 p-4 rounded-lg border border-base-300">
          <div class="flex items-center gap-2 mb-2">
            <Icon name="mdi:arrow-down" class="text-primary" size="16" />
            <span class="text-sm text-base-content/70">Behind</span>
          </div>
          <p class="text-base-content font-medium">{{ gitStatus.behind }}</p>
        </div>
      </div>

      <!-- Changes Indicator -->
      <div v-if="gitStatus.changes && gitStatus.changes.length > 0" class="mb-4 p-3 bg-warning/10 border border-warning/20 rounded-lg">
        <div class="flex items-center gap-2">
          <Icon name="mdi:alert-circle" class="text-warning" size="16" />
          <span class="text-warning text-sm">
            {{ gitStatus.changes.length }} uncommitted change(s)
          </span>
        </div>
      </div>
    </div>

    <!-- Not a Git Repository -->
    <div v-else class="mb-6 p-4 bg-base-100 border border-base-300 rounded-lg">
      <div class="flex items-center gap-2">
        <Icon name="mdi:git" class="text-base-content/60" size="20" />
        <span class="text-base-content/60">This application is not under version control</span>
      </div>
      <div class="mt-3">
        <NativeButton
          @click="initializeGit"
          variant="primary"
          size="sm"
          :loading="operationLoading === 'init'"
          icon="mdi:git"
        >
          Initialize Git Repository
        </NativeButton>
      </div>
    </div>

    <!-- Git Operations -->
    <div v-if="gitStatus.isGitRepo" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <!-- Basic Operations -->
      <div class="space-y-3">
        <h4 class="text-sm font-medium text-base-content/70 mb-3">Basic Operations</h4>

        <NativeButton
          @click="executeGitOperation('status')"
          variant="secondary"
          size="sm"
          :loading="operationLoading === 'status'"
          icon="mdi:information-outline"
          class="w-full justify-start"
        >
          Status
        </NativeButton>

        <NativeButton
          @click="executeGitOperation('log')"
          variant="secondary"
          size="sm"
          :loading="operationLoading === 'log'"
          icon="mdi:history"
          class="w-full justify-start"
        >
          Log
        </NativeButton>

        <NativeButton
          @click="executeGitOperation('fetch')"
          variant="secondary"
          size="sm"
          :loading="operationLoading === 'fetch'"
          icon="mdi:download"
          class="w-full justify-start"
        >
          Fetch
        </NativeButton>
      </div>

      <!-- Pull Operations -->
      <div class="space-y-3">
        <h4 class="text-sm font-medium text-base-content/70 mb-3">Pull Operations</h4>

        <NativeButton
          @click="executeGitOperation('pull')"
          variant="primary"
          size="sm"
          :loading="operationLoading === 'pull'"
          icon="mdi:arrow-down-bold"
          class="w-full justify-start"
        >
          Pull
        </NativeButton>

        <NativeButton
          @click="executeGitOperation('force_pull')"
          variant="warning"
          size="sm"
          :loading="operationLoading === 'force_pull'"
          icon="mdi:arrow-down-bold-box"
          class="w-full justify-start"
        >
          Force Pull
        </NativeButton>

        <NativeButton
          @click="executeGitOperation('branch')"
          variant="secondary"
          size="sm"
          :loading="operationLoading === 'branch'"
          icon="mdi:source-branch"
          class="w-full justify-start"
        >
          Branches
        </NativeButton>
      </div>

      <!-- Stash Operations -->
      <div class="space-y-3">
        <h4 class="text-sm font-medium text-base-content/70 mb-3">Stash Operations</h4>

        <NativeButton
          @click="executeGitOperation('stash')"
          variant="success"
          size="sm"
          :loading="operationLoading === 'stash'"
          icon="mdi:package-variant"
          class="w-full justify-start"
        >
          Stash
        </NativeButton>

        <NativeButton
          @click="executeGitOperation('apply_stash')"
          variant="success"
          size="sm"
          :loading="operationLoading === 'apply_stash'"
          icon="mdi:package-variant-closed"
          class="w-full justify-start"
        >
          Apply Stash
        </NativeButton>

        <NativeButton
          @click="executeGitOperation('reset_hard')"
          variant="error"
          size="sm"
          :loading="operationLoading === 'reset_hard'"
          icon="mdi:undo-variant"
          class="w-full justify-start"
        >
          Reset Hard
        </NativeButton>
      </div>
    </div>

    <!-- Operation Output -->
    <div v-if="lastOutput" class="mt-6">
      <div class="flex items-center justify-between mb-2">
        <h4 class="text-sm font-medium text-base-content/70">Last Operation Output</h4>
        <NativeButton
          @click="lastOutput = ''"
          variant="secondary"
          size="sm"
          icon="mdi:close"
        >
          Clear
        </NativeButton>
      </div>
      <div class="bg-base-100 p-4 rounded-lg border border-base-300">
        <pre class="text-base-content text-sm whitespace-pre-wrap font-mono">{{ lastOutput }}</pre>
      </div>
    </div>

    <!-- Branch Checkout -->
    <div v-if="showBranchInput" class="mt-6 p-4 bg-base-100 border border-base-300 rounded-lg">
      <h4 class="text-sm font-medium text-base-content/70 mb-3">Checkout Branch</h4>
      <div class="flex gap-2">
        <NativeTextField
          v-model="checkoutBranch"
          placeholder="Branch name"
          class="flex-1"
        />
        <NativeButton
          @click="executeGitOperation('checkout', checkoutBranch)"
          variant="primary"
          size="sm"
          :loading="operationLoading === 'checkout'"
          icon="mdi:check"
        >
          Checkout
        </NativeButton>
        <NativeButton
          @click="showBranchInput = false"
          variant="secondary"
          size="sm"
          icon="mdi:close"
        >
          Cancel
        </NativeButton>
      </div>
    </div>
  </div>
</template>

<script setup>
const props = defineProps({
  appPath: {
    type: String,
    required: true
  },
  appName: {
    type: String,
    default: 'Application'
  }
})

const emit = defineEmits(['operation-completed', 'error'])

const gitStatus = ref({})
const operationLoading = ref('')
const lastOutput = ref('')
const showBranchInput = ref(false)
const checkoutBranch = ref('')
const loading = ref(false)

const snackbar = useSnackbar()

onMounted(() => {
  refreshStatus()
})

const refreshStatus = async () => {
  loading.value = true
  try {
    const response = await useNuxtApp().$axiosApi.get(`/git/status?path=${encodeURIComponent(props.appPath)}`)
    gitStatus.value = response.data
  } catch (error) {
    console.error('Failed to get git status:', error)
    emit('error', error)
  } finally {
    loading.value = false
  }
}

const initializeGit = async () => {
  operationLoading.value = 'init'
  lastOutput.value = ''

  try {
    const response = await useNuxtApp().$axiosApi.post('/git/operation', {
      path: props.appPath,
      action: 'init'
    })

    if (response.data.success) {
      lastOutput.value = response.data.output || response.data.message
      snackbar.add({
        type: "success",
        text: response.data.message || 'Git repository initialized successfully'
      })

      // Refresh status to update the UI
      await refreshStatus()

      emit('operation-completed', { action: 'init', result: response.data })
    } else {
      throw new Error(response.data.error || response.data.message)
    }
  } catch (error) {
    console.error('Git init failed:', error)
    const errorMessage = error.response?.data?.error || error.response?.data?.message || error.message || 'Git init failed'
    lastOutput.value = errorMessage

    snackbar.add({
      type: "error",
      text: errorMessage
    })

    emit('error', error)
  } finally {
    operationLoading.value = ''
  }
}

const executeGitOperation = async (action, branch = '') => {
  operationLoading.value = action
  lastOutput.value = ''

  try {
    const operationData = {
      path: props.appPath,
      action: action
    }

    if (branch) {
      operationData.branch = branch
    }

    const response = await useNuxtApp().$axiosApi.post('/git/operation', operationData)

    if (response.data.success) {
      lastOutput.value = response.data.output || response.data.message

      // Show success message
      snackbar.add({
        type: "success",
        text: response.data.message || `${action} completed successfully`
      })

      // Refresh status for certain operations
      if (['pull', 'force_pull', 'checkout', 'reset_hard', 'init'].includes(action)) {
        await refreshStatus()
      }

      emit('operation-completed', { action, result: response.data })
    } else {
      throw new Error(response.data.error || response.data.message)
    }
  } catch (error) {
    console.error(`Git ${action} failed:`, error)

    const errorMessage = error.response?.data?.error || error.response?.data?.message || error.message || `Git ${action} failed`
    lastOutput.value = errorMessage

    snackbar.add({
      type: "error",
      text: errorMessage
    })

    emit('error', error)
  } finally {
    operationLoading.value = ''
  }
}

// Expose methods for parent components
defineExpose({
  refreshStatus,
  executeGitOperation
})
</script>
