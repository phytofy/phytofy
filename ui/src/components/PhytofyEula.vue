<!-- Copyright (c) 2020 OSRAM; Licensed under the MIT license. -->

<template>
  <v-dialog v-model="eulaVisible" persistent scrollable>
    <v-card>
      <v-card-title>End User Licensing Agreement</v-card-title>
      <v-divider></v-divider>
      <v-card-text>
        <div v-html="eula"></div>
      </v-card-text>
      <v-divider></v-divider>
      <v-card-actions>
        <v-checkbox
          v-model="eulaAccepted"
          :disabled="eulaAccepted"
          label="I accept"
        ></v-checkbox>
        <v-spacer></v-spacer>
        <v-btn text :disabled="!eulaAccepted" @click="close">Close</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script lang="ts">
import Vue from "vue";

export default Vue.extend({
  name: "PhytofyEula",

  data() {
    return {
      publicPath: process.env.BASE_URL,
      eula: "",
    };
  },

  computed: {
    eulaVisible: {
      get() {
        return this.$store.state.eulaVisible;
      },
      set(eulaVisible) {
        this.$store.commit("setEulaVisible", eulaVisible);
      },
    },

    eulaAccepted: {
      get() {
        return this.$store.state.eulaAccepted;
      },
      set(eulaAccepted) {
        this.$store.commit("setEulaAccepted", eulaAccepted);
      },
    },
  },

  created() {
    this.load(`${this.publicPath}eula.html`);
  },

  methods: {
    close() {
      this.$store.commit("setEulaVisible", false);
    },

    load(url: string) {
      window
        .fetch(url)
        .then((result) => {
          result
            .text()
            .then((result: string) => {
              this.eula = result;
            })
            .catch(() => {
              this.eula =
                'See the license <a href="https://www.osram.com/phytofy">here</a>';
            });
        })
        .catch(() => {
          this.eula =
            'See the license <a href="https://www.osram.com/phytofy">here</a>';
        });
    },
  },
});
</script>
