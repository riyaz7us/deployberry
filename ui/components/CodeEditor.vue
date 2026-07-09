<script setup>
const props = defineProps({
  filePath: {
    type: String,
    required: true
  },
  isOpen: {
    type: Boolean,
    required: true
  }
})

const emit = defineEmits(['close', 'save'])

const { $axiosApi } = useNuxtApp()

const content = ref('')
const fileName = ref('')
const fileExtension = ref('')
const isSaving = ref(false)
const editor = ref(null)

const getLanguage = (ext) => {
  const languageMap = {
    'js': 'javascript',
    'ts': 'typescript',
    'py': 'python',
    'html': 'html',
    'css': 'css',
    'json': 'json',
    'md': 'markdown',
    'txt': 'plaintext',
    'xml': 'xml',
    'yaml': 'yaml',
    'yml': 'yaml',
    'sql': 'sql',
    'sh': 'bash',
    'bash': 'bash',
    'php': 'php',
    'java': 'java',
    'cpp': 'cpp',
    'c': 'c',
    'cs': 'csharp',
    'rb': 'ruby',
    'go': 'go',
    'rs': 'rust',
    'swift': 'swift',
    'kt': 'kotlin'
  }
  return languageMap[ext.toLowerCase()] || 'plaintext'
}

const loadFile = async () => {
  try {
    const response = await $axiosApi.post("/filemanager/read", {
      path: props.filePath
    })
    content.value = response.data

    fileName.value = props.filePath.split('/').pop()
    fileExtension.value = fileName.value.split('.').pop()

    // Apply syntax highlighting
    if (editor.value) {
      const highlighted = hljs.highlight(
        content.value,
        { language: getLanguage(fileExtension.value) }
      ).value
      editor.value.innerHTML = highlighted
    }
  } catch (error) {
    console.error('Error loading file:', error)
  }
}

const saveFile = async () => {
  if (isSaving.value) return

  isSaving.value = true
  try {
    // Get content from the contenteditable div
    const currentContent = editor.value ? editor.value.textContent : content.value

    const formData = new FormData()
    const blob = new Blob([currentContent], { type: 'text/plain' })
    formData.append('file', blob, fileName.value)
    formData.append('destination', props.filePath)

    await $axiosApi.post("/files/upload", formData, {
      headers: {
        "Content-Type": "multipart/form-data",
      },
    })

    emit('save')
  } catch (error) {
    console.error('Error saving file:', error)
  } finally {
    isSaving.value = false
  }
}

const handleInput = (e) => {
  // Update content when user types
  content.value = e.target.textContent
}

const handleKeyDown = (e) => {
  if (e.key === 'Tab') {
    e.preventDefault()

    // Insert 2 spaces for indentation
    const selection = window.getSelection()
    const range = selection.getRangeAt(0)

    // Create a text node with 2 spaces
    const spaceNode = document.createTextNode('  ')
    range.insertNode(spaceNode)

    // Move cursor after the inserted spaces
    range.setStartAfter(spaceNode)
    range.setEndAfter(spaceNode)
    selection.removeAllRanges()
    selection.addRange(range)

    // Update content
    content.value = e.target.textContent
  }
}

// Watch for file path changes
watchEffect(() => {
  if (props.isOpen && props.filePath) {
    loadFile()
  }
})
</script>

<template>
  <div v-if="isOpen" class="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50 border-primary border-2">
    <div class="bg-base-100 rounded-lg shadow-lg w-11/12 h-5/6 flex flex-col">
      <!-- Header -->
      <div class="px-4 py-3 border-b border-base-300 flex justify-between items-center">
        <div class="flex items-center space-x-2">
          <Icon name="mdi:file-code" class="text-base-content/70" />
          <h3 class="text-lg font-semibold">{{ fileName }}</h3>
        </div>
        <div class="flex items-center space-x-2">
          <button @click="saveFile"
                  :disabled="isSaving"
                  class="px-4 py-2 btn btn-primary disabled:opacity-50 flex items-center">
            <Icon :name="isSaving ? 'mdi:loading' : 'mdi:content-save'" class="mr-2" />
            {{ isSaving ? 'Saving...' : 'Save' }}
          </button>
          <button @click="$emit('close')"
                  class="p-2 text-base-content/70 hover:text-base-content">
            <Icon name="mdi:close" />
          </button>
        </div>
      </div>
      
      <!-- Editor -->
      <div class="flex-1 overflow-hidden">
        <div
          ref="editor"
          contenteditable="true"
          @input="handleInput"
          @keydown="handleKeyDown"
          class="w-full h-full p-4 font-mono text-sm overflow-auto focus:outline-none min-h-[60vh] bg-base-100 text-base-content"
          style="white-space: pre-wrap; word-wrap: break-word;"
        ></div>
      </div>
    </div>
  </div>
</template>