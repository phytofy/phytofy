<!-- Copyright (c) 2020 OSRAM; Licensed under the MIT license. -->

<template>
  <v-main>
    <span ref="zone" id="zone">
      <v-card class="d-flex flex-column ma-0 pa-0 fill-height" flat>
        <v-card class="ma-0 pa-0" align="center" flat>
          <v-container class="pa-0 ma-0" fluid>
            <v-row>
              <v-col>
                <v-tooltip top>
                  <template v-slot:activator="{ on, attrs }">
                    <v-btn
                      @click="createSchedule"
                      :disabled="!ready"
                      icon
                      small
                      v-bind="attrs"
                      v-on="on"
                    >
                      <v-icon>mdi-plus</v-icon>
                    </v-btn>
                  </template>
                  <span>Create</span>
                </v-tooltip>
              </v-col>
              <v-col>
                <v-tooltip top>
                  <template v-slot:activator="{ on, attrs }">
                    <v-btn
                      @click="editSchedule"
                      :disabled="!ready || !selectedOne"
                      icon
                      small
                      v-bind="attrs"
                      v-on="on"
                    >
                      <v-icon>mdi-pencil</v-icon>
                    </v-btn>
                  </template>
                  <span>Edit</span>
                </v-tooltip>
              </v-col>
              <v-col>
                <v-tooltip top>
                  <template v-slot:activator="{ on, attrs }">
                    <v-btn
                      @click="deleteSchedules"
                      :disabled="!ready || selectedNone"
                      icon
                      small
                      v-bind="attrs"
                      v-on="on"
                    >
                      <v-icon>mdi-delete</v-icon>
                    </v-btn>
                  </template>
                  <span>Delete</span>
                </v-tooltip>
              </v-col>
              <v-col>
                <input
                  v-show="false"
                  type="file"
                  ref="upload"
                  id="upload"
                  @change="importSchedules"
                />
                <v-tooltip top>
                  <template v-slot:activator="{ on, attrs }">
                    <v-btn
                      @click="$refs.upload.click()"
                      :disabled="!ready"
                      icon
                      small
                      v-bind="attrs"
                      v-on="on"
                    >
                      <v-icon>mdi-publish</v-icon>
                    </v-btn>
                  </template>
                  <span>Import</span>
                </v-tooltip>
              </v-col>
              <v-col>
                <v-tooltip top>
                  <template v-slot:activator="{ on, attrs }">
                    <v-btn
                      @click="exportSchedules"
                      :disabled="!ready"
                      icon
                      small
                      v-bind="attrs"
                      v-on="on"
                    >
                      <v-icon>mdi-download</v-icon>
                    </v-btn>
                  </template>
                  <span>Export</span>
                </v-tooltip>
              </v-col>
              <v-col>
                <v-tooltip top>
                  <template v-slot:activator="{ on, attrs }">
                    <v-btn
                      @click="applySchedules"
                      :disabled="!ready"
                      icon
                      small
                      v-bind="attrs"
                      v-on="on"
                    >
                      <v-icon>mdi-send</v-icon>
                    </v-btn>
                  </template>
                  <span>Apply</span>
                </v-tooltip>
              </v-col>
            </v-row>
          </v-container>
        </v-card>
        <v-card class="ma-0 pa-0 flex-grow-1 d-flex" flat>
          <v-row>
            <v-col xs="12" class="d-flex">
              <div
                class="flex-grow-1"
                v-resize="rightSizeSchedules"
                v-mutate="rightSizeSchedules"
                ref="schedulesBinder"
              >
                <v-simple-table
                  fixed-header
                  dense
                  style="border-radius: 0"
                  :height="schedulesHeight"
                >
                  <template v-slot:default>
                    <thead>
                      <tr>
                        <th class="text-left">
                          <v-checkbox
                            :value="selectedAll"
                            :indeterminate="!(selectedAll || selectedNone)"
                            @click="toggleGlobalSelection"
                          ></v-checkbox>
                        </th>
                        <th class="text-left">Dates</th>
                        <th class="text-left">Times</th>
                        <th class="text-left">Channel Levels</th>
                        <th class="text-left">Serial Number</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="(schedule, idx) in schedules" :key="idx">
                        <td>
                          <v-checkbox v-model="schedule.selected"></v-checkbox>
                        </td>
                        <td>
                          {{ schedule.startDate }} - {{ schedule.stopDate }}
                        </td>
                        <td>
                          {{ schedule.startTime }} - {{ schedule.stopTime }}
                        </td>
                        <td>{{ schedule.levels.join(",") }}</td>
                        <td>{{ schedule.serial }}</td>
                      </tr>
                    </tbody>
                  </template>
                </v-simple-table>
              </div>
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

interface Schedule {
  id: string;
  selected: boolean;
  startDate: string;
  stopDate: string;
  startTime: string;
  stopTime: string;
  levels: number[];
  serial: number | null;
}

const blankSchedule = (): Schedule => ({
  id: uuidv4(),
  selected: false,
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
      selected: false,
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

const filterSelectedSchedules = (schedules: Schedule[]): Schedule[] => {
  return schedules.filter((schedule: Schedule) => schedule.selected);
};

const fillSelectedSchedules = (schedules: Schedule[], value: boolean): void => {
  schedules.forEach((schedule: Schedule) => {
    schedule.selected = value;
  });
};

export default Vue.extend({
  name: "PhytofySchedules",

  components: {
    PhytofySchedule,
  },

  data: () => ({
    schedule: blankSchedule(),
    schedules: [] as Schedule[],
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
        text: "Channel Levels",
        align: "start",
        sortable: false,
        value: "levels",
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
    schedulesHeight: 0,
  }),

  computed: {
    selectedSchedules() {
      return filterSelectedSchedules(this.schedules);
    },

    selectedOne() {
      return filterSelectedSchedules(this.schedules).length === 1;
    },

    selectedNone() {
      return filterSelectedSchedules(this.schedules).length === 0;
    },

    selectedAll() {
      return (
        filterSelectedSchedules(this.schedules).length === this.schedules.length
      );
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
    rightSizeSchedules() {
      this.schedulesHeight =
        window.innerHeight -
        (this.$refs.schedulesBinder as Element).getBoundingClientRect().top;
    },

    toggleGlobalSelection() {
      if (this.selectedAll) {
        fillSelectedSchedules(this.selectedSchedules, false);
      } else {
        fillSelectedSchedules(this.schedules, true);
      }
    },

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
      this.schedules = this.schedules.filter(
        (schedule: Schedule) => !schedule.selected
      );
      fillSelectedSchedules(this.schedules, false);
    },

    enterEditingSchedule(create: boolean) {
      if (create) {
        this.schedule = blankSchedule();
        this.editing = true;
      } else {
        if (this.selectedOne) {
          this.schedule = { ...this.selectedSchedules[0] };
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
      fillSelectedSchedules(this.schedules, false);
      this.editing = false;
    },
  },
});
</script>
