<template>
  <div>
    <!-- Main Content -->
    <main class="container mx-auto px-4 py-6">
      <!-- Breadcrumbs -->
      <div class="flex flex-wrap items-center gap-1 sm:gap-2 mb-4 text-sm sm:text-base">
        <button @click="curPath='/'" class="text-blue-600 hover:text-blue-800 flex items-center">
          <Icon name="mdi:home" class="text-lg" />
        </button>
        <span class="text-slate-500">/</span>
        <div v-for="crumb in breadcrumbs" :key="crumb.path" class="flex items-center gap-1 sm:gap-2">
          <button @click="curPath=crumb.path" class="text-blue-600 hover:text-blue-800 truncate max-w-[120px] sm:max-w-none">
            {{ crumb.name }}
          </button>
          <span class="text-slate-500">/</span>
        </div>
      </div>

      <!-- Action Buttons -->
      <div class="flex flex-wrap gap-3 mb-6 space-x-0">
        <label class="btn btn-primary px-4 py-2 rounded cursor-pointer inline-flex items-center text-sm">
          <Icon name="mdi:upload" class="mr-2" />
          <span>Upload File</span>
          <input type="file" class="hidden" @change="uploadFile" />
        </label>

        <button @click="showCreateFolder = true" class="btn btn-secondary px-4 py-2 rounded inline-flex items-center text-sm">
          <Icon name="mdi:folder-plus" class="mr-2" />
          <span>New Folder</span>
        </button>
      </div>

      <!-- File List -->
      <div class="rounded-lg shadow bg-base-200 card-bordered overflow-hidden">
        <div class="hidden sm:grid grid-cols-12 gap-4 p-4 font-semibold border-b border-base-300">
          <div class="col-span-6 sm:col-span-5">Name</div>
          <div class="col-span-2">Size</div>
          <div class="col-span-3">Modified</div>
          <div class="col-span-2 text-right">Actions</div>
        </div>
        <div class="text-center py-8 text-slate-500" v-if="!files||files.length==0">
          EMPTY ¯\_(ツ)_/¯
        </div>
        <div v-for="file in files" :key="file.path" class="grid grid-cols-12 gap-2 sm:gap-4 p-3 sm:p-4 hover:bg-base-300 items-center border-b border-base-300/30 last:border-b-0" @dblclick="file.is_directory?curPath=file.path:openEditor(file)">
          <!-- Name -->
          <div class="col-span-8 sm:col-span-5 flex items-center min-w-0">
            <Icon :name="file.is_directory ? 'mdi:folder' : 'mdi:file'" class="mr-3 flex-shrink-0" :class="file.is_directory ? 'text-yellow-400' : 'text-gray-400'" size="20" />
            <span @click="file.is_directory?curPath=file.path:openEditor(file)" class="truncate cursor-pointer font-medium text-base-content">{{ file.name }}</span>
          </div>
          <!-- Size (Desktop) -->
          <div class="hidden sm:block col-span-2 text-base-content/70 text-sm">{{ file.is_directory ? "" : formatSize(file.size) }}</div>
          <!-- Modified (Desktop) -->
          <div class="hidden sm:block col-span-3 text-base-content/60 text-sm">{{ formatDate(file.modified) }}</div>
          <!-- Actions -->
          <div class="col-span-4 sm:col-span-2 flex justify-end gap-1 sm:gap-2">
            <button v-if="isEditableFile(file.name)" @click="openEditor(file)" class="text-blue-600 hover:text-blue-800 p-1 rounded hover:bg-base-300 transition-colors" title="Edit">
              <Icon name="mdi:pencil-box" size="20" />
            </button>
            <button v-if="file.is_directory" @click="curPath=file.path" class="text-blue-600 hover:text-blue-800 p-1 rounded hover:bg-base-300 transition-colors" title="Open Folder">
              <Icon name="mdi:folder-open" size="20" />
            </button>
            <button v-else @click="downloadFile(file)" class="text-blue-600 hover:text-blue-800 p-1 rounded hover:bg-base-300 transition-colors" title="Download">
              <Icon name="mdi:download" size="20" />
            </button>
            <button
              @click="
                selectedFile = file;
                showRename = true;
                newFileName = file.name;
              "
              class="text-yellow-600 hover:text-yellow-800 p-1 rounded hover:bg-base-300 transition-colors"
              title="Rename"
            >
              <Icon name="mdi:pencil" size="20" />
            </button>
            <button @click="deleteFile(file)" class="text-red-600 hover:text-red-800 p-1 rounded hover:bg-base-300 transition-colors" title="Delete">
              <Icon name="mdi:delete" size="20" />
            </button>
          </div>
          <!-- Sub-info for mobile (shows size/modified below name on small screens) -->
          <div class="col-span-8 sm:hidden pl-8 text-xs text-base-content/50 flex flex-wrap gap-2">
            <span v-if="!file.is_directory">{{ formatSize(file.size) }}</span>
            <span v-if="!file.is_directory">•</span>
            <span>{{ formatDate(file.modified) }}</span>
          </div>
        </div>
      </div>
    </main>

    <!-- Create Folder Modal -->
    <div v-if="showCreateFolder" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
      <div class="primary btn p-6 rounded-lg shadow-lg w-96">
        <h3 class="text-lg font-bold mb-4">Create New Folder</h3>
        <input type="text" v-model="newFolderName" class="w-full px-3 py-2 border rounded mb-4" placeholder="Folder name" />
        <div class="flex justify-end space-x-2">
          <button @click="showCreateFolder = false" class="px-4 py-2 text-gray-600 hover:text-gray-800">Cancel</button>
          <button @click="createFolder" class="px-4 py-2 btn btn-primary rounded">Create</button>
        </div>
      </div>
    </div>

    <!-- Rename Modal -->
    <div v-if="showRename" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
      <div class="bg-white p-6 rounded-lg shadow-lg w-96">
        <h3 class="text-lg font-bold mb-4">Rename Item</h3>
        <input type="text" v-model="newFileName" class="w-full px-3 py-2 border rounded mb-4" placeholder="New name" />
        <div class="flex justify-end space-x-2">
          <button @click="showRename = false" class="px-4 py-2 text-gray-600 hover:text-gray-800">Cancel</button>
          <button @click="renameFile" class="px-4 py-2 btn btn-primary rounded">Rename</button>
        </div>
      </div>
    </div>
    <CodeEditor v-if="showEditor" :file-path="editingFilePath" :is-open="showEditor" @close="closeEditor" @save="handleFileSaved" />
  </div>
