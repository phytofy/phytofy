<!-- Copyright (c) 2020 OSRAM; Licensed under the MIT license. -->

<template>
  <v-dialog v-model="$props.visible" persistent scrollable>
    <v-card>
      <v-card-title>Schedule</v-card-title>
      <v-divider></v-divider>
      <v-card-text>
        <v-form ref="schedule" v-model="valid">
          <v-row>
            <v-col align="center">
              <v-row class="no-gutters">
                <v-col>
                  <v-card class="d-flex justify-space-between" flat tile>
                    <v-menu
                      ref="startDateMenuRef"
                      v-model="startDateMenuVisible"
                      :close-on-content-click="false"
                      :return-value.sync="$props.value.startDate"
                      transition="scale-transition"
                      offset-y
                      min-width="290px"
                    >
                      <template v-slot:activator="{ on, attrs }">
                        <v-text-field
                          v-model="$props.value.startDate"
                          label="Start Date"
                          prepend-icon="mdi-calendar"
                          readonly
                          v-bind="attrs"
                          v-on="on"
                          :rules="[datesRule]"
                        ></v-text-field>
                      </template>
                      <v-date-picker
                        v-model="$props.value.startDate"
                        @change="validate"
                        scrollable
                      >
                        <v-spacer></v-spacer>
                        <v-btn
                          text
                          color="primary"
                          @click="startDateMenuVisible = false"
                          >Cancel</v-btn
                        >
                        <v-btn
                          text
                          color="primary"
                          @click="
                            $refs.startDateMenuRef.save($props.value.startDate)
                          "
                          >OK</v-btn
                        >
                      </v-date-picker>
                    </v-menu>
                    <v-menu
                      ref="stopDateMenuRef"
                      v-model="stopDateMenuVisible"
                      :close-on-content-click="false"
                      :return-value.sync="$props.value.stopDate"
                      transition="scale-transition"
                      offset-y
                      min-width="290px"
                    >
                      <template v-slot:activator="{ on, attrs }">
                        <v-text-field
                          v-model="$props.value.stopDate"
                          label="Stop Date"
                          prepend-icon="mdi-calendar"
                          readonly
                          v-bind="attrs"
                          v-on="on"
                          :rules="[datesRule]"
                        ></v-text-field>
                      </template>
                      <v-date-picker
                        v-model="$props.value.stopDate"
                        @change="validate"
                        scrollable
                      >
                        <v-spacer></v-spacer>
                        <v-btn
                          text
                          color="primary"
                          @click="stopDateMenuVisible = false"
                          >Cancel</v-btn
                        >
                        <v-btn
                          text
                          color="primary"
                          @click="
                            $refs.stopDateMenuRef.save($props.value.stopDate)
                          "
                          >OK</v-btn
                        >
                      </v-date-picker>
                    </v-menu>
                    <v-menu
                      ref="startTimeMenuRef"
                      v-model="startTimeMenuVisible"
                      :close-on-content-click="false"
                      :return-value.sync="$props.value.startTime"
                      transition="scale-transition"
                      offset-y
                      min-width="290px"
                    >
                      <template v-slot:activator="{ on, attrs }">
                        <v-text-field
                          v-model="$props.value.startTime"
                          label="Start Time"
                          prepend-icon="mdi-clock-outline"
                          readonly
                          v-bind="attrs"
                          v-on="on"
                          :rules="[timesRule]"
                        ></v-text-field>
                      </template>
                      <v-time-picker
                        v-model="$props.value.startTime"
                        @change="validate"
                        full-width
                        format="24hr"
                      >
                        <v-spacer></v-spacer>
                        <v-btn
                          text
                          color="primary"
                          @click="startTimeMenuVisible = false"
                          >Cancel</v-btn
                        >
                        <v-btn
                          text
                          color="primary"
                          @click="
                            $refs.startTimeMenuRef.save($props.value.startTime)
                          "
                          >OK</v-btn
                        >
                      </v-time-picker>
                    </v-menu>
                    <v-menu
                      ref="stopTimeMenuRef"
                      v-model="stopTimeMenuVisible"
                      :close-on-content-click="false"
                      :return-value.sync="$props.value.stopTime"
                      transition="scale-transition"
                      offset-y
                      min-width="290px"
                    >
                      <template v-slot:activator="{ on, attrs }">
                        <v-text-field
                          v-model="$props.value.stopTime"
                          label="Stop Time"
                          prepend-icon="mdi-clock-outline"
                          readonly
                          v-bind="attrs"
                          v-on="on"
                          :rules="[timesRule]"
                        ></v-text-field>
                      </template>
                      <v-time-picker
                        v-model="$props.value.stopTime"
                        @change="validate"
                        full-width
                        format="24hr"
                      >
                        <v-spacer></v-spacer>
                        <v-btn
                          text
                          color="primary"
                          @click="stopTimeMenuVisible = false"
                          >Cancel</v-btn
                        >
                        <v-btn
                          text
                          color="primary"
                          @click="
                            $refs.stopTimeMenuRef.save($props.value.stopTime)
                          "
                          >OK</v-btn
                        >
                      </v-time-picker>
                    </v-menu>
                    <v-select
                      :items="preexisting"
                      v-model="$props.value.serial"
                      label="Serial Number"
                      prepend-icon="mdi-white-balance-iridescent"
                    ></v-select>
                  </v-card>
                </v-col>
              </v-row>
              <PhytofyLevels v-model="$props.value.levels" />
            </v-col>
          </v-row>
        </v-form>
      </v-card-text>
      <v-divider></v-divider>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn text @click="cancel">Cancel</v-btn>
        <v-btn text @click="save" :disabled="!valid || empty()">Save</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script lang="ts">
