<template>
  <div class="trigger-action-editor p-4 border border-gray-400 rounded-lg bg-white shadow-md">
    <h3 class="text-lg font-semibold text-gray-800 mb-3">Edit Event/Trigger</h3>

    <div class="mb-4">
      <label class="block text-sm font-medium text-gray-700 mb-1">Condition (Predicate):</label>
      <PredicateEditor v-if="localTriggerAction.condition" :predicate="localTriggerAction.condition" @update:predicate="updateCondition" />
      <button v-else @click="initializeCondition" class="btn-sm bg-blue-500 hover:bg-blue-600 text-white">Add Condition</button>
    </div>

    <div class="mb-4">
      <h4 class="text-md font-medium text-gray-700 mb-2">Effects:</h4>
      <div v-for="(effect, index) in localTriggerAction.effects" :key="index" class="mb-3">
        <EffectEditor :effect="effect" @update:effect="updateEffect(index, $event)" @delete-effect="removeEffect(index)" />
      </div>
      <button @click="addEffect" class="btn-sm bg-green-500 hover:bg-green-600 text-white">Add Effect</button>
      <p v-if="!localTriggerAction.effects || localTriggerAction.effects.length === 0" class="text-sm text-gray-500 mt-1">No effects defined. Click "Add Effect".</p>
    </div>

    <div class="mb-4">
      <label class="flex items-center">
        <input type="checkbox" v-model="localTriggerAction.is_final" @change="updateTriggerAction" class="form-checkbox h-5 w-5 text-indigo-600" />
        <span class="ml-2 text-sm text-gray-700">Is Final (stops further trigger evaluation in this scope)</span>
      </label>
    </div>

    <div class="flex justify-end mt-4">
      <button @click="saveTriggerAction" class="px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-opacity-50">
        Save Event/Trigger
      </button>
       <button @click="$emit('cancel-edit')" class="ml-2 px-4 py-2 bg-gray-200 text-gray-700 rounded-md hover:bg-gray-300 focus:outline-none">
        Cancel
      </button>
    </div>
  </div>
</template>

<script>
import PredicateEditor from './PredicateEditor.vue';
import EffectEditor from './EffectEditor.vue';

export default {
  name: 'TriggerActionEditor',
  components: {
    PredicateEditor,
    EffectEditor,
  },
  props: {
    triggerAction: {
      type: Object,
      required: true,
      default: () => ({
        condition: null, // Will be initialized by button click
        effects: [],
        is_final: false,
      })
    }
  },
  data() {
    // Deep clone to avoid mutating prop directly
    let initialData = JSON.parse(JSON.stringify(this.triggerAction));

    // Ensure effects is an array
    if (!initialData.effects) {
      initialData.effects = [];
    }
    // condition can be null initially, user will add it.

    return {
      localTriggerAction: initialData,
    };
  },
  watch: {
    triggerAction: {
      handler(newVal) {
        let newLocalData = JSON.parse(JSON.stringify(newVal));
        if (!newLocalData.effects) {
          newLocalData.effects = [];
        }
        // No specific transformations needed for condition here, PredicateEditor handles its defaults.
        this.localTriggerAction = newLocalData;
      },
      deep: true,
      immediate: true,
    }
  },
  methods: {
    initializeCondition() {
      // Default to a 'compare' predicate
      this.localTriggerAction.condition = { comp: { key_one: '', key_two: '', operation: 'CMP_EQ' } };
      this.updateTriggerAction();
    },
    updateCondition(newCondition) {
      this.localTriggerAction.condition = newCondition;
      this.updateTriggerAction();
    },
    addEffect() {
      if (!this.localTriggerAction.effects) {
        this.localTriggerAction.effects = [];
      }
      this.localTriggerAction.effects.push({
        description: '',
        new_location_id: '',
        tweak_value: '',
        tweak_amount: 0,
        new_state: 'RS_UNKNOWN', // Default string for enum
      });
      this.updateTriggerAction();
    },
    updateEffect(index, updatedEffect) {
      this.localTriggerAction.effects.splice(index, 1, updatedEffect);
      this.updateTriggerAction();
    },
    removeEffect(index) {
      this.localTriggerAction.effects.splice(index, 1);
      this.updateTriggerAction();
    },
    updateTriggerAction() {
      // This method can be used if we need to do something on every change.
      // For now, it's implicitly handled by v-model and direct assignments.
      // console.log("TriggerAction updated internally", this.localTriggerAction);
    },
    saveTriggerAction() {
      this.$emit('save-trigger-action', JSON.parse(JSON.stringify(this.localTriggerAction)));
    }
  },
  mounted() {
    // Ensure basic structure if prop is empty or malformed
    if (!this.localTriggerAction.effects) {
        this.localTriggerAction.effects = [];
    }
    // Condition is handled by the 'Add Condition' button if null
  }
};
</script>

<style scoped>
.input-field {
  @apply mt-1 block w-full px-3 py-2 bg-white border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm;
}
.btn-sm {
  @apply px-2 py-1 rounded-md text-sm;
}
</style>
