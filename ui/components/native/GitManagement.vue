<template>
  <div class="bg-slate-800/50 border border-slate-700 rounded-lg p-6">
    <div class="flex items-center justify-between mb-6">
      <div class="flex items-center gap-3">
        <Icon name="mdi:git" size="24" class="text-orange-500" />
        <h3 class="text-lg font-semibold text-white">Git Repository Management</h3>
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
        <div class="bg-slate-700/50 p-4 rounded-lg">
          <div class="flex items-center gap-2 mb-2">
            <Icon name="mdi:source-branch" class="text-blue-400" size="16" />
            <span class="text-sm text-slate-400">Branch</span>
          </div>
          <p class="text-white font-medium">{{ gitStatus.currentBranch || 'Unknown' }}</p>
        </div>

        <div class="bg-slate-700/50 p-4 rounded-lg">
          <div class="flex items-center gap-2 mb-2">
            <Icon name="mdi:git" class="text-green-400" size="16" />
            <span class="text-sm text-slate-400">Remote</span>
          </div>
          <p class="text-white font-medium text-xs">{{ gitStatus.remoteURL || 'No remote' }}</p>
        </div>

        <div class="bg-slate-700/50 p-4 rounded-lg">
          <div class="flex items-center gap-2 mb-2">
            <Icon name="mdi:arrow-up" class="text-orange-400" size="16" />
            <span class="text-sm text-slate-400">Ahead</span>
          </div>
          <p class="text-white font-medium">{{ gitStatus.ahead }}</p>
        </div>

        <div class="bg-slate-700/50 p-4 rounded-lg">
          <div class="flex items-center gap-2 mb-2">
            <Icon name="mdi:arrow-down" class="text-blue-400" size="16" />
            <span class="text-sm text-slate-400">Behind</span>
          </div>
          <p class="text-white font-medium">{{ gitStatus.behind }}</p>
        </div>
      </div>

      <!-- Changes Indicator -->
      <div v-if="gitStatus.changes && gitStatus.changes.length > 0" class="mb-4 p-3 bg-yellow-500/10 border border-yellow-500/20 rounded-lg">
        <div class="flex items-center gap-2">
          <Icon name="mdi:alert-circle" class="text-yellow-400" size="16" />
          <span class="text-yellow-400 text-sm">
            {{ gitStatus.changes.length }} uncommitted change(s)
          </span>
        </div>
      </div>
    </div>

    <!-- Not a Git Repository -->
    <div v-else class="mb-6 p-4 bg-red-500/10 border border-red-500/20 rounded-lg">
      <div class="flex items-center gap-2">
        <Icon name="mdi:git" class="text-red-400" size="20" />
        <span class="text-red-400">This directory is not a Git repository</span>
      </div>
    </div>

    <!-- Git Operations -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <!-- Basic Operations -->
      <div class="space-y-3">
        <h4 class="text-sm font-medium text-slate-300 mb-3">Basic Operations</h4>

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
        <h4 class="text-sm font-medium text-slate-300 mb-3">Pull Operations</h4>

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
        <h4 class="text-sm font-medium text-slate-300 mb-3">Stash Operations</h4>

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
          variant="danger"
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
      <h4 class="text-sm font-medium text-slate-300 mb-2">Last Operation Output</h4>
      <div class="bg-slate-900 p-4 rounded-lg border border-slate-600">
        <pre class="text-green-400 text-sm whitespace-pre-wrap font-mono">{{ lastOutput }}</pre>
      </div>
    </div>

    <!-- Branch Checkout -->
    <div v-if="showBranchInput" class="mt-6 p-4 bg-slate-700/30 border border-slate-600 rounded-lg">
      <h4 class="text-sm font-medium text-slate-300 mb-3">Checkout Branch</h4>
      <div class="flex gap-2">
        <input
          v-model="checkoutBranch"
          type="text"
          placeholder="Branch name"
          class="flex-1 px-3 py-2 bg-slate-800 border border-slate-600 rounded-lg text-white placeholder-slate-400 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20"
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
  path: {
    type: String,
    required: true
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
    const response = await useNuxtApp().$axiosApi.get(`/git/status?path=${encodeURIComponent(props.path)}`)
    gitStatus.value = response.data
  } catch (error) {
    console.error('Failed to get git status:', error)
    emit('error', error)
  } finally {
    loading.value = false
  }
}

const executeGitOperation = async (action, branch = '') => {
  operationLoading.value = action
  lastOutput.value = ''

  try {
    const operationData = {
      path: props.path,
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
      if (['pull', 'force_pull', 'checkout', 'reset_hard'].includes(action)) {
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
