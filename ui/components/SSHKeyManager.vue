<template>
  <div>
    <!-- SSH Key Manager Button -->
    <NativeButton
      @click="showDialog = true"
      icon="mdi:key-variant"
      variant="secondary"
      size="sm"
    >
      SSH Keys
    </NativeButton>

    <!-- SSH Key Dialog -->
    <NativeDialog v-model="showDialog">
      <div class="space-y-6">
        <!-- Header -->
        <div class="text-center">
          <Icon name="mdi:key-variant" size="48" class="text-primary mx-auto mb-4" />
          <h2 class="text-2xl font-bold text-base-content mb-2">SSH Key Management</h2>
          <p class="text-base-content/60">
            {{ keyExists ? 'Your SSH public key' : 'Generate a new SSH key pair' }}
          </p>
        </div>

        <!-- Loading State -->
        <div v-if="loading" class="text-center py-8">
          <Icon name="mdi:loading" class="animate-spin text-primary mx-auto mb-4" size="32" />
          <p class="text-base-content/60">
            {{ keyExists ? 'Loading SSH key...' : 'Generating SSH key...' }}
          </p>
        </div>

        <!-- SSH Key Display -->
        <div v-else-if="sshKey" class="space-y-4">
          <!-- Key Info -->
          <div class="bg-slate-700/30 border border-slate-600 rounded-lg p-4">
            <div class="flex items-center gap-3 mb-3">
              <Icon
                :name="keyExists ? 'mdi:key-variant' : 'mdi:key-plus'"
                class="text-success"
                size="24"
              />
              <div>
                <h3 class="text-base-content font-medium">
                  {{ keyExists ? 'Existing SSH Key' : 'New SSH Key Generated' }}
                </h3>
                <p class="text-base-content/60 text-sm">
                  {{ keyExists ? 'Retrieved from' : 'Saved to' }} {{ keyPath }}
                </p>
              </div>
            </div>
            <div class="text-xs text-base-content/40">
              <p><strong>Algorithm:</strong> RSA (4096-bit)</p>
              <p><strong>Comment:</strong> generated-by-panel17</p>
            </div>
          </div>

          <!-- Public Key Textarea -->
          <NativeTextField
            v-model="publicKey"
            label="Public Key (Add to Git Server)"
            :textarea="true"
            :rows="8"
            :monospace="true"
            readonly
            hint="Copy this public key and add it to your Git server (GitHub, GitLab, etc.)"
          />

          <!-- Action Buttons -->
          <div class="flex flex-col sm:flex-row gap-3 justify-end">
            <NativeButton
              @click="copyToClipboard"
              :loading="copying"
              icon="mdi:content-copy"
              variant="secondary"
            >
              {{ copying ? 'Copying...' : 'Copy Public Key' }}
            </NativeButton>

            <NativeButton
              @click="showDialog = false"
              variant="primary"
            >
              Close
            </NativeButton>
          </div>
        </div>

        <!-- Error State -->
        <div v-else-if="error" class="text-center py-8">
          <Icon name="mdi:alert-circle" class="text-error mx-auto mb-4" size="32" />
          <h3 class="text-base-content font-medium mb-2">Error</h3>
          <p class="text-base-content/60 text-sm">{{ error }}</p>
          <div class="mt-4">
            <NativeButton
              @click="loadSSHKey"
              icon="mdi:refresh"
              variant="secondary"
            >
              Try Again
            </NativeButton>
          </div>
        </div>
      </div>
    </NativeDialog>
  </div>
</template>

<script setup>

const showDialog = ref(false)
const loading = ref(false)
const copying = ref(false)
const error = ref('')
const sshKey = ref(null)
const publicKey = ref('')
const keyExists = ref(false)
const keyPath = ref('')

const snackbar = useSnackbar()

// Load SSH key when dialog opens
watch(showDialog, (newValue) => {
  if (newValue && !sshKey.value) {
    loadSSHKey()
  }
})

// Load SSH key from server
const loadSSHKey = async () => {
  loading.value = true
  error.value = ''

  try {
    const response = await useNuxtApp().$axiosApi.get('/git/ssh-key')

    if (response.data.success) {
      sshKey.value = response.data
      publicKey.value = response.data.public_key
      keyExists.value = response.data.key_exists
      keyPath.value = response.data.key_path

      snackbar.add({
        type: 'success',
        text: response.data.message,
      })
    } else {
      throw new Error(response.data.message || 'Failed to load SSH key')
    }
  } catch (err) {
    error.value = err.response?.data?.error || err.message || 'Failed to load SSH key'
    console.error('SSH key error:', err)
  } finally {
    loading.value = false
  }
}

// Copy public key to clipboard
const copyToClipboard = async () => {
  if (!publicKey.value) return

  copying.value = true

  try {
    await navigator.clipboard.writeText(publicKey.value.trim())

    snackbar.add({
      type: 'success',
      text: 'Public key copied to clipboard!',
    })
  } catch (err) {
    // Fallback for browsers that don't support clipboard API
    try {
      const textArea = document.createElement('textarea')
      textArea.value = publicKey.value.trim()
      document.body.appendChild(textArea)
      textArea.select()
      document.execCommand('copy')
      document.body.removeChild(textArea)

      snackbar.add({
        type: 'success',
        text: 'Public key copied to clipboard!',
      })
    } catch (fallbackErr) {
      snackbar.add({
        type: 'error',
        text: 'Failed to copy to clipboard. Please select and copy manually.',
      })
    }
  } finally {
    copying.value = false
  }
}

// Reset state when dialog closes
watch(showDialog, (newValue) => {
  if (!newValue) {
    error.value = ''
    sshKey.value = null
    publicKey.value = ''
    keyExists.value = false
    keyPath.value = ''
  }
})
</script>

<style scoped>
/* Additional styling if needed */
</style>
