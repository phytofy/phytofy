<!-- Copyright (c) 2020 OSRAM; Licensed under the MIT license. -->

<template>
  <v-dialog v-model="$props.visible" persistent scrollable>
    <v-card>
      <v-card-text>
        <v-form ref="schedule" v-model="valid">
          <v-row class="no-gutters py-1">
            <v-col class="ma-0 pl-0 pr-2 py-0" cols="6" sm="3">
              <v-dialog
                ref="startDateRef"
                v-model="startDateVisible"
                :return-value.sync="$props.value.startDate"
                transition="scale-transition"
                persistent
              >
                <template v-slot:activator="{ on, attrs }">
                  <v-text-field
                    v-model="$props.value.startDate"
                    hint="Start Date"
                    persistent-hint
                    readonly
                    v-bind="attrs"
                    v-on="on"
                    dense
                    :rules="[datesRule]"
                  ></v-text-field>
                </template>
                <v-date-picker
                  v-if="startDateVisible"
                  v-model="$props.value.startDate"
                  @change="validate"
                  @input="$refs.startDateRef.save($props.value.startDate)"
                  full-width
                  :landscape="landscape()"
                ></v-date-picker>
              </v-dialog>
            </v-col>
            <v-col class="ma-0 pl-0 pr-2 py-0" cols="6" sm="3">
              <v-dialog
                ref="stopDateMenuRef"
                v-model="stopDateVisible"
                :close-on-content-click="false"
                :return-value.sync="$props.value.stopDate"
                transition="scale-transition"
                persistent
              >
                <template v-slot:activator="{ on, attrs }">
                  <v-text-field
                    v-model="$props.value.stopDate"
                    hint="Stop Date"
                    persistent-hint
                    readonly
                    v-bind="attrs"
                    v-on="on"
                    dense
                    :rules="[datesRule]"
                  ></v-text-field>
                </template>
                <v-date-picker
                  v-if="stopDateVisible"
                  v-model="$props.value.stopDate"
                  @change="validate"
                  @input="$refs.stopDateMenuRef.save($props.value.stopDate)"
                  full-width
                  :landscape="landscape()"
                ></v-date-picker>
              </v-dialog>
            </v-col>
            <v-col class="ma-0 pl-0 pr-2 py-0" cols="6" sm="3">
              <v-dialog
                ref="startTimeMenuRef"
                v-model="startTimeVisible"
                :return-value.sync="$props.value.startTime"
                transition="scale-transition"
                persistent
              >
                <template v-slot:activator="{ on, attrs }">
                  <v-text-field
                    v-model="$props.value.startTime"
                    hint="Start Time"
                    persistent-hint
                    readonly
                    v-bind="attrs"
                    v-on="on"
                    dense
                    :rules="[timesRule]"
                  ></v-text-field>
                </template>
                <v-time-picker
                  v-if="startTimeVisible"
                  v-model="$props.value.startTime"
                  @change="validate"
                  @input="$refs.startTimeMenuRef.save($props.value.startTime)"
                  format="24hr"
                  full-width
                  :landscape="landscape()"
                ></v-time-picker>
              </v-dialog>
            </v-col>
            <v-col class="ma-0 pl-0 pr-2 py-0" cols="6" sm="3">
              <v-dialog
                ref="stopTimeMenuRef"
                v-model="stopTimeVisible"
                :return-value.sync="$props.value.stopTime"
                transition="scale-transition"
                persistent
              >
                <template v-slot:activator="{ on, attrs }">
                  <v-text-field
                    v-model="$props.value.stopTime"
                    hint="Stop Time"
                    persistent-hint
                    readonly
                    v-bind="attrs"
                    v-on="on"
                    dense
                    :rules="[timesRule]"
                  ></v-text-field>
                </template>
                <v-time-picker
                  v-if="stopTimeVisible"
                  v-model="$props.value.stopTime"
                  @change="validate"
                  @input="$refs.stopTimeMenuRef.save($props.value.stopTime)"
                  format="24hr"
                  full-width
                  :landscape="landscape()"
                ></v-time-picker>
              </v-dialog>
            </v-col>
            <v-col cols="12">
              <v-select
                :items="preexisting"
                v-model="$props.value.serial"
                hint="Serial Number"
                persistent-hint
                dense
              ></v-select>
            </v-col>
            <v-col cols="12">
              <PhytofyLevels v-model="$props.value.levels" />
            </v-col>
          </v-row>
        </v-form>
      </v-card-text>
      <v-divider></v-divider>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn text @click="cancel" dense>Cancel</v-btn>
        <v-btn text @click="save" :disabled="!valid || empty()" dense
          >Save</v-btn
        >
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
    startDateVisible: false,
    startTimeVisible: false,
    stopDateVisible: false,
    stopTimeVisible: false,
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

    landscape() {
      return window.innerWidth > window.innerHeight;
    },
  },
});
</script>
