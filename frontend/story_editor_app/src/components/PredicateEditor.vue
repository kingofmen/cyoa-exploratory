<template>
  <div class="predicate-editor p-2 border border-gray-300 rounded-md bg-gray-50">
    <div class="flex items-center mb-2">
      <select v-model="predicateType" @change="onPredicateTypeChange" class="mr-2">
        <option value="compare">Compare</option>
        <option value="combine">Combine</option>
      </select>
      <button @click="$emit('delete-predicate')" class="text-red-500 hover:text-red-700 text-xs">Remove</button>
    </div>

    <!-- Compare Editor -->
    <div v-if="predicateType === 'compare'" class="compare-editor">
      <div class="grid grid-cols-3 gap-2 mb-2">
        <input type="text" v-model="localPredicate.comp.key_one" placeholder="Key 1 (e.g., var_name)" class="input-sm" @input="updatePredicate"/>
        <select v-model="localPredicate.comp.operation" class="input-sm" @change="updatePredicate">
          <option value="CMP_GT">&gt;</option>
          <option value="CMP_LT">&lt;</option>
          <option value="CMP_EQ">=</option>
          <option value="CMP_GTE">&gt;=</option>
          <option value="CMP_LTE">&lt;=</option>
          <option value="CMP_NEQ">!=</option>
          <option value="CMP_STREQ">String Equals</option>
          <option value="CMP_STRIN">String In</option>
        </select>
        <input type="text" v-model="localPredicate.comp.key_two" placeholder="Key 2 (e.g., value or var_name)" class="input-sm" @input="updatePredicate"/>
      </div>
    </div>

    <!-- Combine Editor -->
    <div v-if="predicateType === 'combine'" class="combine-editor">
      <div class="mb-2">
        <select v-model="localPredicate.comb.operation" class="input-sm" @change="updatePredicate">
          <option value="IF_ALL">All Of (AND)</option>
          <option value="IF_ANY">Any Of (OR)</option>
          <option value="IF_NONE">None Of (NOT)</option>
        </select>
      </div>
      <div v-for="(subPredicate, index) in localPredicate.comb.operands" :key="index" class="ml-4 mb-2">
        <PredicateEditor :predicate="subPredicate" @update:predicate="updateSubPredicate(index, $event)" @delete-predicate="removeSubPredicate(index)"/>
      </div>
      <button @click="addOperand" class="btn-sm bg-green-500 hover:bg-green-600 text-white mt-2">Add Operand</button>
    </div>
  </div>
</template>

