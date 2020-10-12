<!-- Copyright (c) 2020 OSRAM; Licensed under the MIT license. -->

<template>
  <v-slider
    class="phytofy-level-slider"
    :value="limiter"
    :min="$props.min"
    :max="$props.max"
    :step="$props.step"
    :color="$props.color"
    :vertical="true"
    :key="key"
    track-color="grey--lighten"
    @input="setLevel"
    @end="redraw"
    @click="redraw"
  ></v-slider>
</template>

<style scoped>
.phytofy-level-slider >>> .v-slider {
  min-height: 100px !important;
}
</style>

<script lang="ts">
import Vue from "vue";

export default Vue.extend({
  name: "PhytofyLevelSlider",

  props: {
    value: {
      type: Number,
      required: true
    },
    min: {
      type: Number,
      default: 0
    },
    max: {
      type: Number,
      default: 100
    },
    step: {
      type: Number,
      default: 0.1
    },
    limit: {
      type: Number,
      default: 100
    },
    color: {
      type: String,
      default: "primary"
    },
  },

  data: () => ({
    level: 0,
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

  computed: {
    limiter() {
      return Math.min(this.$props.value, this.$props.limit);
    },
  },

  methods: {
    setLevel(level: number) {
      this.conditionNumber(level);
    },

    conditionNumber(value: number) {
      let conditionedNumber: number = value;
      const min: number = this.$props.min;
      if (conditionedNumber < min) {
        conditionedNumber = min;
      }
      const max: number = Math.min(this.$props.max, this.$props.limit);
      if (conditionedNumber > max) {
        conditionedNumber = max;
      }
      if (this.level !== conditionedNumber) {
        this.level = conditionedNumber;
      }
      if (
        conditionedNumber !== value ||
        conditionedNumber !== this.$props.value
      ) {
        this.$emit("input", conditionedNumber);
      }
    },

    redraw() {
      this.key++; // Enforce re-rendering
      this.$emit("change", this.level);
    },
  },
});
</script>
