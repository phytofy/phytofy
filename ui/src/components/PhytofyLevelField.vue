<!-- Copyright (c) 2020 OSRAM; Licensed under the MIT license. -->

<template>
  <v-text-field
    type="number"
    :value="level"
    :min="$props.min"
    :max="$props.max"
    :step="$props.step"
    :hint="$props.hint"
    persistent-hint
    :maxlength="5"
    :key="key"
    suffix="%"
    dense
    @input="setLevel"
  ></v-text-field>
</template>

<script lang="ts">
import Vue from "vue";

interface TrimDecimalsResult {
  valueString: string;
  valueNumber: number;
}

const trimDecimals = (valueNumber: number): TrimDecimalsResult => {
  let valueString: string;
  if (valueNumber % 1 === 0) {
    valueString = valueNumber.toFixed(0);
  } else {
    valueString = valueNumber.toFixed(1);
    valueNumber = Number(valueString);
  }
  return { valueNumber, valueString };
};

export default Vue.extend({
  name: "PhytofyLevelField",

  props: {
    value: {
      type: Number,
      required: true,
    },
    min: {
      type: Number,
      default: 0,
    },
    max: {
      type: Number,
      default: 100,
    },
    step: {
      type: Number,
      default: 0.1,
    },
    limit: {
      type: Number,
      default: 100,
    },
    hint: {
      type: String,
    },
  },

  data: () => ({
    level: "",
    key: 0,
  }),

  mounted() {
    this.conditionNumber(this.$props.value);
  },

  watch: {
    value(newValue) {
      this.conditionNumber(newValue);
    },
  },

  methods: {
    setLevel(level: string) {
      this.conditionString(level);
    },

    conditionNumber(value: number) {
      let correction = false;
      let conditionedNumber: number = value;
      const min: number = this.$props.min;
      if (conditionedNumber < min) {
        conditionedNumber = min;
        correction = true;
      }
      const max: number = Math.min(this.$props.max, this.$props.limit);
      if (conditionedNumber > max) {
        conditionedNumber = max;
        correction = true;
      }
      const result: TrimDecimalsResult = trimDecimals(conditionedNumber);
      conditionedNumber = result.valueNumber;
      const conditionedString: string = result.valueString;
      if (this.level !== conditionedString) {
        this.level = conditionedString;
      }
      if (correction) {
        this.key++; // Enforce re-rendering
      }
      if (
        conditionedNumber !== value ||
        conditionedNumber !== this.$props.value
      ) {
        this.$emit("input", conditionedNumber);
        this.$emit("change", conditionedNumber);
      }
    },

    conditionString(value: string) {
      let conditionedString: string =
        value === "" ? this.level : value.replace(/[^0-9.]/g, "");
      let conditionedNumber = Number(conditionedString);
      const min: number = this.$props.min;
      if (conditionedNumber < min) {
        conditionedNumber = min;
        conditionedString = trimDecimals(conditionedNumber).valueString;
      }
      const max: number = Math.min(this.$props.max, this.$props.limit);
      if (conditionedNumber > max) {
        conditionedNumber = max;
        conditionedString = trimDecimals(conditionedNumber).valueString;
      }
      const result: TrimDecimalsResult = trimDecimals(conditionedNumber);
      if (conditionedNumber !== result.valueNumber) {
        conditionedNumber = result.valueNumber;
        conditionedString = result.valueString;
      }
      if (this.level !== conditionedString) {
        this.level = conditionedString;
      }
      if (conditionedString !== value) {
        this.key++; // Enforce re-rendering
      }
      if (conditionedNumber !== this.$props.value) {
        this.$emit("input", conditionedNumber);
        this.$emit("change", conditionedNumber);
      }
    },
  },
});
</script>
