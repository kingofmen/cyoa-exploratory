<template>
  <div class="effect-editor p-3 border border-gray-300 rounded-md bg-gray-50 shadow-sm">
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Description:</label>
        <input
          type="text"
          v-model="localEffect.description"
          placeholder="Effect description (optional)"
          @input="updateEffect"
          class="input-field"
        />
      </div>
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">New Location ID:</label>
        <input
          type="text"
          v-model="localEffect.new_location_id"
          placeholder="Enter location ID (optional)"
          @input="updateEffect"
          class="input-field"
        />
        <!-- TODO: Consider a dropdown populated with actual location IDs from the story -->
      </div>
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Tweak Value (Variable Name):</label>
        <input
          type="text"
          v-model="localEffect.tweak_value"
          placeholder="Variable to change (e.g., score)"
          @input="updateEffect"
          class="input-field"
        />
      </div>
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Tweak Amount:</label>
        <input
          type="number"
          v-model.number="localEffect.tweak_amount"
          placeholder="Amount (e.g., 10 or -5)"
          @input="updateEffect"
          class="input-field"
        />
      </div>
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">New Run State:</label>
        <select v-model="localEffect.new_state" @change="updateEffect" class="input-field">
          <option value="RS_UNKNOWN">Unknown</option>
          <option value="RS_ACTIVE">Active</option>
          <option value="RS_HIATUS">Hiatus</option>
          <option value="RS_COMPLETE">Complete</option>
        </select>
      </div>
    </div>
    <button @click="$emit('delete-effect')" class="mt-3 text-red-500 hover:text-red-700 text-xs">Remove Effect</button>
  </div>
</template>

<script>
export default {
  name: 'EffectEditor',
  props: {
    effect: {
      type: Object,
      required: true,
      default: () => ({
        description: '',
        new_location_id: '',
        tweak_value: '',
        tweak_amount: 0,
        new_state: 'RS_UNKNOWN', // Default to string representation
      })
    }
  },
  data() {
    // Deep clone and ensure correct types
    const effectCopy = JSON.parse(JSON.stringify(this.effect));
    effectCopy.tweak_amount = Number(effectCopy.tweak_amount) || 0;
    effectCopy.new_state = this.runStateToString(effectCopy.new_state); // Ensure it's a string for select
    return {
      localEffect: effectCopy,
    };
  },
  watch: {
    effect: {
      handler(newVal) {
        const newEffectCopy = JSON.parse(JSON.stringify(newVal));
        newEffectCopy.tweak_amount = Number(newEffectCopy.tweak_amount) || 0;
        newEffectCopy.new_state = this.runStateToString(newEffectCopy.new_state);
        this.localEffect = newEffectCopy;
      },
      deep: true,
      immediate: true
    }
  },
  methods: {
    runStateMap: {
      RS_UNKNOWN: 0, RS_ACTIVE: 1, RS_HIATUS: 2, RS_COMPLETE: 3,
      0: "RS_UNKNOWN", 1: "RS_ACTIVE", 2: "RS_HIATUS", 3: "RS_COMPLETE",
    },
    runStateToString(value) {
      return (typeof value === 'number') ? this.runStateMap[value] : value;
    },
    runStateToNumber(value) {
      return (typeof value === 'string') ? this.runStateMap[value] : value;
    },
    updateEffect() {
      const effectToEmit = JSON.parse(JSON.stringify(this.localEffect));
      // Convert enum back to number before emitting
      effectToEmit.new_state = this.runStateToNumber(effectToEmit.new_state);
      // Ensure tweak_amount is a number
      effectToEmit.tweak_amount = Number(effectToEmit.tweak_amount) || 0;
      this.$emit('update:effect', effectToEmit);
    }
  },
  mounted() {
    // Ensure new_state is initialized if not present in prop
    if (!this.localEffect.new_state) {
        this.localEffect.new_state = 'RS_UNKNOWN';
    }
    this.updateEffect(); // Emit initial state
  }
};
</script>

<style scoped>
.input-field {
  @apply mt-1 block w-full px-3 py-2 bg-white border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm;
}
</style>
