<template>
  <div class="action-editor p-4 border border-gray-400 rounded-lg bg-white shadow-md">
    <h3 class="text-xl font-semibold text-gray-800 mb-4">Edit Action</h3>

    <div class="mb-4">
      <label :for="'actionTitle-' + localAction.id" class="block text-sm font-medium text-gray-700 mb-1">Title:</label>
      <input
        :id="'actionTitle-' + localAction.id"
        type="text"
        v-model="localAction.title"
        placeholder="Enter action title"
        class="input-field w-full"
      />
    </div>

    <div class="mb-4">
      <label :for="'actionDescription-' + localAction.id" class="block text-sm font-medium text-gray-700 mb-1">Description:</label>
      <textarea
        :id="'actionDescription-' + localAction.id"
        v-model="localAction.description"
        placeholder="Enter action description"
        rows="3"
        class="input-field w-full"
      ></textarea>
    </div>

    <div class="mt-6">
      <h4 class="text-lg font-medium text-gray-800 mb-2">Triggers:</h4>
      <button
        @click="createNewTrigger"
        class="mb-3 px-3 py-1.5 bg-green-500 text-white rounded hover:bg-green-600 focus:outline-none focus:ring-2 focus:ring-green-500 focus:ring-opacity-50 text-sm">
        Add New Trigger
      </button>
      <div v-if="currentTrigger" class="mt-4 p-3 border border-gray-300 rounded-md bg-gray-50">
        <TriggerActionEditor
          :triggerAction="currentTrigger"
          :availableLocations="availableLocations"
          @save-trigger-action="handleSaveTrigger"
          @cancel-edit="handleCancelTriggerEdit" />
      </div>
      <ul v-if="localAction.triggers && localAction.triggers.length" class="list-disc pl-5 mt-2">
        <li v-for="(trigger, index) in localAction.triggers" :key="trigger.id || index" class="mb-2 flex justify-between items-center">
          <span>Trigger {{ index + 1 }} <span v-if="trigger.condition && trigger.condition.comp" class="text-xs text-gray-500 ml-2">({{trigger.condition.comp.keyOne}} {{trigger.condition.comp.operation}} {{trigger.condition.comp.keyTwo}})</span></span>
          <button
            @click="editTrigger(trigger)"
            class="ml-2 px-2 py-1 bg-gray-200 text-gray-700 rounded hover:bg-gray-300 focus:outline-none focus:ring-2 focus:ring-gray-400 focus:ring-opacity-50 text-xs"
          >
            Edit
          </button>
        </li>
      </ul>
      <p v-else class="text-sm text-gray-500">No triggers defined for this action yet.</p>
    </div>

    <div class="flex justify-end mt-6">
      <button @click="saveAction" class="px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-opacity-50">
        Save Action
      </button>
      <button @click="$emit('cancel-edit')" class="ml-2 px-4 py-2 bg-gray-200 text-gray-700 rounded-md hover:bg-gray-300 focus:outline-none">
        Cancel
      </button>
    </div>
  </div>
</template>

<script>
import TriggerActionEditor from './TriggerActionEditor.vue';

export default {
  name: 'ActionEditor',
  components: {
    TriggerActionEditor,
  },
  props: {
    action: {
      type: Object,
      required: true,
      default: () => ({
        id: crypto.randomUUID(),
        title: '',
        description: '',
        triggers: [],
      })
    },
    availableLocations: { // Pass this down to TriggerActionEditor
      type: Array,
      default: () => []
    }
  },
  data() {
    // Deep clone to avoid mutating prop directly
    let initialData = JSON.parse(JSON.stringify(this.action));

    if (!initialData.id) {
        initialData.id = crypto.randomUUID();
    }
    if (!initialData.triggers) {
      initialData.triggers = [];
    }

    return {
      localAction: initialData,
      currentTrigger: null, // For editing a specific trigger within this action.
    };
  },
  watch: {
    action: {
      handler(newVal) {
        let newLocalData = JSON.parse(JSON.stringify(newVal));
        if (!newLocalData.id) {
            newLocalData.id = crypto.randomUUID();
        }
        if (!newLocalData.triggers) {
          newLocalData.triggers = [];
        }
        this.localAction = newLocalData;
        this.currentTrigger = null; // Reset current trigger when action changes
      },
      deep: true,
      immediate: true,
    }
  },
  methods: {
    initializeNewTrigger() {
      return {
        id: crypto.randomUUID(), // Temporary frontend ID.
        condition: null,
        effects: [],
        isFinal: false,
      };
    },
    createNewTrigger() {
      const newTrigger = this.initializeNewTrigger();
      if (!this.localAction.triggers) {
        this.localAction.triggers = [];
      }
      // Do not add to list yet, only set as current for editing.
      this.currentTrigger = newTrigger;
    },
    editTrigger(trigger) {
      this.currentTrigger = JSON.parse(JSON.stringify(trigger)); // Edit a copy
    },
    handleSaveTrigger(updatedTrigger) {
      const index = this.localAction.triggers.findIndex(t => t.id === updatedTrigger.id);
      if (index !== -1) {
        this.localAction.triggers.splice(index, 1, updatedTrigger);
      } else {
        // If it's a new trigger (ID not found, or was a temporary ID that got replaced)
        this.localAction.triggers.push(updatedTrigger);
      }
      this.currentTrigger = null; // Close editor
    },
    handleCancelTriggerEdit() {
      this.currentTrigger = null;
    },
    saveAction() {
      this.$emit('save-action', JSON.parse(JSON.stringify(this.localAction)));
    }
  },
  mounted() {
    if (!this.localAction.id) {
        this.localAction.id = crypto.randomUUID();
    }
    if (!this.localAction.triggers) {
      this.localAction.triggers = [];
    }
  }
};
</script>

<style scoped>
.input-field {
  @apply mt-1 block w-full px-3 py-2 bg-white border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm;
}
/* Add any other specific styles for ActionEditor if needed */
</style>
