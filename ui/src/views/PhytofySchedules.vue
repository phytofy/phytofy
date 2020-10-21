<!-- Copyright (c) 2020 OSRAM; Licensed under the MIT license. -->

<template>
  <v-main>
    <span ref="zone" id="zone">
      <v-card class="d-flex flex-column ma-0 pa-0 fill-height" flat>
        <v-card class="mx-4 mt-4 mb-2 px-4 py-2" align="center">
          <v-container class="pa-0 ma-0" fluid>
            <v-row>
              <v-col lg="2" sm="4">
                <v-btn @click="createSchedule" :disabled="!ready">Create</v-btn>
              </v-col>
              <v-col lg="2" sm="4">
                <v-btn
                  @click="editSchedule"
                  :disabled="!ready || selectedNotOneSchedule"
                  >Edit</v-btn
                >
              </v-col>
              <v-col lg="2" sm="4">
                <v-btn
                  @click="deleteSchedules"
                  :disabled="!ready || selectedNoSchedule"
                  >Delete</v-btn
                >
              </v-col>
              <v-col lg="2" sm="4">
                <input
                  v-show="false"
                  type="file"
                  ref="upload"
                  id="upload"
                  @change="importSchedules"
                />
                <v-btn @click="$refs.upload.click()" :disabled="!ready">{{
                  !ready ? "Please Wait" : dragging ? "Drop CSV" : "Import CSV"
                }}</v-btn>
              </v-col>
              <v-col lg="2" sm="4">
                <v-btn @click="exportSchedules" :disabled="!ready"
                  >Export CSV</v-btn
                >
              </v-col>
              <v-col lg="2" sm="4">
                <v-btn @click="applySchedules" :disabled="!ready">Apply</v-btn>
              </v-col>
            </v-row>
          </v-container>
        </v-card>
        <v-card class="mx-4 mt-2 mb-4 px-4 py-2 flex-grow-1 d-flex">
          <v-row>
            <v-col xs="12" class="d-flex">
              <v-data-table
                class="flex-grow-1"
                v-model="selected"
                show-select
                :headers="headers"
                :items="schedules"
                :items-per-page="10"
              />
            </v-col>
          </v-row>
        </v-card>
      </v-card>
    </span>
    <PhytofySchedule
      v-model="schedule"
      :visible="editing"
      @cancel="cancelSchedule"
      @save="saveSchedule"
    />
    <v-snackbar
      v-model="error"
      color="error"
      :multi-line="true"
      :timeout="60000"
    >
      {{ errorMessage }}
      <v-btn @click="error = false">Dismiss</v-btn>
    </v-snackbar>
  </v-main>
</template>

<script lang="ts">
import Vue from "vue";
import Plotly from "plotly.js";
import PhytofySchedule from "@/components/PhytofySchedule.vue";
import { v4 as uuidv4 } from "uuid";
import * as api from "../api";
import { ImportSchedulesRequest } from "../api";

interface Schedule {
  id: string;
  startDate: string;
  stopDate: string;
  startTime: string;
  stopTime: string;
  levels: number[];
  serial: number | null;
}

const blankSchedule = (): Schedule => ({
  id: uuidv4(),
  startDate: "",
  stopDate: "",
  startTime: "",
  stopTime: "",
  levels: [0, 0, 0, 0, 0, 0],
  serial: null,
});

const unpackSchedule = (entry: string[]): Schedule[] => {
  return entry
    .slice(10, entry.length)
    .filter((serial) => serial !== "")
    .map((serial) => ({
      id: uuidv4(),
      startDate: entry[0],
      stopDate: entry[1],
      startTime: entry[2],
      stopTime: entry[3],
      levels: entry.slice(4, 10).map((level) => Number(level)),
      serial: Number(serial),
    }));
};

const packSchedule = (schedule: Schedule): string => {
  const fields = [
    schedule.startDate,
    schedule.stopDate,
    schedule.startTime,
    schedule.stopTime,
    schedule.levels.map((level: number) => level.toString()),
    schedule.serial === null ? "" : schedule.serial.toString(),
  ];
  return fields.join();
};

