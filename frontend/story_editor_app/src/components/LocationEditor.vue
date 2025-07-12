<template>
  <div class="mt-4 p-4 border border-gray-300 rounded-md bg-white">
    <h4 class="text-lg font-medium text-gray-800 mb-2">Edit Location: {{ editingLocation.title }}</h4>
    <label :for="'locationTitle-' + editingLocation.id" class="block text-sm font-medium text-gray-700 mb-1">Title:</label>
    <input
        :id="'locationTitle-' + editingLocation.id"
        type="text"
        v-model="editingLocation.title"
        placeholder="Enter location title"
        class="focus:ring-indigo-500 focus:border-indigo-500 w-full mb-3"
    />
    <label :for="'locationContent-' + editingLocation.id" class="block text-sm font-medium text-gray-700 mb-1">Description:</label>
    <textarea
        :id="'locationContent-' + editingLocation.id"
        v-model="editingLocation.content"
        placeholder="Enter location description"
        rows="4"
        class="focus:ring-indigo-500 focus:border-indigo-500 w-full mb-3"
    ></textarea>

    <!-- Possible Actions -->
    <div class="mt-4">
        <h5 class="text-md font-medium text-gray-800 mb-2">Possible Actions</h5>
        <div v-for="(possibleAction, index) in editingLocation.possibleActions" :key="index" class="p-3 border border-gray-200 rounded-md mb-3 bg-gray-50">
            <div class="flex justify-between items-center mb-2">
                <select v-model="possibleAction.actionId" class="input-field">
                    <option value="">-- Select Action --</option>
                    <option v-for="action in availableActions" :key="action.id" :value="action.id">
                        {{ action.title }}
                    </option>
                </select>
                <button @click="removePossibleAction(index)" class="text-red-500 hover:text-red-700">Remove</button>
            </div>
            <PredicateEditor :predicate="possibleAction.condition" @update:predicate="updatePredicate(index, $event)" />
        </div>
        <button @click="addPossibleAction" class="px-3 py-1 bg-green-500 text-white rounded hover:bg-green-600">Add Possible Action</button>
    </div>

    <div class="flex justify-end space-x-2 mt-4">
        <button
            @click="saveLocation"
            class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-opacity-50">
            Save Location
        </button>
        <button
            @click="cancelEdit"
            class="px-4 py-2 bg-gray-300 text-gray-700 rounded hover:bg-gray-400 focus:outline-none focus:ring-2 focus:ring-gray-200 focus:ring-opacity-50">
            Cancel
        </button>
    </div>
  </div>
</template>

<script>
import PredicateEditor from './PredicateEditor.vue';

export default {
  components: {
    PredicateEditor,
  },
  props: {
    location: {
      type: Object,
      required: true,
    },
    availableActions: {
      type: Array,
      default: () => [],
    },
  },
  data() {
    return {
      // Create a local copy to prevent modifying the prop directly
      editingLocation: this.initEditingLocation(),
    };
  },
  watch: {
    location: {
      immediate: true, // Trigger the watcher upon component creation
      handler(newLocation) {
        // Update the local copy when the prop changes
        this.editingLocation = this.initEditingLocation();
      }
    }
  },
  methods: {
    initEditingLocation() {
        const loc = this.location ? JSON.parse(JSON.stringify(this.location)) : {};
        if (!loc.possibleActions) {
            loc.possibleActions = [];
        }
        return loc;
    },
    addPossibleAction() {
        this.editingLocation.possibleActions.push({
            actionId: '',
            condition: { comp: { keyOne: '', keyTwo: '', operation: 'CMP_EQ' } },
        });
    },
    removePossibleAction(index) {
        this.editingLocation.possibleActions.splice(index, 1);
    },
    updatePredicate(index, predicate) {
        this.editingLocation.possibleActions[index].condition = predicate;
    },
    saveLocation() {
      if (this.editingLocation) {
        this.$emit('save-location', this.editingLocation);
      }
    },
    cancelEdit() {
      this.$emit('cancel-edit');
    },
  },
};
</script>

<style scoped>
/* Scoped styles for LocationEditor can go here if needed */
</style>
