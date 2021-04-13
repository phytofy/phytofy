<!-- Copyright (c) 2020 OSRAM; Licensed under the MIT license. -->

<template>
  <v-main>
    <v-card class="d-flex flex-column ma-0 pa-4" flat>
      <v-card outlined max-width="300">
        <v-card-title>Logs</v-card-title>
        <v-card-text>
          Download a ZIP file containing the recent logs. There, you can check
          for warning and error messages collected during operation.
        </v-card-text>
        <v-card-actions>
          <v-btn icon @click="logs"><v-icon>mdi-bug</v-icon></v-btn>
        </v-card-actions>
      </v-card>
    </v-card>
  </v-main>
</template>

<script lang="ts">
import Vue from "vue";

export default Vue.extend({
  name: "PhytofyTools",

  methods: {
    logs() {
      const element = document.createElement("a");
      element.setAttribute("href", "/api/logs");
      element.setAttribute("download", "PHYTOFY.logs.zip");
      element.style.display = "none";
      document.body.appendChild(element);
      element.click();
      document.body.removeChild(element);
    },
  },
});
</script>