export default Vue.extend({
  name: "PhytofySchedules",

  components: {
    PhytofySchedule,
  },

  data: () => ({
    schedule: blankSchedule(),
    schedules: [] as Schedule[],
    selected: [],
    headers: [
      {
        text: "Start Date",
        align: "start",
        sortable: true,
        value: "startDate",
      },
      {
        text: "Stop Date",
        align: "start",
        sortable: true,
        value: "stopDate",
      },
      {
        text: "Start Time",
        align: "start",
        sortable: true,
        value: "startTime",
      },
      {
        text: "Stop Time",
        align: "start",
        sortable: true,
        value: "stopTime",
      },
      {
        text: "UV-A",
        align: "start",
        sortable: false,
        value: "levels[0]",
      },
      {
        text: "Blue",
        align: "start",
        sortable: false,
        value: "levels[1]",
      },
      {
        text: "Green",
        align: "start",
        sortable: false,
        value: "levels[2]",
      },
      {
        text: "Hyper Red",
        align: "start",
        sortable: false,
        value: "levels[3]",
      },
      {
        text: "Far Red",
        align: "start",
        sortable: false,
        value: "levels[4]",
      },
      {
        text: "White",
        align: "start",
        sortable: false,
        value: "levels[5]",
      },
      {
        text: "Serial Number",
        align: "start",
        sortable: true,
        value: "serial",
      },
    ],
    editing: false,
    dragging: false,
    ready: true,
    error: false,
    errorMessage: "",
    api: (null as unknown) as api.DefaultApi,
  }),

  computed: {
    selectedNotOneSchedule() {
      return this.selected.length !== 1;
    },

    selectedNoSchedule() {
      return this.selected.length === 0;
    },
  },

  mounted() {
    const config: api.Configuration = {
      basePath: `${window.location.origin}/api`,
    };
    this.api = new api.DefaultApi(config);
    const zone = this.$refs.zone as HTMLElement;
    zone.addEventListener("dragenter", this.dragEnter);
    zone.addEventListener("dragover", this.dragOver);
    zone.addEventListener("dragleave", this.dragLeave);
    zone.addEventListener("drop", this.drop);
  },

  methods: {
    createSchedule() {
      this.enterEditingSchedule(true);
    },

    editSchedule() {
      this.enterEditingSchedule(false);
    },

    saveSchedule() {
      this.leaveEditingSchedule(true);
    },

    cancelSchedule() {
      this.leaveEditingSchedule(false);
    },

    dragEnter(event: DragEvent) {
      if (event === null) return;
      event.preventDefault();
      this.dragging = true;
    },

    dragOver(event: DragEvent) {
      if (event === null) return;
      event.preventDefault();
      this.dragging = true;
    },

    dragLeave(event: DragEvent) {
      if (event === null) return;
      event.preventDefault();
      this.dragging = false;
    },

    drop(event: DragEvent) {
      if (event === null || event.dataTransfer === null) return;
      event.preventDefault();
      this.dragging = false;
      if (event.dataTransfer.items) {
        const files = [];
        for (let i = 0; i < event.dataTransfer.items.length; i++) {
          if (event.dataTransfer.items[i].kind === "file") {
            const file = event.dataTransfer.items[i].getAsFile();
            files.push(file);
          }
        }
        this.loadSchedules(files as File[]);
      } else {
        this.loadSchedules(Array.from(event.dataTransfer.files));
      }
    },

    importSchedules() {
      const uploadElement = this.$refs.upload as HTMLInputElement;
      const files = uploadElement.files;
      if (files !== null) {
        this.loadSchedules(Array.from(files));
      }
    },

    replaceSchedules(csv: string[][]) {
      const nested = csv.map((entry) => unpackSchedule(entry)) as Schedule[][];
      this.schedules = ([] as Schedule[]).concat(...nested);
    },

    loadSchedules(files: File[]) {
      if (!this.ready) {
        return;
      }
      const reader = new FileReader();
      let counter = 0;
      let csv = "";
      reader.onload = () => {
        csv += reader.result;
        counter++;
        if (counter == files.length) {
          this.replaceSchedules(Plotly.d3.csv.parseRows(csv));
        }
      };
      files.forEach((entry: File) => {
        reader.readAsText(entry);
      });
    },

    exportSchedules() {
      const data = this.schedules
        .map((schedule: Schedule) => packSchedule(schedule))
        .join("\n");
      const url = window.URL.createObjectURL(
        new Blob([data], { type: "text/csv" })
      );
      const link = document.createElement("a");
      link.href = url;
      link.setAttribute("download", "schedules.csv");
      link.click();
    },

    applySchedules() {
      this.ready = false;
      const entries: api.Schedules = [];
      for (let i = 0; i < this.schedules.length; i++) {
        const schedule = this.schedules[i];
        const levels: api.Levels = schedule.levels;
        const start: api.Time =
          Date.parse(schedule.startDate + "T" + schedule.startTime) / 1000;
        const stop: api.Time =
          Date.parse(schedule.stopDate + "T" + schedule.stopTime) / 1000;
        if (schedule.serial === null) {
          continue;
        }
        const serial: api.Serial = schedule.serial;
        const serials: api.Serials = [serial];
        const entry: api.Schedule = { start, stop, levels, serials };
        entries.push(entry);
      }
      const request: api.ImportSchedulesRequest = { schedules: entries };
      this.api
        .apiImportSchedules(request)
        .then((response: api.ImportSchedulesReply) => {
          this.ready = true;
          if (response.error !== undefined) {
            this.errorMessage = response.error;
            this.error = true;
          }
        })
        .catch((response) => {
          this.ready = true;
          response
            .json()
            .then((response: api.ImportSchedulesReply) => {
              this.errorMessage =
                response.error !== undefined
                  ? response.error
                  : "Unknown failure";
              this.error = true;
            })
            .catch((error: Error) => {
              this.errorMessage = error.toString();
              this.error = true;
            });
        });
    },

    deleteSchedules() {
      const selected = this.selectedSchedules();
      this.schedules = this.schedules.filter(
        (schedule: Schedule) => selected.indexOf(schedule.id) === -1
      );
      this.selected = [];
    },

    selectedSchedules() {
      return this.selected.map((schedule: Schedule) => schedule.id);
    },

    enterEditingSchedule(create: boolean) {
      if (create) {
        this.schedule = blankSchedule();
        this.editing = true;
      } else {
        if (this.selected.length === 1) {
          this.schedule = { ...(this.selected[0] as Schedule) };
          this.editing = true;
        }
      }
    },

    leaveEditingSchedule(commit: boolean) {
      if (commit) {
        const ids = this.schedules.map((schedule: Schedule) => schedule.id);
        if (ids.indexOf(this.schedule.id) === -1) {
          this.schedules.push(this.schedule);
        } else {
          this.schedules = this.schedules.map((schedule: Schedule) =>
            this.schedule.id === schedule.id ? this.schedule : schedule
          );
        }
      }
      this.selected = [];
      this.editing = false;
    },
  },
});
</script>