</template>
<script setup>
useHead({
  title: "File Manager",
  link: [{ rel: "stylesheet", href: "https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.11.1/build/styles/default.min.css" }],
  script: [
    { src: "https://cdn.jsdelivr.net/gh/highlightjs/cdn-release@11.11.1/build/highlight.min.js" }
  ],
});

import dayjs from "dayjs";

// Base API URL - you can set this in your .env file
const config = useRuntimeConfig();
const baseApiUrl = config.public.apiBase || "//localhost:8000";

const curPath = ref("/");
const files = ref([]);
const selectedFile = ref(null);
const isUploading = ref(false);
const showCreateFolder = ref(false);
const newFolderName = ref("");
const showRename = ref(false);
const newFileName = ref("");
const breadcrumbs = ref([]);
const showEditor = ref(false);
const editingFilePath = ref(null);

const { $axiosApi } = useNuxtApp();

const isEditableFile = (fileName) => {
  if (!fileName) return false;
  const editableExtensions = ["txt", "html", "css", "js", "json", "md", "xml", "yaml", "yml", "php", "py", "java", "rb", "go", "rs", "swift", "kt", "sh", "sql"];
  const extension = fileName.split(".").pop().toLowerCase();
  return editableExtensions.includes(extension);
};

const openEditor = (file) => {
  if(isEditableFile(file.name)){
    editingFilePath.value = file.path;
    showEditor.value = true;
  }
};

const closeEditor = () => {
  showEditor.value = false;
  editingFilePath.value = null;
};

const handleFileSaved = async () => {
  await loadFiles(); // Refresh file list
};

watch(()=>{return curPath.value},()=>{loadFiles()});

const loadFiles = async () => {
  try {
    const response = await $axiosApi.get("/files?path="+curPath.value);
    files.value = response.data;
    updateBreadcrumbs();
  } catch (error) {
    console.error("Error loading files:", error);
  }
};

const updateBreadcrumbs = () => {
  if (!curPath.value) return;
  const parts = curPath.value.split("/").filter((p) => p);
  breadcrumbs.value = parts.map((part, index) => ({
    name: part,
    path: "/" + parts.slice(0, index + 1).join("/"),
  }));
};

const formatSize = (bytes) => {
  const sizes = ["Bytes", "KB", "MB", "GB", "TB"];
  if (bytes === 0) return "0 Byte";
  const i = parseInt(Math.floor(Math.log(bytes) / Math.log(1024)));
  return Math.round(bytes / Math.pow(1024, i), 2) + " " + sizes[i];
};

const formatDate = (timestamp) => {
  return dayjs(timestamp * 1000).format("YYYY-MM-DD HH:mm:ss");
};

const uploadFile = async (event) => {
  const file = event.target.files[0];
  if (!file) return;

  isUploading.value = true;
  const formData = new FormData();
  formData.append("file", file);
  formData.append("destination", curPath.value + "/" + file.name);

  try {
    await $axiosApi.post("/files/upload", formData, {
      headers: {
        "Content-Type": "multipart/form-data",
      },
    });
    await loadFiles();
  } catch (error) {
    console.error("Upload failed:", error);
  } finally {
    isUploading.value = false;
  }
};

const deleteFile = async (file) => {
  if (!confirm("Are you sure you want to delete this item?")) return;

  try {
    await $axiosApi.post("/files/delete", {
      path: file.path,
    });
    await loadFiles();
  } catch (error) {
    console.error("Delete failed:", error);
  }
};

const createFolder = async () => {
  if (!newFolderName.value) return;

  try {
    await $axiosApi.post("/files/mkdir", {
      path: curPath.value + "/" + newFolderName.value,
    });
    newFolderName.value = "";
    showCreateFolder.value = false;
    await loadFiles();
  } catch (error) {
    console.error("Create folder failed:", error);
  }
};

const renameFile = async () => {
  if (!newFileName.value || !selectedFile.value) return;

  try {
    await $axiosApi.post("/files/rename", {
      old_path: selectedFile.value.path,
      new_name: newFileName.value,
    });
    newFileName.value = "";
    showRename.value = false;
    selectedFile.value = null;
    await loadFiles();
  } catch (error) {
    console.error("Rename failed:", error);
  }
};

const downloadFile = async (file) => {
  try {
    const response = await $axiosApi.post("/filemanager/download", {
      path: file.path
    }, {
      responseType: 'blob'
    });
    
    // Create download link
    const url = window.URL.createObjectURL(new Blob([response.data]));
    const link = document.createElement('a');
    link.href = url;
    link.setAttribute('download', file.name);
    document.body.appendChild(link);
    link.click();
    link.remove();
    window.URL.revokeObjectURL(url);
  } catch (error) {
    console.error("Download failed:", error);
  }
};

const navigateToPath = (path) => {
  curPath.value = path;
  loadFiles();
};

// Load files on component mount
onMounted(() => {
  loadFiles();
});
</script>
