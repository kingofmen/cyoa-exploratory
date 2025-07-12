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
        <label class="block text-sm font-medium text-gray-700 mb-1">New Location:</label>
        <select
          v-model="localEffect.newLocationId"
          @change="updateEffect"
          class="input-field"
        >
          <option value="">--- Select Location ---</option>
          <option v-for="location in availableLocations" :key="location.id" :value="location.id">
            {{ location.title }}
          </option>
        </select>
      </div>
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Tweak Value (Variable Name):</label>
        <input
          type="text"
          v-model="localEffect.tweakValue"
          placeholder="Variable to change (e.g., score)"
          @input="updateEffect"
          class="input-field"
        />
      </div>
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Tweak Amount:</label>
        <input
          type="number"
          v-model.number="localEffect.tweakAmount"
          placeholder="Amount (e.g., 10 or -5)"
          @input="updateEffect"
          class="input-field"
        />
      </div>
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">New Run State:</label>
        <select v-model="localEffect.newState" @change="updateEffect" class="input-field">
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
        newLocationId: '',
        tweakValue: '',
        tweakAmount: 0,
        newState: 'RS_UNKNOWN', // Default to string representation
      })
    },
    availableLocations: {
      type: Array,
      default: () => []
    }
  },
  data() {
    return {
      runStateMap: {
        RS_UNKNOWN: 0, RS_ACTIVE: 1, RS_HIATUS: 2, RS_COMPLETE: 3,
        0: "RS_UNKNOWN", 1: "RS_ACTIVE", 2: "RS_HIATUS", 3: "RS_COMPLETE",
      },
    };
  },
  created() {
    // Deep clone and ensure correct types.
    const effectCopy = JSON.parse(JSON.stringify(this.effect));
    effectCopy.tweakAmount = Number(effectCopy.tweakAmount) || 0;
    effectCopy.newState = this.runStateToString(effectCopy.newState); // Ensure it's a string for select
    this.localEffect = effectCopy
  },
  watch: {
    effect: {
      handler(newVal) {
        const newEffectCopy = JSON.parse(JSON.stringify(newVal));
        newEffectCopy.tweakAmount = Number(newEffectCopy.tweakAmount) || 0;
        newEffectCopy.newState = this.runStateToString(newEffectCopy.newState);
        this.localEffect = newEffectCopy;
      },
      deep: true,
      immediate: true
    }
  },
  methods: {
    runStateToString(value) {
      return (typeof value === 'number') ? this.runStateMap[value] : value;
    },
    runStateToNumber(value) {
      return (typeof value === 'string') ? this.runStateMap[value] : value;
    },
    updateEffect() {
      const effectToEmit = JSON.parse(JSON.stringify(this.localEffect));
      // Convert enum back to number before emitting
      effectToEmit.newState = this.runStateToNumber(effectToEmit.newState);
      // Ensure tweakAmount is a number
      effectToEmit.tweakAmount = Number(effectToEmit.tweakAmount) || 0;
      this.$emit('update:effect', effectToEmit);
    }
  },
  mounted() {
    // Ensure newState is initialized if not present in prop
    if (!this.localEffect.newState) {
        this.localEffect.newState = 'RS_UNKNOWN';
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