import Vue from "vue";
import { mapState } from "vuex";
import PhytofyLevels from "@/components/PhytofyLevels.vue";

const inOrder = (array: number[]) => {
  for (let i = 1; i < array.length; i++) {
    if (array[i - 1] > array[i]) {
      return false;
    }
  }
  return true;
};

const timesInOrder = (times: string[]) =>
  inOrder(times.map((time: string) => Date.parse(`1970-01-01T${time}:00Z`)));

const datesInOrder = (dates: string[]) =>
  inOrder(dates.map((date: string) => Date.parse(date)));

const skipEmpty = (array: string[]) => array.filter((item) => item !== "");

export default Vue.extend({
  name: "PhytofySchedule",

  components: {
    PhytofyLevels,
  },

  props: {
    value: {
      type: Object,
      required: true,
    },
    visible: {
      type: Boolean,
      required: true,
    },
  },

  computed: {
    ...mapState(["serials"]),

    preexisting() {
      const serials = [...new Set(this.serials)];
      if (this.$props.value.serial !== null) {
        serials.push(this.$props.value.serial);
      }
      return serials.sort();
    },
  },

  data: () => ({
    startDateMenuVisible: false,
    startTimeMenuVisible: false,
    stopDateMenuVisible: false,
    stopTimeMenuVisible: false,
    valid: false,
  }),

  watch: {
    value: {
      immediate: true,
      handler() {
        this.$nextTick(() => {
          if (typeof this.$refs.schedule !== "undefined") {
            (this.$refs.schedule as Vue & {
              validate: () => boolean;
            }).validate();
          }
        });
      },
    },
  },

  methods: {
    cancel() {
      this.$emit("cancel");
    },

    save() {
      this.$emit("save");
    },

    validate() {
      this.$nextTick(() => {
        if (typeof this.$refs.schedule !== "undefined") {
          (this.$refs.schedule as Vue & {
            validate: () => boolean;
          }).validate();
        }
      });
    },

    empty() {
      const startDateEmpty = this.$props.value.startDate === "";
      const startTimeEmpty = this.$props.value.startTime === "";
      const stopDateEmpty = this.$props.value.stopDate === "";
      const stopTimeEmpty = this.$props.value.stopTime === "";
      return startDateEmpty || startTimeEmpty || stopDateEmpty || stopTimeEmpty;
    },

    timesRule() {
      return (
        timesInOrder(
          skipEmpty([this.$props.value.startTime, this.$props.value.stopTime])
        ) || "Stop time must follow start time"
      );
    },

    datesRule() {
      return (
        datesInOrder(
          skipEmpty([this.$props.value.startDate, this.$props.value.stopDate])
        ) || "Stop date must follow start date"
      );
    },
  },
});
</script>
