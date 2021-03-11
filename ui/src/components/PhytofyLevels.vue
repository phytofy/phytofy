<!-- Copyright (c) 2020 OSRAM; Licensed under the MIT license. -->

<template>
  <v-row class="no-gutters">
    <template v-for="(channel, i) in channels">
      <v-col :key="i" class="ma-0 pa-0" cols="12" sm="6" md="4" lg="2">
        <v-row class="no-gutters">
          <v-col class="ma-0 pl-0 pr-2 py-0" cols="3" lg="4">
            <PhytofyLevelField
              :value="levelsUncommitted[i]"
              :limit="levelLimits[i]"
              :hint="channel.hint"
              @input="
                (newValue) =>
                  setLevel(newValue, levelsUncommitted, levelLimits, i, true)
              "
            />
          </v-col>
          <v-col :key="i" class="ma-0 pa-0" cols="9" lg="8">
            <PhytofyLevelSlider
              :value="levelsUncommitted[i]"
              :limit="levelLimits[i]"
              :color="channel.color"
              @input="
                (newValue) =>
                  setLevel(newValue, levelsUncommitted, levelLimits, i, true)
              "
            />
          </v-col>
        </v-row>
      </v-col>
    </template>
    <v-col class="ma-0 pa-0" cols="12">
      <v-row class="no-gutters">
        <v-col class="ma-0 pl-0 pr-2 py-0" cols="3" sm="2" md="1">
          <PhytofyLevelField
            :value="totalUncommitted"
            :max="300"
            :limit="totalLimit"
            hint="Total"
            @input="
              (newValue) =>
                setTotalAndLevels(
                  totalCommitted,
                  newValue,
                  totalLimit,
                  value,
                  true
                )
            "
          />
        </v-col>
        <v-col class="ma-0 pa-0" cols="9" sm="10" md="11">
          <PhytofyLevelSlider
            :value="totalUncommitted"
            :max="300"
            :limit="totalLimit"
            color="#000000"
            @input="
              (newValue) =>
                setTotalAndLevels(
                  totalCommitted,
                  newValue,
                  totalLimit,
                  value,
                  false
                )
            "
            @change="
              (newValue) =>
                setTotalAndLevels(
                  totalCommitted,
                  newValue,
                  totalLimit,
                  value,
                  true
                )
            "
          />
        </v-col>
      </v-row>
    </v-col>
  </v-row>
</template>

<script lang="ts">
import Vue from "vue";
import PhytofyLevelField from "@/components/PhytofyLevelField.vue";
import PhytofyLevelSlider from "@/components/PhytofyLevelSlider.vue";

const LEVEL_STEP = 0.1;

const LEVEL_TOTAL_LIMIT = 300.0;

const trimDecimals = (value: number) => {
  return Math.round(value * 10.0) / 10.0;
};

const calculateLevels = (
  levels: number[],
  oldTotal: number,
  newTotal: number
) => {
  return levels
    .slice()
    .map((level: number) =>
      newTotal === 0 ? 0 : trimDecimals((level * newTotal) / oldTotal)
    );
};

const calculateTotal = (levels: number[]) =>
  trimDecimals(levels.reduce((a, b) => a + b, 0));

const calculateLevelLimits = (levels: number[], total: number) =>
  levels.map((level: number) =>
    Math.min(100.0, level + LEVEL_TOTAL_LIMIT - total)
  );

const calculateTotalLimit = (levels: number[], total: number) => {
  const maximum: number = Math.max(...levels);
  let totalLimit = 0.0;
  if (total !== 0) {
    totalLimit = trimDecimals(
      Math.min(100.0 / maximum, LEVEL_TOTAL_LIMIT / total) * total
    );
  }
  return totalLimit;
};

