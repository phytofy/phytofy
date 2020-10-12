<!-- Copyright (c) 2020 OSRAM; Licensed under the MIT license. -->

<template>
  <v-app>
    <PhytofyNavigationDrawer v-model="drawer" />
    <PhytofyAppBar v-model="drawer" />
    <PhytofyView />
    <PhytofyEula />
  </v-app>
</template>

<script lang="ts">
import Vue from "vue";
import PhytofyAppBar from "@/components/PhytofyAppBar.vue";
import PhytofyNavigationDrawer from "@/components/PhytofyNavigationDrawer.vue";
import PhytofyView from "@/components/PhytofyView.vue";
import PhytofyEula from "@/components/PhytofyEula.vue";
import * as api from "./api";

export default Vue.extend({
  name: "PhytofyApp",

  components: {
    PhytofyAppBar,
    PhytofyNavigationDrawer,
    PhytofyView,
    PhytofyEula,
  },

  data: () => ({
    drawer: true,
    api: (null as unknown) as api.DefaultApi,
  }),

  mounted() {
    const config: api.Configuration = {
      basePath: `${window.location.origin}/api`,
    };
    this.api = new api.DefaultApi(config);
    setInterval(() => {
      this.api
        .apiGetSerials({})
        .then((result) => {
          this.$store.commit("setSerials", result.serials);
        })
        .catch((/* error: Error */) => {
          // console.log(error); // Ignoring the error
        });
    }, 5000);
  },
});
</script>