<script>
export default {
  name: 'PredicateEditor',
  props: {
    predicate: {
      type: Object,
      required: true,
      default: () => ({ comp: { key_one: '', key_two: '', operation: 'CMP_EQ' } }) // Default to a compare operation
    }
  },
  data() {
    // Deep clone the predicate to avoid mutating the prop directly
    const localPredicateCopy = JSON.parse(JSON.stringify(this.predicate));

    // Ensure the predicate structure is initialized
    if (!localPredicateCopy.comp && !localPredicateCopy.comb) {
      localPredicateCopy.comp = { key_one: '', key_two: '', operation: 'CMP_EQ' };
    } else if (localPredicateCopy.comp && typeof localPredicateCopy.comp.operation === 'number') {
      // Convert numeric enum to string for select binding if necessary
      localPredicateCopy.comp.operation = this.compareOpToString(localPredicateCopy.comp.operation);
    } else if (localPredicateCopy.comb && typeof localPredicateCopy.comb.operation === 'number') {
      localPredicateCopy.comb.operation = this.combineOpToString(localPredicateCopy.comb.operation);
    }


    let type = 'compare';
    if (localPredicateCopy.comb) {
      type = 'combine';
      if (!localPredicateCopy.comb.operands) {
        localPredicateCopy.comb.operands = [];
      }
    } else if (localPredicateCopy.comp) {
      type = 'compare';
    }


    return {
      localPredicate: localPredicateCopy,
      predicateType: type,
    };
  },
  watch: {
    predicate: {
      handler(newVal) {
        this.localPredicate = JSON.parse(JSON.stringify(newVal));
        if (this.localPredicate.comp) {
          this.predicateType = 'compare';
          this.localPredicate.comp.operation = this.compareOpToString(this.localPredicate.comp.operation);
        } else if (this.localPredicate.comb) {
          this.predicateType = 'combine';
          this.localPredicate.comb.operation = this.combineOpToString(this.localPredicate.comb.operation);
          if (!this.localPredicate.comb.operands) {
            this.localPredicate.comb.operands = [];
          }
        } else {
          // Default to compare if structure is missing
          this.predicateType = 'compare';
          this.localPredicate = { comp: { key_one: '', key_two: '', operation: 'CMP_EQ' } };
        }
      },
      deep: true,
      immediate: true // Important to initialize on component creation
    }
  },
  methods: {
    compareOpMap: {
      CMP_GT: 0, CMP_LT: 1, CMP_EQ: 2, CMP_GTE: 3, CMP_LTE: 4, CMP_NEQ: 5, CMP_STREQ: 6, CMP_STRIN: 7,
      0: "CMP_GT", 1: "CMP_LT", 2: "CMP_EQ", 3: "CMP_GTE", 4: "CMP_LTE", 5: "CMP_NEQ", 6: "CMP_STREQ", 7: "CMP_STRIN",
    },
    combineOpMap: {
      IF_ALL: 0, IF_ANY: 1, IF_NONE: 2,
      0: "IF_ALL", 1: "IF_ANY", 2: "IF_NONE",
    },
    compareOpToString(opValue) {
      return (typeof opValue === 'number') ? this.compareOpMap[opValue] : opValue;
    },
    compareOpToNumber(opString) {
      return (typeof opString === 'string') ? this.compareOpMap[opString] : opString;
    },
    combineOpToString(opValue) {
      return (typeof opValue === 'number') ? this.combineOpMap[opValue] : opValue;
    },
    combineOpToNumber(opString) {
      return (typeof opString === 'string') ? this.combineOpMap[opString] : opString;
    },
    onPredicateTypeChange() {
      if (this.predicateType === 'compare') {
        this.localPredicate = { comp: { key_one: '', key_two: '', operation: 'CMP_EQ' } };
        delete this.localPredicate.comb;
      } else {
        this.localPredicate = { comb: { operands: [], operation: 'IF_ALL' } };
        delete this.localPredicate.comp;
      }
      this.updatePredicate();
    },
    addOperand() {
      if (!this.localPredicate.comb) {
        this.localPredicate.comb = { operands: [], operation: 'IF_ALL' };
      }
      if (!this.localPredicate.comb.operands) {
        this.localPredicate.comb.operands = [];
      }
      this.localPredicate.comb.operands.push({ comp: { key_one: '', key_two: '', operation: 'CMP_EQ' } });
      this.updatePredicate();
    },
    updateSubPredicate(index, updatedSubPredicate) {
      this.localPredicate.comb.operands.splice(index, 1, updatedSubPredicate);
      this.updatePredicate();
    },
    removeSubPredicate(index) {
      this.localPredicate.comb.operands.splice(index, 1);
      this.updatePredicate();
    },
    updatePredicate() {
      // Create a copy to emit, converting enums back to numbers if necessary
      const predicateToEmit = JSON.parse(JSON.stringify(this.localPredicate));
      if (predicateToEmit.comp) {
        predicateToEmit.comp.operation = this.compareOpToNumber(predicateToEmit.comp.operation);
      }
      if (predicateToEmit.comb) {
        predicateToEmit.comb.operation = this.combineOpToNumber(predicateToEmit.comb.operation);
        // Recursively ensure sub-predicates also have numeric enums
        if (predicateToEmit.comb.operands) {
            predicateToEmit.comb.operands = predicateToEmit.comb.operands.map(op => this.convertEnumsToNumbers(op));
        }
      }
      this.$emit('update:predicate', predicateToEmit);
    },
    // Helper function to recursively convert enums in a predicate tree
    convertEnumsToNumbers(predicate) {
        const converted = JSON.parse(JSON.stringify(predicate));
        if (converted.comp) {
            converted.comp.operation = this.compareOpToNumber(converted.comp.operation);
        }
        if (converted.comb) {
            converted.comb.operation = this.combineOpToNumber(converted.comb.operation);
            if (converted.comb.operands) {
                converted.comb.operands = converted.comb.operands.map(op => this.convertEnumsToNumbers(op));
            }
        }
        return converted;
    }
  },
  mounted() {
    // Ensure that on mount, if the predicate is empty or malformed, it defaults correctly.
    if (!this.localPredicate.comp && !this.localPredicate.comb) {
        this.onPredicateTypeChange(); // Initialize to default 'compare'
    }
  }
};
</script>

<style scoped>
.input-sm {
  /* Tailwind classes for smaller inputs if you're using Tailwind */
  @apply p-1 border border-gray-300 rounded-md text-sm;
}
.btn-sm {
  @apply px-2 py-1 rounded-md text-sm;
}
</style>
