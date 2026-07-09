<template>
  <div class="min-h-screen">
    <main class="container mx-auto px-4 py-6">
      <!-- MySQL Status and Controls -->
      <div class="rounded-lg shadow mb-6 p-4">
        <h2 class="text-xl font-bold mb-4">Database Status</h2>
        <div v-if="isInstalled">
          <div class="mb-2">
            <span class="font-semibold">Type:</span> {{ dbType }}
          </div>
          <div class="mb-4">
            <span class="font-semibold">Version:</span> {{ isInstalled }}
          </div>
          <div class="flex flex-wrap gap-3 mb-4 space-x-0">
            <button @click="controlMySQL('activate')"
                    :disabled="isLoading"
                    class="btn btn-success px-4 py-2 rounded">
              Start
            </button>
            <button @click="controlMySQL('deactivate')"
                    :disabled="isLoading"
                    class="btn btn-error px-4 py-2 rounded">
              Stop
            </button>
            <button @click="controlMySQL('restart')"
                    :disabled="isLoading"
                    class="btn btn-warning px-4 py-2 rounded">
              Restart
            </button>
            <button @click="showCredentialsModal = true"
                    class="btn btn-primary px-4 py-2 rounded">
              Update Credentials
            </button>
          </div>
        </div>
        <div v-else class="text-red-600 mb-4">
          No database server is installed
          <button @click="showInstallModal = true"
                  class="ml-4 btn btn-success px-4 py-2 rounded">
            Install Database
          </button>
        </div>
      </div>

      <!-- Database Management -->
      <div class="rounded-lg shadow mb-6">
        <div class="p-4 border-b">
          <div class="flex justify-between items-center">
            <h2 class="text-xl font-bold">Databases</h2>
            <button @click="showCreateDatabase = true"
                    class="btn btn-primary px-4 py-2 rounded inline-flex items-center">
              <Icon name="mdi:plus" class="mr-2" />
              <span>New Database</span>
            </button>
          </div>
        </div>

        <!-- Database List -->
        <div class="hidden sm:grid grid-cols-12 gap-4 p-4 font-semibold border-b">
          <div class="col-span-6">Name</div>
          <div class="col-span-3 text-right">Actions</div>
          <div class="col-span-3 text-right">Backups</div>
        </div>
        
        <div v-for="db in databases" 
             :key="db.name"
             class="grid grid-cols-12 gap-3 sm:gap-4 p-4 border-b hover:bg-base-200 items-center">
          <div class="col-span-12 sm:col-span-6 flex items-center min-w-0">
            <Icon name="mdi:database" class="mr-3 text-blue-400 flex-shrink-0" />
            <span class="truncate font-medium text-base-content">{{ db.name }}</span>
          </div>
          <div class="col-span-6 sm:col-span-3 flex sm:justify-end">
            <button @click="deleteDatabase(db)"
                    class="text-red-500 hover:text-red-700 flex items-center gap-1 text-sm py-1 px-2 rounded hover:bg-base-300 transition-colors">
              <Icon name="mdi:delete" />
              <span class="sm:hidden">Delete</span>
            </button>
          </div>
          <div class="col-span-6 sm:col-span-3 flex justify-end">
            <button @click="navigateToBackups(db)"
                    class="text-blue-500 hover:text-blue-700 flex items-center gap-1 text-sm py-1 px-2 rounded hover:bg-base-300 transition-colors">
              <Icon name="mdi:backup-restore" />
              <span>Manage Backups</span>
            </button>
          </div>
        </div>
      </div>

      <!-- User Management -->
      <div class="rounded-lg shadow">
        <div class="p-4 border-b">
          <div class="flex justify-between items-center">
            <h2 class="text-xl font-bold">Users</h2>
            <button @click="showCreateUser = true"
                    class="btn btn-primary px-4 py-2 rounded inline-flex items-center">
              <Icon name="mdi:account-plus" class="mr-2" />
              <span>New User</span>
            </button>
          </div>
        </div>

        <!-- User List -->
        <div class="hidden sm:grid grid-cols-12 gap-4 p-4 font-semibold border-b">
          <div class="col-span-3">Username</div>
          <div class="col-span-2">Host</div>
          <div class="col-span-4">Database Access</div>
          <div class="col-span-3 text-right">Actions</div>
        </div>
        
        <div v-for="(user,ui) in users" 
             :key="ui"
             class="grid grid-cols-12 gap-3 sm:gap-4 p-4 border-b hover:bg-base-200 items-start">
          <div class="col-span-12 sm:col-span-3 flex items-center min-w-0">
            <Icon name="mdi:account" class="mr-3 text-gray-400 flex-shrink-0" />
            <span class="truncate font-medium text-base-content">{{ user.username }}</span>
          </div>
          <div class="col-span-12 sm:col-span-2 text-sm sm:text-base text-base-content/85">
            <span class="sm:hidden font-semibold">Host: </span>{{ user.host || 'localhost' }}
          </div>
          <div class="col-span-12 sm:col-span-4 text-sm">
            <span class="sm:hidden font-semibold block mb-1">Database Access: </span>
            <div v-for="grant in user.grants" :key="grant.database" class="bg-base-300/30 p-1 px-2 rounded mb-1 last:mb-0">
              <span class="font-medium text-base-content">{{ grant.database }}:</span>
              <span class="text-base-content/60 ml-1 text-xs">{{ grant.privileges.join(', ') }}</span>
            </div>
          </div>
          <div class="col-span-12 sm:col-span-3 flex sm:justify-end gap-3 border-t border-base-300/30 pt-3 sm:pt-0 sm:border-t-0 flex-wrap">
            <button @click="showPrivileges(user)"
                    class="text-blue-500 hover:text-blue-700 flex items-center gap-1 text-sm py-1 px-2 rounded hover:bg-base-300 transition-colors">
              <Icon name="mdi:shield-account" />
              <span>Manage Access</span>
            </button>
            <button @click="deleteUser(user)"
                    class="text-red-500 hover:text-red-700 flex items-center gap-1 text-sm py-1 px-2 rounded hover:bg-base-300 transition-colors">
              <Icon name="mdi:delete" />
              <span class="sm:hidden">Delete</span>
            </button>
          </div>
        </div>
      </div>
    </main>

    <!-- Create Database Modal -->
    <div v-if="showCreateDatabase" 
         class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
      <div class="p-6 rounded-lg shadow-lg w-96">
        <h3 class="text-lg font-bold mb-4">Create New Database</h3>
        <NativeTextField
          v-model="newDatabaseName"
          placeholder="Database name"
          class="mb-4"
        />
        <div class="flex justify-end space-x-2">
          <button @click="showCreateDatabase = false"
                  class="px-4 py-2 text-gray-600 hover:text-gray-800">
            Cancel
          </button>
          <button @click="createDatabase"
                  class="px-4 py-2 btn btn-primary rounded">
            Create
          </button>
        </div>
      </div>
    </div>

    <!-- Create User Modal -->
    <div v-if="showCreateUser" 
         class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
      <div class="p-6 rounded-lg shadow-lg w-96">
        <h3 class="text-lg font-bold mb-4">Create New User</h3>
        <NativeTextField
          v-model="newUser.username"
          placeholder="Username"
          class="mb-4"
        />
        <NativeTextField
          v-model="newUser.password"
          type="password"
          placeholder="Password"
          class="mb-4"
        />
        <div class="flex justify-end space-x-2">
          <button @click="showCreateUser = false"
                  class="px-4 py-2 text-gray-600 hover:text-gray-800">
            Cancel
          </button>
          <button @click="createUser"
                  class="px-4 py-2 btn btn-primary rounded">
            Create
          </button>
        </div>
      </div>
    </div>

    <!-- Privileges Modal -->
    <div v-if="showPrivilegesModal" 
         class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
      <div class="p-6 rounded-lg shadow-lg w-2/3">
        <div class="flex justify-between items-center mb-4">
          <h3 class="text-lg font-bold">Manage Privileges for {{ selectedUser?.username }}</h3>
          <button @click="closePrivilegesModal" class="text-gray-500 hover:text-gray-700">
            <Icon name="mdi:close" />
          </button>
        </div>
        
        <div class="mb-4">
          <label class="block text-sm font-medium text-gray-700 mb-2">Database</label>
          <select v-model="selectedDatabase"
                  class="w-full px-3 py-2 border rounded">
            <option v-for="db in databases" 
                    :key="db.name" 
                    :value="db.name">
              {{ db.name }}
            </option>
          </select>
        </div>

        <div class="mb-4">
          <label class="block text-sm font-medium text-gray-700 mb-2">Privileges</label>
          <div class="grid grid-cols-3 gap-4">
            <label v-for="priv in availablePrivileges" 
                   :key="priv"
                   class="flex items-center">
              <input type="checkbox"
                     v-model="selectedPrivileges"
                     :value="priv"
                     class="mr-2">
              {{ priv }}
            </label>
          </div>
        </div>

        <div class="flex justify-end space-x-2">
          <button @click="grantPrivileges"
                  class="px-4 py-2 btn btn-success rounded">
            Grant Privileges
          </button>
          <button @click="revokePrivileges"
                  class="px-4 py-2 btn btn-error rounded">
            Revoke Privileges
          </button>
        </div>
      </div>
    </div>

    <!-- Install Database Modal -->
    <div v-if="showInstallModal" 
         class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
      <div class="p-6 rounded-lg shadow-lg w-96">
        <h3 class="text-lg font-bold mb-4">Install Database Server</h3>
        <div class="mb-4">
          <label class="block text-sm font-medium text-gray-700 mb-2">Database Type</label>
          <select v-model="installCredentials.type"
                  class="w-full px-3 py-2 border rounded">
            <option value="mysql">MySQL</option>
            <option value="mariadb">MariaDB</option>
          </select>
        </div>
        <div class="mb-4">
          <NativeTextField
            v-model="installCredentials.username"
            label="Root Username"
            placeholder="root"
          />
        </div>
        <div class="mb-4">
          <NativeTextField
            v-model="installCredentials.password"
            type="password"
            label="Root Password"
            placeholder="Enter root password"
          />
        </div>
        <div class="flex justify-end space-x-2">
          <button @click="showInstallModal = false"
                  class="px-4 py-2 text-gray-600 hover:text-gray-800">
            Cancel
          </button>
          <button @click="installMySQL"
                  :disabled="isInstalling"
                  class="px-4 py-2 btn btn-success rounded">
            {{ isInstalling ? 'Installing...' : 'Install' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Update Credentials Modal -->
    <div v-if="showCredentialsModal" 
         class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
      <div class="p-6 rounded-lg shadow-lg w-96">
        <h3 class="text-lg font-bold mb-4">Update MySQL Credentials</h3>
        <div class="mb-4">
          <NativeTextField
            v-model="credentials.username"
            label="Username"
            placeholder="Enter username"
          />
        </div>
        <div class="mb-4">
          <NativeTextField
            v-model="credentials.password"
            type="password"
            label="Password"
            placeholder="Enter password"
          />
        </div>
        <div class="flex justify-end space-x-2">
          <button @click="showCredentialsModal = false"
                  class="px-4 py-2 text-gray-600 hover:text-gray-800">
            Cancel
          </button>
          <button @click="updateCredentials"
                  :disabled="isUpdating"
                  class="px-4 py-2 btn btn-primary rounded">
            {{ isUpdating ? 'Updating...' : 'Update' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
const { $axiosApi } = useNuxtApp()
const router = useRouter()

// State variables
const isInstalled = ref(false)
const isLoading = ref(false)
const databases = ref([])
const users = ref([])
const showCreateDatabase = ref(false)
const showCreateUser = ref(false)
const showPrivilegesModal = ref(false)
const newDatabaseName = ref('')
const newUser = ref({ username: '', password: '' })
const selectedUser = ref(null)
const selectedDatabase = ref('')
const selectedPrivileges = ref([])
const availablePrivileges = [
  'ALL',
  'SELECT',
  'INSERT',
  'UPDATE',
  'DELETE',
  'CREATE',
  'DROP',
  'REFERENCES',
  'INDEX',
  'ALTER',
  'CREATE TEMPORARY TABLES',
  'LOCK TABLES',
  'EXECUTE',
  'CREATE VIEW',
  'SHOW VIEW',
  'CREATE ROUTINE',
  'ALTER ROUTINE',
  'EVENT',
  'TRIGGER'
]
const showInstallModal = ref(false)
const showCredentialsModal = ref(false)
const isInstalling = ref(false)
const isUpdating = ref(false)
const installCredentials = ref({ 
  type: 'mysql',
  username: 'root', 
  password: '' 
})
const credentials = ref({ username: '', password: '' })
const dbType = ref('')

const checkMySQLInstalled = () => {
  $axiosApi.get('/databases/sql/installed')
    .then(response => {
      const data = response.data
      isInstalled.value = data.installed
      dbType.value = data.type || ''
    })
    .catch(error => {
      isInstalled.value = false
      dbType.value = ''
      console.error('Error checking database installation:', error)
    })
}

const loadDatabases = () => {
  $axiosApi.get('/databases/sql/databases')
    .then(response => {
      const data = response.data
      if (data.success) {
        databases.value = data.databases
      }
    })
    .catch(error => {
      console.error('Error loading databases:', error)
    })
}

const loadUsers = () => {
  $axiosApi.get('/databases/sql/users')
    .then(response => {
      users.value = response.data.users
    })
    .catch(error => {
      console.error('Error loading users:', error)
    })
}

// MySQL control functions
const controlMySQL = (action) => {
  isLoading.value = true
  const dbKey = dbType.value || 'mysql'
  $axiosApi.post(`/databases/${dbKey}/${action}`)
    .then(() => {
      // Success - no additional action needed
    })
    .catch(error => {
      console.error(`Error during MySQL ${action}:`, error)
    })
    .finally(() => {
      isLoading.value = false
    })
}

// Database management
const createDatabase = () => {
  if (!newDatabaseName.value) return

  $axiosApi.post('/databases/sql/database/create', { name: newDatabaseName.value })
    .then(() => {
      showCreateDatabase.value = false
      newDatabaseName.value = ''
      loadDatabases()
    })
    .catch(error => {
      console.error('Error creating database:', error)
    })
}

const deleteDatabase = (db) => {
  if (!confirm(`Are you sure you want to delete database ${db.name}?`)) return

  $axiosApi.post('/databases/sql/database/delete', { database: db.name })
    .then(() => {
      loadDatabases()
    })
    .catch(error => {
      console.error('Error deleting database:', error)
    })
}

// User management
const createUser = () => {
  if (!newUser.value.username || !newUser.value.password) return

  $axiosApi.post('/databases/sql/user/create', newUser.value)
    .then(() => {
      showCreateUser.value = false
      newUser.value = { username: '', password: '' }
      loadUsers()
    })
    .catch(error => {
      console.error('Error creating user:', error)
    })
}

const deleteUser = (user) => {
  if (!confirm(`Are you sure you want to delete user ${user.username}?`)) return

  $axiosApi.post('/databases/sql/user/delete', { user: user.username })
    .then(() => {
      loadUsers()
    })
    .catch(error => {
      console.error('Error deleting user:', error)
    })
}

// Privileges management
const showPrivileges = (user) => {
  selectedUser.value = user
  showPrivilegesModal.value = true
}

const closePrivilegesModal = () => {
  showPrivilegesModal.value = false
  selectedUser.value = null
  selectedDatabase.value = ''
  selectedPrivileges.value = []
}

const grantPrivileges = () => {
  if (!selectedUser.value || !selectedDatabase.value || !selectedPrivileges.value.length) return

  $axiosApi.post('/databases/sql/privileges/grant', {
    username: selectedUser.value.username,
    database: selectedDatabase.value,
    privileges: selectedPrivileges.value
  })
    .then(() => {
      closePrivilegesModal()
    })
    .catch(error => {
      console.error('Error granting privileges:', error)
    })
}

const revokePrivileges = () => {
  if (!selectedUser.value || !selectedDatabase.value || !selectedPrivileges.value.length) return

  $axiosApi.post('/databases/sql/privileges/revoke', {
    username: selectedUser.value.username,
    database: selectedDatabase.value,
    privileges: selectedPrivileges.value
  })
    .then(() => {
      closePrivilegesModal()
    })
    .catch(error => {
      console.error('Error revoking privileges:', error)
    })
}

const navigateToBackups = (db) => {
  router.push(`/sqlbackups?database=${db.name}`)
}

const installMySQL = () => {
  if (!installCredentials.value.username) {
    alert('Please provide username and password')
    return
  }

  isInstalling.value = true
  $axiosApi.post(`/databases/${installCredentials.value.type}/install`, installCredentials.value)
    .then(() => {
      showInstallModal.value = false
      checkMySQLInstalled()
    })
    .catch(error => {
      console.error('Error installing MySQL:', error)
      alert(error.response?.data?.error || 'Failed to install MySQL')
    })
    .finally(() => {
      isInstalling.value = false
    })
}

const updateCredentials = () => {
  if (!credentials.value.username) {
    alert('Please provide username and password')
    return
  }

  isUpdating.value = true
  $axiosApi.post('/databases/sql/credentials/update', credentials.value)
    .then(() => {
      showCredentialsModal.value = false
      alert('Credentials updated successfully')
    })
    .catch(error => {
      console.error('Error updating credentials:', error)
      alert(error.response?.data?.error || 'Failed to update credentials')
    })
    .finally(() => {
      isUpdating.value = false
    })
}

const loadCredentials = () => {
  $axiosApi.get('/databases/sql/credentials')
    .then(response => {
      const data = response.data
      credentials.value = {
        username: data.username,
        password: data.password
      }
    })
    .catch(error => {
      console.error('Error loading credentials:', error)
    })
}

// Load data on component mount
onMounted(() => {
  checkMySQLInstalled()
  loadDatabases()
  loadUsers()
  loadCredentials()
})
</script> 