export default Vue.extend({
  name: "PhytofyLevels",

  components: {
    PhytofyLevelField,
    PhytofyLevelSlider,
  },

  props: {
    value: {
      type: Array,
      required: true,
    },
  },

  data: () => ({
    channels: [
      { hint: "UV-A", color: "#4A148C" },
      { hint: "Blue", color: "#0D47A1" },
      { hint: "Green", color: "#1B5E20" },
      { hint: "Hyper R.", color: "#B71C1C" },
      { hint: "Far R.", color: "#880E4F" },
      { hint: "White", color: "#F57F17" },
    ],
    levelStep: LEVEL_STEP,
    levelsUncommitted: [0, 0, 0, 0, 0, 0],
    totalUncommitted: 0,
    totalCommitted: 0,
    levelLimits: [100, 100, 100, 100, 100, 100],
    totalLimit: LEVEL_TOTAL_LIMIT,
  }),

  mounted() {
    this.apply(this.$props.value);
  },

  watch: {
    value(newValue) {
      this.apply(newValue);
    },
  },

  methods: {
    apply(newValue: number[]) {
      const initialLevels = newValue.slice();
      const initialTotal = calculateTotal(initialLevels);
      this.levelsUncommitted = initialLevels;
      this.totalUncommitted = initialTotal;
      this.totalCommitted = initialTotal;
      this.levelLimits = calculateLevelLimits(initialLevels, initialTotal);
      this.totalLimit = calculateTotalLimit(initialLevels, initialTotal);
    },

    setLevelsCommitted(levelsCommitted: number[]) {
      this.$emit("input", levelsCommitted);
    },

    setLevelsUncommitted(levelsUncommitted: number[]) {
      this.levelsUncommitted = levelsUncommitted;
    },

    setTotalUncommitted(totalUncommitted: number) {
      this.totalUncommitted = totalUncommitted;
    },

    setTotalCommitted(totalCommitted: number) {
      this.totalCommitted = totalCommitted;
    },

    setLevelLimits(levelLimits: number[]) {
      this.levelLimits = levelLimits;
    },

    setTotalLimit(totalLimit: number) {
      this.totalLimit = totalLimit;
    },

    adjustLevelLimits(levels: number[], total: number) {
      this.setLevelLimits(calculateLevelLimits(levels, total));
    },

    adjustTotalLimit(levels: number[], total: number) {
      this.setTotalLimit(calculateTotalLimit(levels, total));
    },

    adjustLevels(levels: number[], commit: boolean) {
      this.setLevelsUncommitted(levels);
      if (commit) {
        this.setLevelsCommitted(levels);
      }
    },

    adjustTotal(total: number, commit: boolean) {
      this.setTotalUncommitted(total);
      if (commit) {
        this.setTotalCommitted(total);
      }
    },

    adjust(levels: number[], total: number, commit: boolean, fixed: boolean) {
      this.adjustLevels(levels, commit);
      this.adjustTotal(total, commit);
      this.adjustLevelLimits(levels, total);
      if (!fixed) {
        this.adjustTotalLimit(levels, total);
      }
    },

    setLevel(
      level: number,
      levels: number[],
      levelLimits: number[],
      index: number,
      commit: boolean
    ) {
      level = Number(level.toString().replace(/[^0-9.]/g, ""));
      if (isNaN(level)) {
        return;
      }
      level = level < 0 ? 0 : level;
      level = levelLimits[index] < level ? levelLimits[index] : level;
      const adjustedLevels = levels.slice();
      adjustedLevels[index] = level;
      const adjustedTotal = calculateTotal(adjustedLevels);
      this.adjust(adjustedLevels, adjustedTotal, commit, false);
    },

    setTotalAndLevels(
      oldTotal: number,
      newTotal: number,
      totalLimit: number,
      levels: number[],
      commit: boolean
    ) {
      if (newTotal < 0 || totalLimit < newTotal) {
        return;
      }
      const adjustedLevels = calculateLevels(levels, oldTotal, newTotal);
      this.adjust(adjustedLevels, newTotal, commit, true);
    },
  },
});
</script>
